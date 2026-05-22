package garanti

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "garanti"

// Gateway Garanti BBVA sanal POS implementasyonu (3D Secure, 3D Pay).
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
	secLevel, ok := modelToSecure[req.PaymentModel]
	if !ok {
		return model.FormData{}, fmt.Errorf("garanti: desteklenmeyen 3D modeli: %s", req.PaymentModel)
	}
	tx, ok := txTypeToGaranti[req.TxType]
	if !ok || (req.TxType != model.TxTypePayAuth && req.TxType != model.TxTypePayPreAuth) {
		return model.FormData{}, fmt.Errorf("garanti: desteklenmeyen işlem türü: %s", req.TxType)
	}

	c := g.cfg.Credentials
	inputs := map[string]string{
		"secure3dsecuritylevel": secLevel,
		"mode":                  g.mode(),
		"apiversion":            apiVersion,
		"terminalprovuserid":    c.Username,
		"terminaluserid":        c.Username,
		"terminalmerchantid":    c.ClientID,
		"terminalid":            c.TerminalID,
		"txntype":               tx,
		"txnamount":             formatAmountKurus(req.Order.Amount),
		"txncurrencycode":       mapCurrency(req.Order.Currency),
		"txninstallmentcount":   mapInstallment(req.Order.Installment),
		"orderid":               req.Order.ID,
		"successurl":            req.Order.SuccessURL,
		"errorurl":              req.Order.FailURL,
		"customeripaddress":     req.Order.IP,
	}
	if req.Card != nil {
		inputs["cardnumber"] = req.Card.Number
		inputs["cardexpiredatemonth"] = req.Card.ExpireMonth
		inputs["cardexpiredateyear"] = req.Card.ExpireYear
		inputs["cardcvv2"] = req.Card.CVV
	}
	inputs["secure3dhash"] = Create3DHash(credsFromModel(c), inputs)

	return model.FormData{
		Gateway: g.cfg.Endpoints.Gateway3D,
		Method:  http.MethodPost,
		Inputs:  inputs,
	}, nil
}

// HandleCallback banka 3D dönüşünü işler.
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("garanti: boş callback payload")
	}

	if !g.cfg.SkipHashCheck && !Check3DHash(g.cfg.Credentials.StoreKey, payload) {
		return model.PaymentResult{
			Success:      false,
			OrderID:      orderFromPayload(payload).ID,
			ErrorCode:    "HASH_MISMATCH",
			ErrorMessage: "3D hash doğrulaması başarısız",
			RawResponse:  payloadToRaw(payload),
		}, nil
	}

	order := orderFromPayload(payload)
	modelName := paymentModelFromPayload(payload)

	if !is3DAuthSuccess(payloadVal(payload, "mdstatus")) {
		r := map3DCommon(payload)
		r.Success = false
		return r, nil
	}

	switch modelName {
	case model.PaymentModel3DPay:
		return map3DPayResult(payload), nil
	default:
		return g.complete3DSecure(ctx, payload, order)
	}
}

func (g *Gateway) complete3DSecure(ctx context.Context, payload map[string]string, order model.Order) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	txType := payloadVal(payload, "txntype")
	amount := payloadVal(payload, "txnamount")
	currency := payloadVal(payload, "txncurrencycode")

	req := gvpsRequest{
		Mode:    g.mode(),
		Version: apiVersion,
		Terminal: gvpsTerminal{
			ProvUserID: c.Username,
			UserID:     c.Username,
			ID:         c.TerminalID,
			MerchantID: c.ClientID,
		},
		Customer: &gvpsCustomer{IPAddress: firstNonEmpty(payloadVal(payload, "customeripaddress"), order.IP, "127.0.0.1")},
		Order:    gvpsOrder{OrderID: firstNonEmpty(payloadVal(payload, "orderid"), order.ID)},
		Transaction: gvpsTransaction{
			Type:                  txType,
			InstallmentCnt:        payloadVal(payload, "txninstallmentcount"),
			Amount:                amount,
			CurrencyCode:          currency,
			CardholderPresentCode: "13",
			MotoInd:               "N",
			Secure3D: &gvpsSecure3D{
				AuthenticationCode: payloadVal(payload, "cavv"),
				SecurityLevel:      payloadVal(payload, "eci"),
				TxnID:              payloadVal(payload, "xid"),
				Md:                 payloadVal(payload, "md"),
			},
		},
	}
	mc := credsFromModel(c)
	req.Terminal.HashData = CreateAPIHash(mc,
		req.Order.OrderID, c.TerminalID, "",
		amount, currency, txType,
	)

	body, err := encodeGVPS(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.post(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeGVPS(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	for k, v := range payloadToRaw(payload) {
		if _, ok := raw[k]; !ok {
			raw[k] = v
		}
	}
	result := mapAPIResult(resp, raw, order)
	if !result.Success {
		return result, nil
	}
	// 3D auth alanlarını ekle
	result.TransactionID = payloadVal(payload, "transid")
	if result.AuthCode == "" {
		result.AuthCode = payloadVal(payload, "authcode")
	}
	if result.HostRefNum == "" {
		result.HostRefNum = payloadVal(payload, "hostrefnum")
	}
	return result, nil
}

// Cancel ödeme iptali.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	return g.nonSecureOp(ctx, req.Order, model.TxTypeCancel, true)
}

// Refund ödeme iadesi.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	tx := model.TxTypeRefund
	if req.Partial {
		tx = model.TxTypeRefundPartial
	}
	return g.nonSecureOp(ctx, req.Order, tx, true)
}

// Status sipariş sorgulama.
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	order := req.Order
	if order.Amount <= 0 {
		order.Amount = 1
	}
	return g.nonSecureOp(ctx, order, model.TxTypeStatus, false)
}

func (g *Gateway) nonSecureOp(ctx context.Context, order model.Order, txKey string, isRefund bool) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	tx := txTypeToGaranti[txKey]
	amount := formatAmountKurus(order.Amount)
	if txKey == model.TxTypeStatus || txKey == model.TxTypeCancel {
		if order.Amount <= 0 {
			amount = formatAmountKurus(1)
		}
	}
	currency := mapCurrency(order.Currency)
	if currency == "" {
		currency = mapCurrency(model.CurrencyTRY)
	}

	provUser := c.Username
	if isRefund && c.RefundUsername != "" {
		provUser = c.RefundUsername
	}

	req := gvpsRequest{
		Mode:    g.mode(),
		Version: apiVersion,
		Terminal: gvpsTerminal{
			ProvUserID: provUser,
			UserID:     provUser,
			ID:         c.TerminalID,
			MerchantID: c.ClientID,
		},
		Customer: &gvpsCustomer{IPAddress: firstNonEmpty(order.IP, "127.0.0.1")},
		Order:    gvpsOrder{OrderID: order.ID},
		Transaction: gvpsTransaction{
			Type:                  tx,
			InstallmentCnt:        mapInstallment(order.Installment),
			Amount:                amount,
			CurrencyCode:          currency,
			CardholderPresentCode: "0",
			MotoInd:               "N",
			OriginalRetrefNum:     order.RefRetNum,
		},
	}
	mc := credsFromModel(c)
	req.Terminal.HashData = CreateAPIHash(mc, order.ID, c.TerminalID, "", amount, currency, tx)

	body, err := encodeGVPS(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.post(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeGVPS(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapAPIResult(resp, raw, order), nil
}

func (g *Gateway) mode() string {
	if g.cfg.Credentials.TestMode {
		return "TEST"
	}
	return "PROD"
}

