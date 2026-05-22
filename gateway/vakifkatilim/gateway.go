package vakifkatilim

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "vakifkatilim"

const (
	txSecurity3D    = "3"
	txSecurity3DHost = "5"
)

// Gateway Vakıf Katılım sanal POS (3D Secure, 3D Host).
type Gateway struct {
	cfg  Config
	http *httpClient
}

// New yapılandırılmış gateway döner.
func New(cfg Config) gateway.Gateway {
	if cfg.Endpoints.PaymentAPI == "" && cfg.Endpoints.Gateway3D == "" && cfg.Endpoints.Gateway3DHost == "" {
		cfg.Endpoints = DefaultTestEndpoints
	}
	return &Gateway{cfg: cfg, http: newHTTPClient()}
}

// Name gateway adını döner.
func (g *Gateway) Name() string { return gatewayName }

// Init 3D Secure enrollment veya 3D Host ödeme formu hazırlar.
func (g *Gateway) Init(ctx context.Context, req model.InitRequest) (model.FormData, error) {
	if err := validateInit(req); err != nil {
		return model.FormData{}, err
	}
	if req.TxType != model.TxTypePayAuth {
		return model.FormData{}, fmt.Errorf("vakifkatilim: yalnızca satış (pay_auth) desteklenir")
	}

	switch req.PaymentModel {
	case model.PaymentModel3DHost:
		return g.init3DHost(req), nil
	case model.PaymentModel3DSecure:
		return g.init3DSecure(ctx, req)
	default:
		return model.FormData{}, fmt.Errorf("vakifkatilim: desteklenmeyen 3D modeli: %s", req.PaymentModel)
	}
}

func (g *Gateway) init3DHost(req model.InitRequest) model.FormData {
	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	return model.FormData{
		Gateway: g.cfg.Endpoints.Gateway3DHost,
		Method:  http.MethodPost,
		Inputs: map[string]string{
			"UserName":        c.Username,
			"HashPassword":    HashPassword(c.StoreKey),
			"MerchantId":      c.ClientID,
			"MerchantOrderId": req.Order.ID,
			"Amount":          fmt.Sprintf("%d", formatAmountKurus(req.Order.Amount)),
			"FECCurrencyCode": mapCurrency(currency),
			"OkUrl":           req.Order.SuccessURL,
			"FailUrl":         req.Order.FailURL,
			"PaymentType":     "1",
		},
	}
}

func (g *Gateway) init3DSecure(ctx context.Context, req model.InitRequest) (model.FormData, error) {
	if req.Card == nil {
		return model.FormData{}, errors.New("vakifkatilim: 3D Secure için kart zorunlu")
	}
	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	amount := formatAmountKurus(req.Order.Amount)
	inst, _ := strconv.Atoi(mapInstallment(req.Order.Installment))
	m, y := cardExpiryParts(req.Card)

	hashIn := hashInput(c, map[string]string{
		"MerchantOrderId": req.Order.ID,
		"Amount":          fmt.Sprintf("%d", amount),
		"OkUrl":           req.Order.SuccessURL,
		"FailUrl":         req.Order.FailURL,
	})
	hash := CreateHash(c.StoreKey, hashIn)
	acc := accountFields(c)

	xmlBody := encodeEnrollment(enrollmentFields{
		MerchantID:      acc["MerchantId"],
		CustomerID:      acc["CustomerId"],
		UserName:        acc["UserName"],
		SubMerchantID:   acc["SubMerchantId"],
		HashPassword:    HashPassword(c.StoreKey),
		HashData:        hash,
		TxSecurity:      txSecurity3D,
		Installment:     inst,
		Amount:          amount,
		DisplayAmount:   amount,
		FECCurrency:     mapCurrency(currency),
		MerchantOrderID: req.Order.ID,
		OkURL:           req.Order.SuccessURL,
		FailURL:         req.Order.FailURL,
		CardHolder:      req.Card.HolderName,
		CardNumber:      req.Card.Number,
		CardMonth:       m,
		CardYear:        y,
		CVV:             req.Card.CVV,
	})

	body, err := g.http.postXML(ctx, g.cfg.Endpoints.Gateway3D, xmlBody)
	if err != nil {
		return model.FormData{}, err
	}
	if !isHTMLResponse(body) {
		return model.FormData{}, fmt.Errorf("vakifkatilim: enrollment HTML yanıt bekleniyordu")
	}
	gw, inputs, err := parseHTMLForm(string(body))
	if err != nil {
		return model.FormData{}, err
	}
	return model.FormData{Gateway: gw, Method: http.MethodPost, Inputs: inputs}, nil
}

// HandleCallback banka 3D dönüşünü işler.
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("vakifkatilim: boş callback payload")
	}

	order := orderFromPayload(payload)
	if order.ID == "" {
		order.ID = payloadVal(payload, "MerchantOrderId")
	}

	// 3D Host: provizyon API yok, callback doğrudan sonuç
	if payloadVal(payload, "MD") == "" && payloadVal(payload, "PaymentType") != "" {
		if isAuthSuccess(payloadVal(payload, "ResponseCode")) {
			return mapPaymentResult(payload, order), nil
		}
		return mapAuthFailed(payload, order), nil
	}

	if !isAuthSuccess(payloadVal(payload, "ResponseCode")) {
		return mapAuthFailed(payload, order), nil
	}

	md := payloadVal(payload, "MD")
	if md == "" {
		return model.PaymentResult{}, errors.New("vakifkatilim: MD eksik")
	}

	c := g.cfg.Credentials
	acc := accountFields(c)
	amount := formatAmountKurus(order.Amount)
	if v := payloadVal(payload, "Amount"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			amount = n
		}
	}
	inst, _ := strconv.Atoi(mapInstallment(order.Installment))
	merchantOrder := firstNonEmpty(payloadVal(payload, "MerchantOrderId"), order.ID)

	hashIn := hashInput(c, map[string]string{
		"MerchantOrderId": merchantOrder,
		"Amount":          fmt.Sprintf("%d", amount),
		"OkUrl":           payloadVal(payload, "OkUrl"),
		"FailUrl":         payloadVal(payload, "FailUrl"),
	})
	if hashIn["OkUrl"] == "" {
		hashIn["OkUrl"] = ""
		hashIn["FailUrl"] = ""
	}

	provXML := encodeProvision(provisionFields{
		MerchantID:      acc["MerchantId"],
		CustomerID:      acc["CustomerId"],
		UserName:        acc["UserName"],
		SubMerchantID:   acc["SubMerchantId"],
		OkURL:           payloadVal(payload, "OkUrl"),
		FailURL:         payloadVal(payload, "FailUrl"),
		HashData:        CreateHash(c.StoreKey, hashIn),
		MD:              md,
		Installment:     inst,
		Amount:          amount,
		MerchantOrderID: merchantOrder,
		TxSecurity:      txSecurity3D,
	})

	provURL := strings.TrimRight(g.cfg.Endpoints.PaymentAPI, "/") + "/ThreeDModelProvisionGate"
	provBody, err := g.http.postXML(ctx, provURL, provXML)
	if err != nil {
		return model.PaymentResult{}, err
	}
	payRaw, err := decodeXML(provBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return map3DResult(payload, payRaw, order), nil
}

// Cancel ödeme iptali.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	return g.postOp(ctx, "SaleReversal", req.Order, true)
}

// Refund iade.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	path := "DrawBack"
	if req.Partial {
		path = "PartialDrawBack"
	}
	return g.postOp(ctx, path, req.Order, true)
}

// Status sipariş durumu.
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	acc := accountFields(c)
	fields := acc
	fields["MerchantOrderId"] = req.Order.ID
	fields["HashData"] = CreateHash(c.StoreKey, hashInput(c, map[string]string{
		"MerchantOrderId": req.Order.ID,
		"Amount":          "0",
	}))
	body, err := g.http.postXML(ctx, apiURL(g.cfg.Endpoints.PaymentAPI, "SelectOrderByMerchantOrderId"), encodeStatus(fields))
	if err != nil {
		return model.PaymentResult{}, err
	}
	raw, err := decodeXML(body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapPaymentResult(raw, req.Order), nil
}

func (g *Gateway) postOp(ctx context.Context, path string, order model.Order, includeAmount bool) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	acc := accountFields(c)
	amount := formatAmountKurus(order.Amount)
	fields := map[string]string{
		"MerchantId":      acc["MerchantId"],
		"CustomerId":      acc["CustomerId"],
		"UserName":        acc["UserName"],
		"SubMerchantId":   acc["SubMerchantId"],
		"HashPassword":    HashPassword(c.StoreKey),
		"MerchantOrderId": order.ID,
		"OrderId":         bankOrderID(order),
		"PaymentType":     "1",
	}
	if includeAmount {
		fields["Amount"] = fmt.Sprintf("%d", amount)
	}
	fields["HashData"] = CreateHash(c.StoreKey, hashInput(c, map[string]string{
		"MerchantOrderId": order.ID,
		"Amount":          fields["Amount"],
	}))

	body, err := g.http.postXML(ctx, apiURL(g.cfg.Endpoints.PaymentAPI, path), encodeCancelRefund(fields))
	if err != nil {
		return model.PaymentResult{}, err
	}
	raw, err := decodeXML(body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapPaymentResult(raw, order), nil
}

func apiURL(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + path
}

func validateInit(req model.InitRequest) error {
	if req.Order.ID == "" {
		return errors.New("vakifkatilim: sipariş ID gerekli")
	}
	if req.Order.Amount <= 0 {
		return errors.New("vakifkatilim: tutar gerekli")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errors.New("vakifkatilim: success/fail URL gerekli")
	}
	return nil
}
