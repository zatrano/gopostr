package posnetv1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "posnetv1"

// Gateway Albaraka PosNet V1 JSON API — yalnızca 3D Secure.
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

// Init 3D Secure formu hazırlar.
func (g *Gateway) Init(ctx context.Context, req model.InitRequest) (model.FormData, error) {
	_ = ctx
	if err := validateInit(req); err != nil {
		return model.FormData{}, err
	}

	c := g.cfg.Credentials
	if c.ClientID == "" || c.StoreKey == "" {
		return model.FormData{}, errors.New("posnetv1: ClientID ve StoreKey zorunlu")
	}
	if terminalNo(c) == "" || posNetID(c) == "" {
		return model.FormData{}, errors.New("posnetv1: TerminalID ve PosNetID zorunlu")
	}

	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}

	orderID, err := formatOrderID(req.Order.ID)
	if err != nil {
		return model.FormData{}, err
	}

	txAPI, err := txTypeToAPI(req.TxType)
	if err != nil {
		return model.FormData{}, err
	}

	inputs := map[string]string{
		"MerchantNo":        c.ClientID,
		"TerminalNo":        terminalNo(c),
		"PosnetID":          posNetID(c),
		"TransactionType":   txAPI,
		"OrderId":           orderID,
		"Amount":            fmt.Sprintf("%d", formatAmountKurus(req.Order.Amount)),
		"CurrencyCode":      mapCurrency(currency),
		"MerchantReturnURL": req.Order.SuccessURL,
		"InstallmentCount":  mapInstallment(req.Order.Installment),
		"Language":          mapLang(c.Lang),
		"TxnState":          "INITIAL",
		"OpenNewWindow":     "0",
		"CardNo":            req.Card.Number,
		"ExpiredDate":       cardExpiryYM(req.Card),
		"Cvv":               req.Card.CVV,
		"CardHolderName":    req.Card.HolderName,
		"MacParams":         "MerchantNo:TerminalNo:CardNo:Cvc2:ExpireDate:Amount",
		"UseOOS":            "0",
	}
	inputs["Mac"] = Create3DFormHash(c.StoreKey, inputs)

	return model.FormData{
		Gateway: g.cfg.Endpoints.Gateway3D,
		Method:  http.MethodPost,
		Inputs:  inputs,
	}, nil
}

// HandleCallback 3D doğrulama sonrası Sale/Auth provizyonu tamamlar.
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("posnetv1: boş callback payload")
	}

	order := model.Order{
		ID:       payloadVal(payload, "OrderId"),
		Currency: model.CurrencyTRY,
	}
	if v := payloadVal(payload, "Amount"); v != "" {
		order.Amount = parseAmountKurus(v)
	}

	if !is3DAuthSuccess(payloadVal(payload, "MdStatus")) {
		return map3DFailed(payload, order), nil
	}

	c := g.cfg.Credentials
	if !g.cfg.SkipHashCheck && !Check3DCallbackHash(c.StoreKey, payload) {
		return model.PaymentResult{
			Success:      false,
			OrderID:      order.ID,
			ErrorCode:    "HASH_MISMATCH",
			ErrorMessage: "3D Mac doğrulaması başarısız",
			RawResponse:  payloadToRaw(payload),
		}, nil
	}

	return g.complete3DSecure(ctx, payload, order)
}

func (g *Gateway) complete3DSecure(ctx context.Context, payload map[string]string, order model.Order) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	orderID, err := formatOrderID(firstNonEmpty(order.ID, payloadVal(payload, "OrderId")))
	if err != nil {
		return model.PaymentResult{}, err
	}
	if order.Amount == 0 {
		return model.PaymentResult{}, errMissing("tutar (callback Amount veya Order)")
	}
	if order.Currency == "" {
		order.Currency = model.CurrencyTRY
	}

	txType := model.TxTypePayAuth
	if payloadVal(payload, "TranType") == "Auth" {
		txType = model.TxTypePayPreAuth
	}
	txAPI, err := txTypeToAPI(txType)
	if err != nil {
		return model.PaymentResult{}, err
	}

	mf := merchantFields{
		MerchantNo: c.ClientID,
		TerminalNo: terminalNo(c),
		StoreKey:   c.StoreKey,
		OrderID:    orderID,
	}
	reqBody, err := buildProvisionRequest(mf, order, txType, payload)
	if err != nil {
		return model.PaymentResult{}, err
	}
	body, err := encodeJSON(reqBody)
	if err != nil {
		return model.PaymentResult{}, err
	}

	url := joinAPI(g.cfg.Endpoints.PaymentAPI, txAPI)
	respBody, err := g.http.postJSON(ctx, url, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	_, raw, err := decodeAPI(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapProvision(payload, raw, order), nil
}

// Cancel ödeme iptali.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	return g.postCancelRefund(ctx, req.Order, "Reverse")
}

// Refund ödeme iadesi.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	return g.postCancelRefund(ctx, req.Order, "Return")
}

// Status işlem sorgulama.
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	if req.Order.ID == "" {
		return model.PaymentResult{}, errMissing("sipariş id")
	}
	c := g.cfg.Credentials
	mf := merchantFields{
		MerchantNo: c.ClientID,
		TerminalNo: terminalNo(c),
		StoreKey:   c.StoreKey,
	}
	reqBody, err := buildStatusRequest(mf, req.Order.ID)
	if err != nil {
		return model.PaymentResult{}, err
	}
	body, err := encodeJSON(reqBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	url := joinAPI(g.cfg.Endpoints.PaymentAPI, "TransactionInquiry")
	respBody, err := g.http.postJSON(ctx, url, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	_, raw, err := decodeAPI(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapStatus(raw, req.Order), nil
}

func (g *Gateway) postCancelRefund(ctx context.Context, order model.Order, txAPI string) (model.PaymentResult, error) {
	if order.ID == "" && order.RefRetNum == "" {
		return model.PaymentResult{}, errMissing("sipariş id veya RefRetNum")
	}
	c := g.cfg.Credentials
	mf := merchantFields{
		MerchantNo: c.ClientID,
		TerminalNo: terminalNo(c),
		StoreKey:   c.StoreKey,
	}
	reqBody, err := buildCancelRefundRequest(mf, order, txAPI)
	if err != nil {
		return model.PaymentResult{}, err
	}
	body, err := encodeJSON(reqBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	url := joinAPI(g.cfg.Endpoints.PaymentAPI, txAPI)
	respBody, err := g.http.postJSON(ctx, url, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	_, raw, err := decodeAPI(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapCancelRefund(raw), nil
}
