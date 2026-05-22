package posnet

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "posnet"

// Gateway Yapı Kredi PosNet — yalnızca 3D Secure.
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

// Init oosRequestData ile kayıt alır ve 3D formunu hazırlar (kart zorunlu).
func (g *Gateway) Init(ctx context.Context, req model.InitRequest) (model.FormData, error) {
	if err := validateInit(req); err != nil {
		return model.FormData{}, err
	}
	if req.PaymentModel != model.PaymentModel3DSecure {
		return model.FormData{}, fmt.Errorf("posnet: yalnızca 3D Secure desteklenir")
	}

	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	orderID, err := formatOrderID(req.Order.ID)
	if err != nil {
		return model.FormData{}, err
	}
	tx, ok := txTypeToPosnet[req.TxType]
	if !ok {
		return model.FormData{}, fmt.Errorf("posnet: desteklenmeyen işlem türü: %s", req.TxType)
	}

	holder := req.Card.HolderName
	xmlBody := encodeEnrollment(c.ClientID, c.Password, posNetID(c), oosRequestFields{
		CCNo:           req.Card.Number,
		ExpDate:        cardExpiryYM(req.Card),
		CVC:            req.Card.CVV,
		AmountKurus:    formatAmountKurus(req.Order.Amount),
		CurrencyCode:   mapCurrency(currency),
		Installment:    mapInstallment(req.Order.Installment),
		XID:            orderID,
		CardHolderName: holder,
		TranType:       tx,
	})

	body, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, xmlBody)
	if err != nil {
		return model.FormData{}, err
	}
	raw, err := decodeResponse(body)
	if err != nil {
		return model.FormData{}, err
	}
	if !approvedOK(raw) {
		return model.FormData{}, errors.New(payloadVal(raw, "respText"))
	}

	return model.FormData{
		Gateway: g.cfg.Endpoints.Gateway3D,
		Method:  http.MethodPost,
		Inputs: map[string]string{
			"mid":               c.ClientID,
			"posnetID":          posNetID(c),
			"posnetData":        payloadVal(raw, "data1"),
			"posnetData2":       payloadVal(raw, "data2"),
			"digest":            payloadVal(raw, "sign"),
			"merchantReturnURL": req.Order.SuccessURL,
			"url":               "",
			"lang":              mapLang(c.Lang),
		},
	}, nil
}

// HandleCallback 3D dönüşü: oosResolveMerchantData ardından oosTranData provizyonu.
// Payload'da banka alanları (BankPacket, MerchantPacket, Sign) ve oturumdan orderId, amount, currency bulunmalıdır.
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("posnet: boş callback payload")
	}

	order := orderFromPayload(payload)
	if order.ID == "" {
		return model.PaymentResult{}, errors.New("posnet: orderId gerekli (callback payload)")
	}
	if order.Amount == 0 {
		return model.PaymentResult{}, errors.New("posnet: amount gerekli (callback payload)")
	}

	bank := firstNonEmpty(payloadVal(payload, "BankPacket"), payloadVal(payload, "bankData"))
	merchant := firstNonEmpty(payloadVal(payload, "MerchantPacket"), payloadVal(payload, "merchantData"))
	sign := firstNonEmpty(payloadVal(payload, "Sign"), payloadVal(payload, "sign"))
	if bank == "" || merchant == "" || sign == "" {
		return model.PaymentResult{}, errors.New("posnet: BankPacket, MerchantPacket ve Sign gerekli")
	}

	c := g.cfg.Credentials
	mc := credsFromModel(c)
	orderID, err := formatOrderID(order.ID)
	if err != nil {
		return model.PaymentResult{}, err
	}
	amountKurus := formatAmountKurus(order.Amount)
	curr := mapCurrency(order.Currency)
	mac := CreateTransactionMAC(c.StoreKey, c.Password, orderID, amountKurus, curr, c.ClientID)

	resolveXML := encodeResolve(c.ClientID, c.Password, mac, bank, merchant, sign)
	resolveBody, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, resolveXML)
	if err != nil {
		return model.PaymentResult{}, err
	}
	resolveRaw, err := decodeResponse(resolveBody)
	if err != nil {
		return model.PaymentResult{}, err
	}

	md := payloadVal(resolveRaw, "mdStatus")
	if !is3DAuthSuccess(md) {
		return map3DFailed(payload, resolveRaw, order), nil
	}

	if !g.cfg.SkipHashCheck && !CheckResolveMAC(mc, resolveMACFields(resolveRaw)) {
		return model.PaymentResult{
			Success:      false,
			OrderID:      order.ID,
			ErrorCode:    "HASH_MISMATCH",
			ErrorMessage: "oosResolve MAC doğrulaması başarısız",
			RawResponse:  mergeRaw(payloadToRaw(payload), resolveRaw),
		}, nil
	}

	tranMAC := CreateTransactionMAC(c.StoreKey, c.Password, orderID, amountKurus, curr, c.ClientID)
	tranXML := encodeTran(c.ClientID, c.Password, tranMAC, bank, merchant, sign)
	tranBody, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, tranXML)
	if err != nil {
		return model.PaymentResult{}, err
	}
	tranRaw, err := decodeResponse(tranBody)
	if err != nil {
		return model.PaymentResult{}, err
	}

	return map3DResult(resolveRaw, tranRaw, order), nil
}

// Cancel ödeme iptali (reverse).
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	var orderID string
	if req.Order.RefRetNum == "" && req.Order.ID != "" {
		pref, err := prefixedOrderID(req.Order.ID)
		if err != nil {
			return model.PaymentResult{}, err
		}
		orderID = pref
	}
	xmlBody := encodeCancel(c.ClientID, c.Password, req.Order.RefRetNum, orderID, "")
	return g.postAndMap(ctx, xmlBody, req.Order)
}

// Refund iade (return).
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	var orderID string
	if req.Order.RefRetNum == "" && req.Order.ID != "" {
		pref, err := prefixedOrderID(req.Order.ID)
		if err != nil {
			return model.PaymentResult{}, err
		}
		orderID = pref
	}
	xmlBody := encodeRefund(
		c.ClientID, c.Password,
		formatAmountKurus(req.Order.Amount),
		mapCurrency(currency),
		req.Order.RefRetNum, orderID,
	)
	return g.postAndMap(ctx, xmlBody, req.Order)
}

// Status sipariş durumu (agreement).
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	if req.Order.ID == "" {
		return model.PaymentResult{}, errors.New("posnet: sipariş ID gerekli")
	}
	pref, err := prefixedOrderID(req.Order.ID)
	if err != nil {
		return model.PaymentResult{}, err
	}
	c := g.cfg.Credentials
	xmlBody := encodeStatus(c.ClientID, c.Password, pref)
	return g.postAndMap(ctx, xmlBody, req.Order)
}

func (g *Gateway) postAndMap(ctx context.Context, xmlBody []byte, order model.Order) (model.PaymentResult, error) {
	body, err := g.http.postXML(ctx, g.cfg.Endpoints.PaymentAPI, xmlBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	raw, err := decodeResponse(body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapAPIResult(raw, order), nil
}

