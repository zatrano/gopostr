package interpos

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/crypt"
	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "interpos"

// Gateway Denizbank InterPos implementasyonu (3D Secure, 3D Pay, 3D Host).
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

	secureType, ok := modelToSecureType[req.PaymentModel]
	if !ok {
		return model.FormData{}, fmt.Errorf("interpos: desteklenmeyen 3D modeli: %s", req.PaymentModel)
	}
	tx, ok := txTypeToInter[req.TxType]
	if !ok || (req.TxType != model.TxTypePayAuth && req.TxType != model.TxTypePayPreAuth) {
		return model.FormData{}, fmt.Errorf("interpos: desteklenmeyen işlem türü: %s", req.TxType)
	}

	rnd, err := crypt.RandomString(24)
	if err != nil {
		return model.FormData{}, err
	}

	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}

	inputs := map[string]string{
		"ShopCode":         g.cfg.Credentials.ClientID,
		"TxnType":          tx,
		"SecureType":       secureType,
		"PurchAmount":      formatAmount(req.Order.Amount),
		"OrderId":          req.Order.ID,
		"OkUrl":            req.Order.SuccessURL,
		"FailUrl":          req.Order.FailURL,
		"Rnd":              rnd,
		"Lang":             mapLang(g.cfg.Credentials.Lang),
		"Currency":         mapCurrency(currency),
		"InstallmentCount": mapInstallment(req.Order.Installment),
	}

	if req.Card != nil {
		if ct := mapCardType(req.Card.Brand); ct != "" {
			inputs["CardType"] = ct
		}
		inputs["Pan"] = req.Card.Number
		inputs["Expiry"] = cardExpiryMY(req.Card)
		inputs["Cvv2"] = req.Card.CVV
	} else if req.PaymentModel == model.PaymentModel3DHost {
		// 3D Host: kart formda müşteri tarafında
	} else if req.PaymentModel != model.PaymentModel3DPay && req.PaymentModel != model.PaymentModel3DPayHosting {
		return model.FormData{}, errors.New("interpos: kart bilgisi gerekli")
	}

	inputs["Hash"] = Create3DHash(g.cfg.Credentials.StoreKey, inputs)

	gatewayURL := g.cfg.Endpoints.Gateway3D
	if req.PaymentModel == model.PaymentModel3DHost {
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
		return model.PaymentResult{}, errors.New("interpos: boş callback payload")
	}

	order := orderFromPayload(payload)
	modelName := paymentModelFromPayload(payload)

	if !g.cfg.SkipHashCheck && !Check3DHash(g.cfg.Credentials.StoreKey, payload) {
		return model.PaymentResult{
			Success:      false,
			OrderID:      order.ID,
			ErrorCode:    "HASH_MISMATCH",
			ErrorMessage: "3D hash doğrulaması başarısız",
			RawResponse:  payloadToRaw(payload),
		}, nil
	}

	if !is3DAuthSuccess(payloadVal(payload, "3DStatus")) {
		return map3DFailed(payload, order), nil
	}

	switch modelName {
	case model.PaymentModel3DPay, model.PaymentModel3DPayHosting, model.PaymentModel3DHost:
		return map3DPayHostResult(payload, order), nil
	default:
		return g.complete3DSecure(ctx, payload, order)
	}
}

func (g *Gateway) complete3DSecure(ctx context.Context, payload map[string]string, order model.Order) (model.PaymentResult, error) {
	txType := model.TxTypePayAuth
	if payloadVal(payload, "TxnType") == "PreAuth" {
		txType = model.TxTypePayPreAuth
	}

	fields := accountFields(g.cfg.Credentials)
	for k, v := range map[string]string{
		"TxnType":                 txTypeToInter[txType],
		"SecureType":              modelToSecureType[model.PaymentModel3DSecure],
		"OrderId":                 firstNonEmpty(payloadVal(payload, "OrderId"), order.ID),
		"PurchAmount":             firstNonEmpty(payloadVal(payload, "PurchAmount"), formatAmount(order.Amount)),
		"Currency":                firstNonEmpty(payloadVal(payload, "Currency"), mapCurrency(order.Currency)),
		"InstallmentCount":        payloadVal(payload, "InstallmentCount"),
		"MD":                      payloadVal(payload, "MD"),
		"PayerTxnId":              payloadVal(payload, "PayerTxnId"),
		"Eci":                     payloadVal(payload, "Eci"),
		"PayerAuthenticationCode": payloadVal(payload, "PayerAuthenticationCode"),
		"MOTO":                    "0",
		"Lang":                    mapLang(g.cfg.Credentials.Lang),
	} {
		fields[k] = v
	}

	body, err := g.http.postForm(ctx, g.cfg.Endpoints.PaymentAPI, fields)
	if err != nil {
		return model.PaymentResult{}, err
	}
	raw, err := decodeResponse(body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return map3DResult(payload, raw, order), nil
}

// Cancel ödeme iptali.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	fields := accountFields(g.cfg.Credentials)
	fields["orgOrderId"] = req.Order.ID
	fields["TxnType"] = txTypeToInter[model.TxTypeCancel]
	fields["SecureType"] = modelToSecureType[model.PaymentModel3DSecure]
	fields["Lang"] = mapLang(g.cfg.Credentials.Lang)
	return g.postAndMap(ctx, fields, req.Order)
}

// Refund iade.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	tx := model.TxTypeRefund
	if req.Partial {
		tx = model.TxTypeRefundPartial
	}
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	fields := accountFields(g.cfg.Credentials)
	fields["orgOrderId"] = req.Order.ID
	fields["PurchAmount"] = formatAmount(req.Order.Amount)
	fields["TxnType"] = txTypeToInter[tx]
	fields["SecureType"] = modelToSecureType[model.PaymentModel3DSecure]
	fields["Lang"] = mapLang(g.cfg.Credentials.Lang)
	fields["MOTO"] = "0"
	return g.postAndMap(ctx, fields, req.Order)
}

// Status sipariş durumu.
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	fields := accountFields(g.cfg.Credentials)
	fields["orgOrderId"] = req.Order.ID
	fields["TxnType"] = txTypeToInter[model.TxTypeStatus]
	fields["SecureType"] = modelToSecureType[model.PaymentModel3DSecure]
	fields["Lang"] = mapLang(g.cfg.Credentials.Lang)
	return g.postAndMap(ctx, fields, req.Order)
}

func (g *Gateway) postAndMap(ctx context.Context, fields map[string]string, order model.Order) (model.PaymentResult, error) {
	body, err := g.http.postForm(ctx, g.cfg.Endpoints.PaymentAPI, fields)
	if err != nil {
		return model.PaymentResult{}, err
	}
	raw, err := decodeResponse(body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapAPIResult(raw, order), nil
}

