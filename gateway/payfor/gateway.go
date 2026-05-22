package payfor

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/crypt"
	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "payfor"

// Gateway PayFor (QNB Finansbank, Enpara, Ziraat Katılım PayFor) implementasyonu.
type Gateway struct {
	cfg  Config
	http *httpClient
}

// New yapılandırılmış gateway döner.
func New(cfg Config) gateway.Gateway {
	if cfg.Endpoints.PaymentAPI == "" {
		cfg.Endpoints = DefaultFinansTestEndpoints
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
	sec, ok := modelToSecure[req.PaymentModel]
	if !ok {
		return model.FormData{}, fmt.Errorf("payfor: desteklenmeyen 3D modeli: %s", req.PaymentModel)
	}
	tx, ok := txTypeToPayfor[req.TxType]
	if !ok || (req.TxType != model.TxTypePayAuth && req.TxType != model.TxTypePayPreAuth) {
		return model.FormData{}, fmt.Errorf("payfor: desteklenmeyen işlem türü: %s", req.TxType)
	}

	rnd, err := crypt.RandomString(24)
	if err != nil {
		return model.FormData{}, err
	}

	c := g.cfg.Credentials
	mbrID := c.MbrID
	if mbrID == "" {
		mbrID = "5"
	}
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}

	inputs := map[string]string{
		"MbrId":            mbrID,
		"MerchantID":       c.ClientID,
		"UserCode":         c.Username,
		"OrderId":          req.Order.ID,
		"Lang":             mapLang(c.Lang),
		"SecureType":       sec,
		"TxnType":          tx,
		"PurchAmount":      formatAmount(req.Order.Amount),
		"InstallmentCount": mapInstallment(req.Order.Installment),
		"Currency":         mapCurrency(currency),
		"OkUrl":            req.Order.SuccessURL,
		"FailUrl":          req.Order.FailURL,
		"Rnd":              rnd,
	}
	if req.Card != nil {
		inputs["CardHolderName"] = req.Card.HolderName
		inputs["Pan"] = req.Card.Number
		inputs["Expiry"] = cardExpiryMY(req.Card)
		inputs["Cvv2"] = req.Card.CVV
	}
	inputs["Hash"] = Create3DHash(c.StoreKey, inputs)

	gatewayURL := g.cfg.Endpoints.Gateway3D
	if req.PaymentModel == model.PaymentModel3DHost {
		gatewayURL = g.cfg.Endpoints.Gateway3DHost
		if gatewayURL == "" {
			gatewayURL = DefaultFinansTestEndpoints.Gateway3DHost
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
		return model.PaymentResult{}, errors.New("payfor: boş callback payload")
	}

	order := model.Order{
		ID:     payloadVal(payload, "OrderId"),
		Amount: parseAmount(payloadVal(payload, "PurchAmount"), 0),
		Currency: parseCurrency(payloadVal(payload, "Currency")),
	}

	modelName := paymentModelFromPayload(payload)
	if is3DPayOrHost(modelName) {
		return map3DPayHostResult(payload, order), nil
	}

	if !is3DAuthSuccess(payloadVal(payload, "3DStatus")) {
		return map3DFailed(payload), nil
	}

	if !g.cfg.SkipHashCheck && !Check3DHash(credsFromModel(g.cfg.Credentials), payload) {
		return model.PaymentResult{
			Success:      false,
			OrderID:      order.ID,
			ErrorCode:    "HASH_MISMATCH",
			ErrorMessage: "3D hash doğrulaması başarısız",
			RawResponse:  payloadToRaw(payload),
		}, nil
	}

	return g.complete3DSecure(ctx, payload, order)
}

func (g *Gateway) complete3DSecure(ctx context.Context, payload map[string]string, order model.Order) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	mbrID := c.MbrID
	if mbrID == "" {
		mbrID = "5"
	}
	req := payforRequest{
		MbrId:       mbrID,
		MerchantId:  c.ClientID,
		UserCode:    c.Username,
		UserPass:    c.Password,
		OrderId:     firstNonEmpty(order.ID, payloadVal(payload, "OrderId")),
		SecureType:  "3DModelPayment",
		RequestGuid: payloadVal(payload, "RequestGuid"),
	}
	body, err := encodePayfor(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodePayfor(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	for k, v := range payloadToRaw(payload) {
		raw["callback_"+k] = v
	}
	result := mapAPIResult(resp, raw, order)
	if result.Success {
		result.AuthCode = firstNonEmpty(result.AuthCode, payloadVal(payload, "AuthCode"))
		result.HostRefNum = firstNonEmpty(result.HostRefNum, payloadVal(payload, "HostRefNum"))
	}
	return result, nil
}

// Cancel ödeme iptali.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	return g.apiOp(ctx, req.Order, model.TxTypeCancel, "")
}

// Refund ödeme iadesi.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	amount := ""
	if req.Partial && req.Order.Amount > 0 {
		amount = formatAmount(req.Order.Amount)
	}
	return g.apiOp(ctx, req.Order, model.TxTypeRefund, amount)
}

// Status sipariş sorgulama.
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	mbrID := c.MbrID
	if mbrID == "" {
		mbrID = "5"
	}
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	body, err := encodePayfor(payforRequest{
		MbrId:      mbrID,
		MerchantId: c.ClientID,
		UserCode:   c.Username,
		UserPass:   c.Password,
		OrgOrderId: req.Order.ID,
		SecureType: "Inquiry",
		TxnType:    txTypeToPayfor[model.TxTypeStatus],
		Lang:       mapLang(c.Lang),
		Currency:   mapCurrency(currency),
	})
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodePayfor(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapStatusResult(resp, raw, req.Order), nil
}

func (g *Gateway) apiOp(ctx context.Context, order model.Order, txKey, amount string) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	mbrID := c.MbrID
	if mbrID == "" {
		mbrID = "5"
	}
	currency := order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	req := payforRequest{
		MbrId:      mbrID,
		MerchantId: c.ClientID,
		UserCode:   c.Username,
		UserPass:   c.Password,
		OrgOrderId: order.ID,
		SecureType: "NonSecure",
		TxnType:    txTypeToPayfor[txKey],
		Currency:   mapCurrency(currency),
		Lang:       mapLang(c.Lang),
	}
	if amount != "" {
		req.PurchAmount = amount
	}
	body, err := encodePayfor(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodePayfor(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapAPIResult(resp, raw, order), nil
}

