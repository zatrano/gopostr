package akbank

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "akbank"

// Gateway Akbank yeni sanal POS altyapısı (JSON API + 3D Secure/Pay/Host).
type Gateway struct {
	cfg  Config
	http *httpClient
}

// New yapılandırılmış gateway döner.
func New(cfg Config) gateway.Gateway {
	if cfg.Endpoints.PaymentAPI == "" {
		cfg.Endpoints = DefaultTestEndpoints
	}
	return &Gateway{cfg: cfg, http: newHTTPClient()}
}

// Name gateway adını döner.
func (g *Gateway) Name() string { return gatewayName }

// Init 3D ödeme formu hazırlar.
func (g *Gateway) Init(ctx context.Context, req model.InitRequest) (model.FormData, error) {
	_ = ctx
	if err := validateInit(req); err != nil {
		return model.FormData{}, err
	}
	paymentModel, ok := modelToPayment[req.PaymentModel]
	if !ok {
		return model.FormData{}, fmt.Errorf("akbank: desteklenmeyen 3D modeli: %s", req.PaymentModel)
	}
	txnCodes, ok := txCodeByModel[req.TxType]
	if !ok {
		return model.FormData{}, fmt.Errorf("akbank: desteklenmeyen işlem türü: %s", req.TxType)
	}
	txnCode, ok := txnCodes[req.PaymentModel]
	if !ok {
		return model.FormData{}, fmt.Errorf("akbank: işlem/model uyumsuz: %s/%s", req.TxType, req.PaymentModel)
	}

	rnd, err := RandomNumber()
	if err != nil {
		return model.FormData{}, err
	}

	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}

	inputs := map[string]string{
		"paymentModel":    paymentModel,
		"txnCode":         txnCode,
		"merchantSafeId":  c.ClientID,
		"terminalSafeId":  c.Username,
		"orderId":         req.Order.ID,
		"lang":            mapLang(c.Lang),
		"amount":          formatAmount(req.Order.Amount),
		"currencyCode":    fmt.Sprintf("%d", mapCurrency(currency)),
		"installCount":    fmt.Sprintf("%d", mapInstallment(req.Order.Installment)),
		"okUrl":           req.Order.SuccessURL,
		"failUrl":         req.Order.FailURL,
		"randomNumber":    rnd,
		"requestDateTime": requestDateTime(),
	}
	if c.SubMerchantID != "" {
		inputs["subMerchantId"] = c.SubMerchantID
	}
	if req.Card != nil {
		inputs["creditCard"] = req.Card.Number
		inputs["expiredDate"] = req.Card.ExpireMonth + req.Card.ExpireYear
		inputs["cvv"] = req.Card.CVV
	}
	inputs["hash"] = Create3DHash(c.StoreKey, inputs)

	gatewayURL := g.cfg.Endpoints.Gateway3D
	if req.PaymentModel == model.PaymentModel3DHost || req.PaymentModel == model.PaymentModel3DPayHosting {
		gatewayURL = g.cfg.Endpoints.Gateway3DHost
		if gatewayURL == "" {
			gatewayURL = DefaultTestEndpoints.Gateway3DHost
		}
	}

	return model.FormData{
		Gateway: gatewayURL,
		Method:  http.MethodPost,
		Inputs:  inputs,
	}, nil
}

// HandleCallback banka 3D dönüşünü işler.
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("akbank: boş callback payload")
	}

	if !g.cfg.SkipHashCheck && !Check3DHash(g.cfg.Credentials.StoreKey, payload) {
		return model.PaymentResult{
			Success:      false,
			OrderID:      payloadVal(payload, "orderId"),
			ErrorCode:    "HASH_MISMATCH",
			ErrorMessage: "3D hash doğrulaması başarısız",
			RawResponse:  payloadToRaw(payload),
		}, nil
	}

	order := orderFromPayload(payload)
	if !is3DAuthSuccess(payloadVal(payload, "responseCode")) {
		return map3DFailed(payload), nil
	}

	modelName := paymentModelFromPayload(payload)
	if is3DPayOrHost(modelName) {
		return map3DPayHostResult(payload, order), nil
	}

	return g.complete3DSecure(ctx, payload, order)
}

func (g *Gateway) complete3DSecure(ctx context.Context, payload map[string]string, order model.Order) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	rnd, err := RandomNumber()
	if err != nil {
		return model.PaymentResult{}, err
	}

	txType := model.TxTypePayAuth
	if payloadVal(payload, "txnCode") == "3004" {
		txType = model.TxTypePayPreAuth
	}

	amount := payloadVal(payload, "amount")
	if amount == "" {
		amount = formatAmount(order.Amount)
	}
	currencyCode := mapCurrency(order.Currency)
	if cc := payloadVal(payload, "currencyCode"); cc != "" {
		currencyCode = mapCurrency(parseCurrency(cc))
	}

	req := apiRequest{
		Version:         apiVersion,
		TxnCode:         txnCodeProvision(txType),
		RequestDateTime: requestDateTime(),
		RandomNumber:    rnd,
		Terminal: terminalBlock{
			MerchantSafeID: c.ClientID,
			TerminalSafeID: c.Username,
		},
		Order: orderBlock{OrderID: firstNonEmpty(payloadVal(payload, "orderId"), order.ID)},
		Transaction: &transactionBlock{
			Amount:       amount,
			CurrencyCode: currencyCode,
			MotoInd:      0,
			InstallCount: mapInstallment(order.Installment),
		},
		SecureTransaction: &secureTxBlock{
			SecureID:      payloadVal(payload, "secureId"),
			SecureEcomInd: payloadVal(payload, "secureEcomInd"),
			SecureData:    payloadVal(payload, "secureData"),
			SecureMd:      payloadVal(payload, "secureMd"),
		},
		Customer: &customerBlock{
			IPAddress: firstNonEmpty(order.IP, "127.0.0.1"),
		},
	}
	if c.SubMerchantID != "" {
		req.SubMerchant = &subMerchantBlock{SubMerchantID: c.SubMerchantID}
	}

	body, err := encodeAPI(req)
	if err != nil {
		return model.PaymentResult{}, err
	}

	url := g.cfg.Endpoints.PaymentAPI + "/transaction/process"
	respBody, err := g.http.postJSON(ctx, url, c.StoreKey, body)
	if err != nil {
		return model.PaymentResult{}, err
	}

	resp, raw, err := decodeAPI(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	for k, v := range payloadToRaw(payload) {
		if _, ok := raw[k]; !ok {
			raw["callback_"+k] = v
		}
	}
	return mapAPIResult(resp, raw, order), nil
}

// Cancel ödeme iptali.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	return g.apiOp(ctx, req.Order, txCodeCancel, false)
}

// Refund ödeme iadesi.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	order := req.Order
	if !req.Partial || order.Amount <= 0 {
		order.Amount = 0
	}
	return g.apiOp(ctx, order, txCodeRefund, order.Amount > 0)
}

// Status Akbank API'de desteklenmiyor.
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	_ = ctx
	_ = req
	return model.PaymentResult{}, errors.New("akbank: durum sorgulama desteklenmiyor")
}

func (g *Gateway) apiOp(ctx context.Context, order model.Order, txnCode string, withAmount bool) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	rnd, err := RandomNumber()
	if err != nil {
		return model.PaymentResult{}, err
	}
	currency := order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}

	req := apiRequest{
		Version:         apiVersion,
		TxnCode:         txnCode,
		RequestDateTime: requestDateTime(),
		RandomNumber:    rnd,
		Terminal: terminalBlock{
			MerchantSafeID: c.ClientID,
			TerminalSafeID: c.Username,
		},
		Order: orderBlock{OrderID: order.ID},
	}
	if c.SubMerchantID != "" {
		req.SubMerchant = &subMerchantBlock{SubMerchantID: c.SubMerchantID}
	}
	if withAmount {
		req.Transaction = &transactionBlock{
			Amount:       formatAmount(order.Amount),
			CurrencyCode: mapCurrency(currency),
		}
	}

	body, err := encodeAPI(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	url := g.cfg.Endpoints.PaymentAPI + "/transaction/process"
	respBody, err := g.http.postJSON(ctx, url, c.StoreKey, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeAPI(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapAPIResult(resp, raw, order), nil
}

func validateInit(req model.InitRequest) error {
	if req.Order.ID == "" {
		return errors.New("akbank: sipariş ID zorunlu")
	}
	if req.Order.Amount <= 0 {
		return errors.New("akbank: geçersiz tutar")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errors.New("akbank: success/fail URL zorunlu")
	}
	if req.Order.IP == "" {
		return errors.New("akbank: müşteri IP zorunlu")
	}
	c := req.Order
	_ = c
	return nil
}
