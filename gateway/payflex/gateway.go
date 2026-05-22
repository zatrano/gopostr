package payflex

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "payflex"

// Gateway PayFlex VPOS 7/24 (Vakıfbank, Ziraat) — yalnızca 3D Secure.
type Gateway struct {
	cfg  Config
	http *httpClient
}

// New yapılandırılmış gateway döner.
func New(cfg Config) gateway.Gateway {
	if cfg.Endpoints.PaymentAPI == "" {
		cfg.Endpoints = DefaultVakifTestEndpoints
	}
	return &Gateway{cfg: cfg, http: newHTTPClient()}
}

// Name gateway adını döner.
func (g *Gateway) Name() string { return gatewayName }

// Init MPI enrollment sonrası ACS formunu hazırlar (kart zorunlu).
func (g *Gateway) Init(ctx context.Context, req model.InitRequest) (model.FormData, error) {
	if err := validateInit(req); err != nil {
		return model.FormData{}, err
	}
	if req.PaymentModel != model.PaymentModel3DSecure {
		return model.FormData{}, fmt.Errorf("payflex: yalnızca 3D Secure desteklenir")
	}

	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}

	verifyID, err := newEnrollmentRequestID()
	if err != nil {
		return model.FormData{}, err
	}

	fields := map[string]string{
		"MerchantId":                c.ClientID,
		"MerchantPassword":          c.Password,
		"MerchantType":              fmt.Sprintf("%d", c.MerchantType),
		"PurchaseAmount":            formatAmount(req.Order.Amount),
		"VerifyEnrollmentRequestId": verifyID,
		"Currency":                  mapCurrency(currency),
		"SuccessUrl":                req.Order.SuccessURL,
		"FailureUrl":                req.Order.FailURL,
		"Pan":                       req.Card.Number,
		"ExpiryDate":                cardExpiryYM(req.Card),
		"IsRecurring":               "false",
	}
	if brand := mapBrand(req.Card.Brand); brand != "" {
		fields["BrandName"] = brand
	}
	if req.Order.Installment > 1 {
		fields["InstallmentCount"] = mapInstallment(req.Order.Installment)
	}
	if c.SubMerchantID != "" && c.MerchantType == 2 {
		fields["SubMerchantId"] = c.SubMerchantID
	}

	body, err := g.http.postForm(ctx, g.cfg.Endpoints.Gateway3D, fields)
	if err != nil {
		return model.FormData{}, err
	}

	veres, errMsg, err := decodeEnrollment(body)
	if err != nil {
		return model.FormData{}, err
	}
	if errMsg != "" {
		return model.FormData{}, errors.New(errMsg)
	}

	switch veres.Status {
	case "Y":
		// devam
	case "N":
		return model.FormData{}, errors.New("payflex: kart 3-D Secure programına dahil değil")
	case "U":
		return model.FormData{}, errors.New("payflex: işlem gerçekleştirilemiyor")
	default:
		return model.FormData{}, fmt.Errorf("payflex: enrollment durumu: %s", veres.Status)
	}

	return model.FormData{
		Gateway: veres.ACSUrl,
		Method:  http.MethodPost,
		Inputs: map[string]string{
			"PaReq":   veres.PaReq,
			"TermUrl": veres.TermUrl,
			"MD":      veres.MD,
		},
	}, nil
}

// HandleCallback 3D dönüşü ve provizyon.
// Provizyon için payload'a kart bilgisi eklenmelidir: pan, expiry, cvv, cardHoldersName (oturumdan).
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("payflex: boş callback payload")
	}

	status := payloadVal(payload, "Status")
	if !is3DAuthSuccess(status) {
		return map3DFailed(payload), nil
	}

	card, expiry, err := cardFromPayload(payload)
	if err != nil {
		return model.PaymentResult{}, err
	}

	order := model.Order{
		ID:     firstNonEmpty(payloadVal(payload, "OrderId"), payloadVal(payload, "VerifyEnrollmentRequestId")),
		Amount: parseAmount(payloadVal(payload, "PurchAmount")),
		IP:     payloadVal(payload, "ClientIp"),
	}
	if order.Amount == 0 {
		order.Amount = parseAmount(payloadVal(payload, "PurchaseAmount"))
	}

	return g.complete3DSecure(ctx, payload, order, card, expiry)
}

func (g *Gateway) complete3DSecure(ctx context.Context, payload map[string]string, order model.Order, card model.CardInput, expiry string) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	txType := model.TxTypePayAuth
	if payloadVal(payload, "TxnType") == "Auth" {
		txType = model.TxTypePayPreAuth
	}

	currency := order.Currency
	if currency == "" {
		currency = parseCurrency(payloadVal(payload, "PurchCurrency"))
	}

	req := vposRequest{
		MerchantId:              c.ClientID,
		Password:                c.Password,
		TerminalNo:              c.TerminalID,
		TransactionType:         txTypeToPayflex[txType],
		TransactionId:           order.ID,
		CurrencyAmount:          formatAmount(order.Amount),
		CurrencyCode:            mapCurrency(currency),
		ECI:                     payloadVal(payload, "Eci"),
		CAVV:                    payloadVal(payload, "Cavv"),
		MpiTransactionId:        payloadVal(payload, "VerifyEnrollmentRequestId"),
		OrderId:                 order.ID,
		ClientIp:                firstNonEmpty(order.IP, payloadVal(payload, "ClientIp"), "127.0.0.1"),
		TransactionDeviceSource: "0",
		CardHoldersName:         card.HolderName,
		Cvv:                     card.CVV,
		Pan:                     card.Number,
		Expiry:                  expiry,
	}
	if order.Installment > 1 {
		req.NumberOfInstallments = mapInstallment(order.Installment)
	}

	body, err := encodeVpos(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeVpos(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	for k, v := range payloadToRaw(payload) {
		raw["callback_"+k] = v
	}
	return mapVposResult(resp, raw, order), nil
}

// Cancel ödeme iptali.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	return g.simpleVpos(ctx, req.Order, txTypeToPayflex[model.TxTypeCancel], req.Order.TransactionID, "")
}

// Refund ödeme iadesi.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	amount := ""
	if req.Partial && req.Order.Amount > 0 {
		amount = formatAmount(req.Order.Amount)
	}
	return g.simpleVpos(ctx, req.Order, txTypeToPayflex[model.TxTypeRefund], req.Order.TransactionID, amount)
}

// Status işlem sorgulama (Search API).
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	search := searchRequest{
		MerchantCriteria: merchantCriteria{
			HostMerchantId:   c.ClientID,
			MerchantPassword: c.Password,
		},
		TransactionCriteria: transactionCriteria{
			TransactionId: req.Order.TransactionID,
			OrderId:       req.Order.ID,
		},
	}
	body, err := encodeSearch(search)
	if err != nil {
		return model.PaymentResult{}, err
	}
	queryURL := g.cfg.Endpoints.QueryAPI
	if queryURL == "" {
		queryURL = g.cfg.Endpoints.PaymentAPI
	}
	respBody, err := g.http.postSearch(ctx, queryURL, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeSearch(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapSearchResult(resp, raw, req.Order), nil
}

func (g *Gateway) simpleVpos(ctx context.Context, order model.Order, txType, refTxnID, amount string) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	req := vposRequest{
		MerchantId:             c.ClientID,
		Password:               c.Password,
		TransactionType:        txType,
		ReferenceTransactionId: refTxnID,
		ClientIp:               firstNonEmpty(order.IP, "127.0.0.1"),
	}
	if amount != "" {
		req.CurrencyAmount = amount
	}
	body, err := encodeVpos(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeVpos(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapVposResult(resp, raw, order), nil
}

