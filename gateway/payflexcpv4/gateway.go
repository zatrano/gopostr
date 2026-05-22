package payflexcpv4

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "payflexcpv4"

// Gateway Vakıfbank PayFlex CP v4 (3D Secure / 3D Pay / 3D Host — IsSecure=true).
type Gateway struct {
	cfg  Config
	http *httpClient
}

// New yapılandırılmış gateway döner.
func New(cfg Config) gateway.Gateway {
	ep := cfg.Endpoints
	if ep.RegisterAPI == "" {
		ep.RegisterAPI = DefaultTestEndpoints.RegisterAPI
	}
	if ep.VposAPI == "" {
		ep.VposAPI = DefaultTestEndpoints.VposAPI
	}
	cfg.Endpoints = ep
	return &Gateway{cfg: cfg, http: newHTTPClient()}
}

// Name gateway adını döner.
func (g *Gateway) Name() string { return gatewayName }

// Init RegisterTransaction ile CP oturumu açar.
func (g *Gateway) Init(ctx context.Context, req model.InitRequest) (model.FormData, error) {
	if err := validateInit(req); err != nil {
		return model.FormData{}, err
	}
	c := g.cfg.Credentials
	if c.ClientID == "" || c.Password == "" || c.TerminalID == "" {
		return model.FormData{}, errors.New("payflexcpv4: ClientID, Password ve TerminalID zorunlu")
	}

	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	amountCode := mapCurrency(currency)
	amount := formatAmount(req.Order.Amount)

	fields := map[string]string{
		"HostMerchantId":       c.ClientID,
		"MerchantPassword":     c.Password,
		"HostTerminalId":       c.TerminalID,
		"TransactionType":      "Sale",
		"AmountCode":           amountCode,
		"Amount":               amount,
		"OrderID":              req.Order.ID,
		"IsSecure":             "true",
		"AllowNotEnrolledCard": "false",
		"SuccessUrl":           req.Order.SuccessURL,
		"FailUrl":              req.Order.FailURL,
		"RequestLanguage":      mapLang(c.Lang),
		"Extract":              "",
		"CustomItems":          "",
	}
	if req.Order.Installment > 1 {
		fields["InstallmentCount"] = mapInstallment(req.Order.Installment)
	}
	if req.Card != nil {
		if brand := mapBrand(req.Card.Brand); brand != "" {
			fields["BrandNumber"] = brand
		}
		fields["CVV"] = req.Card.CVV
		fields["PAN"] = req.Card.Number
		fields["ExpireMonth"] = req.Card.ExpireMonth
		fields["ExpireYear"] = cardExpireYear(req.Card)
		fields["CardHoldersName"] = req.Card.HolderName
	}
	fields["HashedData"] = CreateEnrollmentHash(c.ClientID, amountCode, amount, c.Password)

	body := encodeForm(fields)
	respBody, err := g.http.postForm(ctx, g.cfg.Endpoints.RegisterAPI, body)
	if err != nil {
		return model.FormData{}, err
	}
	raw, err := decodeXML(respBody)
	if err != nil {
		return model.FormData{}, err
	}
	if ec := payloadVal(raw, "ErrorCode"); ec != "" {
		return model.FormData{}, fmt.Errorf("payflexcpv4: %s",
			firstNonEmpty(payloadVal(raw, "ResponseMessage"), "kayıt başarısız"))
	}
	url := payloadVal(raw, "CommonPaymentUrl")
	token := payloadVal(raw, "PaymentToken")
	if url == "" || token == "" {
		return model.FormData{}, errors.New("payflexcpv4: CommonPaymentUrl veya PaymentToken boş")
	}
	return model.FormData{
		Gateway: url,
		Method:  http.MethodGet,
		Inputs:  map[string]string{"Ptkn": token},
	}, nil
}

// HandleCallback banka dönüşünü VposTransaction ile doğrular.
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("payflexcpv4: boş callback payload")
	}
	order := model.Order{ID: payloadVal(payload, "OrderID")}
	if rc := payloadVal(payload, "Rc"); rc != "" && rc != procSuccess {
		return mapEarlyDecline(payload, order), nil
	}
	txID := payloadVal(payload, "TransactionId")
	token := payloadVal(payload, "PaymentToken")
	if txID == "" || token == "" {
		return model.PaymentResult{}, errors.New("payflexcpv4: TransactionId ve PaymentToken gerekli")
	}
	return g.queryPayment(ctx, txID, token, payload, order)
}

func (g *Gateway) queryPayment(ctx context.Context, txID, token string, callback map[string]string, order model.Order) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	fields := map[string]string{
		"HostMerchantId": c.ClientID,
		"Password":       c.Password,
		"TransactionId":  txID,
		"PaymentToken":   token,
	}
	body := encodeForm(fields)
	respBody, err := g.http.postForm(ctx, g.cfg.Endpoints.VposAPI, body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	raw, err := decodeXML(respBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	res := mapPaymentResult(raw, order)
	res.RawResponse = mergeRaw(callback, raw)
	return res, nil
}

// Cancel bu gateway'de desteklenmiyor.
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	_ = ctx
	_ = req
	return model.PaymentResult{}, errors.New("payflexcpv4: iptal desteklenmiyor")
}

// Refund bu gateway'de desteklenmiyor.
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	_ = ctx
	_ = req
	return model.PaymentResult{}, errors.New("payflexcpv4: iade desteklenmiyor")
}

// Status bu gateway'de desteklenmiyor (callback içinde VposTransaction kullanılır).
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	_ = ctx
	_ = req
	return model.PaymentResult{}, errors.New("payflexcpv4: durum sorgulama desteklenmiyor")
}
