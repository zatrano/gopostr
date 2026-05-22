package estpos

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/crypt"
	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "estpos"

// Gateway Payten/Asseco EstPOS (EstV3, SHA-512) — TEB, İş Bankası, Halkbank, Şekerbank, Ziraat, eski Akbank.
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
	storeType, ok := modelToStoreType[req.PaymentModel]
	if !ok {
		return model.FormData{}, fmt.Errorf("estpos: desteklenmeyen 3D modeli: %s", req.PaymentModel)
	}
	tx, ok := txTypeToEst[req.TxType]
	if !ok || (req.TxType != model.TxTypePayAuth && req.TxType != model.TxTypePayPreAuth) {
		return model.FormData{}, fmt.Errorf("estpos: desteklenmeyen işlem türü: %s", req.TxType)
	}

	rnd, err := crypt.RandomString(20)
	if err != nil {
		return model.FormData{}, err
	}

	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}

	inputs := map[string]string{
		"clientid":  c.ClientID,
		"storetype": storeType,
		"amount":    formatAmount(req.Order.Amount),
		"oid":       req.Order.ID,
		"okUrl":     req.Order.SuccessURL,
		"failUrl":   req.Order.FailURL,
		"rnd":       rnd,
		"lang":      mapLang(c.Lang),
		"currency":  mapCurrency(currency),
		"taksit":    mapInstallment(req.Order.Installment),
		"islemtipi": tx,
	}

	if req.Card != nil {
		inputs["pan"] = req.Card.Number
		inputs["Ecom_Payment_Card_ExpDate_Month"] = cardExpMonth(req.Card)
		inputs["Ecom_Payment_Card_ExpDate_Year"] = cardExpYear(req.Card)
		inputs["cv2"] = req.Card.CVV
	} else if req.PaymentModel == model.PaymentModel3DHost {
		// 3D Host: kart banka sayfasında
	} else if req.PaymentModel != model.PaymentModel3DPay && req.PaymentModel != model.PaymentModel3DPayHosting {
		return model.FormData{}, errors.New("estpos: kart bilgisi gerekli")
	}

	inputs["hash"] = Create3DHash(c.StoreKey, inputs)

	return model.FormData{
		Gateway: g.cfg.Endpoints.Gateway3D,
		Method:  http.MethodPost,
		Inputs:  inputs,
	}, nil
}

// HandleCallback banka 3D dönüşünü işler.
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("estpos: boş callback payload")
	}

	order := model.Order{
		ID:       firstNonEmpty(payloadVal(payload, "oid"), payloadVal(payload, "OrderId")),
		Amount:   parseAmount(payloadVal(payload, "amount"), 0),
		Currency: parseCurrency(payloadVal(payload, "currency")),
	}

	modelName := paymentModelFromPayload(payload)
	if is3DPayOrHost(modelName) {
		if !g.cfg.SkipHashCheck && !Check3DHash(g.cfg.Credentials.StoreKey, payload) {
			return model.PaymentResult{
				Success:      false,
				OrderID:      order.ID,
				ErrorCode:    "HASH_MISMATCH",
				ErrorMessage: "3D hash doğrulaması başarısız",
				RawResponse:  payloadToRaw(payload),
			}, nil
		}
		return map3DPayHostResult(payload, order), nil
	}

	if !is3DAuthSuccess(payloadVal(payload, "mdStatus")) {
		return map3DFailed(payload), nil
	}

	if !g.cfg.SkipHashCheck && !Check3DHash(g.cfg.Credentials.StoreKey, payload) {
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
	tx := payloadVal(payload, "islemtipi")
	if tx == "" {
		tx = "Auth"
	}
	currency := order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	ip := order.IP
	if ip == "" {
		ip = "127.0.0.1"
	}

	req := cc5Request{
		Name:                    c.Username,
		Password:                c.Password,
		ClientID:                c.ClientID,
		Type:                    tx,
		OrderID:                 order.ID,
		Total:                   formatAmount(order.Amount),
		Currency:                mapCurrency(currency),
		Taksit:                  payloadVal(payload, "taksit"),
		IPAddress:               ip,
		Number:                  payloadVal(payload, "md"),
		PayerTxnID:              payloadVal(payload, "xid"),
		PayerSecurityLevel:      payloadVal(payload, "eci"),
		PayerAuthenticationCode: payloadVal(payload, "cavv"),
		Mode:                    "P",
	}
	body, err := encodeRequest(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.post(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeResponse(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	for k, v := range payloadToRaw(payload) {
		raw["callback_"+k] = v
	}
	return mapAPIResult(resp, raw, order), nil
}

// Cancel ödeme iptali.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	return g.apiOp(ctx, req.Order, txTypeToEst[model.TxTypeCancel], "")
}

// Refund ödeme iadesi.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	amount := ""
	if req.Partial && req.Order.Amount > 0 {
		amount = formatAmount(req.Order.Amount)
	}
	return g.apiOp(ctx, req.Order, txTypeToEst[model.TxTypeRefund], amount)
}

// Status sipariş sorgulama.
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	body, err := encodeRequest(cc5Request{
		Name:     c.Username,
		Password: c.Password,
		ClientID: c.ClientID,
		OrderID:  req.Order.ID,
		Extra:    &cc5Extra{OrderStatus: "QUERY"},
	})
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.post(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeResponse(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapStatusResult(resp, raw, req.Order), nil
}

func (g *Gateway) apiOp(ctx context.Context, order model.Order, txType, amount string) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	currency := order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	req := cc5Request{
		Name:     c.Username,
		Password: c.Password,
		ClientID: c.ClientID,
		Type:     txType,
		OrderID:  order.ID,
		Currency: mapCurrency(currency),
	}
	if amount != "" {
		req.Total = amount
	}
	body, err := encodeRequest(req)
	if err != nil {
		return model.PaymentResult{}, err
	}
	respBody, err := g.http.post(ctx, g.cfg.Endpoints.PaymentAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resp, raw, err := decodeResponse(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapAPIResult(resp, raw, order), nil
}
