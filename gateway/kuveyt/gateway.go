package kuveyt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/model"
)

const gatewayName = "kuveyt"

// Gateway Kuveyt Türk sanal POS — yalnızca 3D Secure (kartlı satış).
type Gateway struct {
	cfg  Config
	http *httpClient
}

// New yapılandırılmış gateway döner.
func New(cfg Config) gateway.Gateway {
	if cfg.Endpoints.PaymentAPI == "" && cfg.Endpoints.Gateway3D == "" && cfg.Endpoints.QueryAPI == "" {
		cfg.Endpoints = DefaultTestEndpoints
	}
	return &Gateway{cfg: cfg, http: newHTTPClient()}
}

// Name gateway adını döner.
func (g *Gateway) Name() string { return gatewayName }

// Init enrollment isteği gönderir; bankanın döndürdüğü HTML 3D formunu parse eder.
func (g *Gateway) Init(ctx context.Context, req model.InitRequest) (model.FormData, error) {
	if err := validateInit(req); err != nil {
		return model.FormData{}, err
	}
	if req.PaymentModel != model.PaymentModel3DSecure {
		return model.FormData{}, fmt.Errorf("kuveyt: yalnızca 3D Secure desteklenir")
	}
	if req.TxType != model.TxTypePayAuth {
		return model.FormData{}, fmt.Errorf("kuveyt: yalnızca satış (pay_auth) desteklenir")
	}
	if req.Card == nil {
		return model.FormData{}, errors.New("kuveyt: kart bilgisi zorunlu")
	}

	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	amount := formatAmountKurus(req.Order.Amount)
	inst, _ := strconv.Atoi(mapInstallment(req.Order.Installment))
	ip := req.Order.IP
	if ip == "" {
		ip = "127.0.0.1"
	}
	m, y := cardExpiryParts(req.Card)

	hashIn := hashFieldsFrom(c, map[string]string{
		"MerchantOrderId": req.Order.ID,
		"Amount":          fmt.Sprintf("%d", amount),
		"OkUrl":           req.Order.SuccessURL,
		"FailUrl":         req.Order.FailURL,
	})
	hash := CreateHash(c.StoreKey, hashIn)

	xmlBody := encodeEnrollment(enrollmentParams{
		APIVersion:      apiVersion,
		HashData:        hash,
		TxnType:         "Sale",
		TxSecurity:      "3",
		Installment:     inst,
		Amount:          amount,
		DisplayAmount:   amount,
		CurrencyCode:    mapCurrency(currency),
		MerchantOrderID: req.Order.ID,
		OkURL:           req.Order.SuccessURL,
		FailURL:         req.Order.FailURL,
		ClientIP:        ip,
		MerchantID:      c.ClientID,
		CustomerID:      c.Password,
		UserName:        c.Username,
		CardHolder:      req.Card.HolderName,
		CardType:        mapCardType(req.Card.Brand),
		CardNumber:      req.Card.Number,
		CardExpMonth:    m,
		CardExpYear:     y,
		CardCVV:         req.Card.CVV,
	})

	body, err := g.http.postXML(ctx, g.cfg.Endpoints.Gateway3D, xmlBody)
	if err != nil {
		return model.FormData{}, err
	}
	if !isHTMLResponse(body) {
		return model.FormData{}, fmt.Errorf("kuveyt: enrollment HTML yanıt bekleniyordu")
	}
	gw, inputs, err := parseHTMLForm(string(body))
	if err != nil {
		return model.FormData{}, err
	}
	return model.FormData{Gateway: gw, Method: http.MethodPost, Inputs: inputs}, nil
}

// HandleCallback AuthenticationResponse XML işler ve provizyon tamamlar.
func (g *Gateway) HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error) {
	if len(payload) == 0 {
		return model.PaymentResult{}, errors.New("kuveyt: boş callback payload")
	}
	authRaw := payloadVal(payload, "AuthenticationResponse")
	if authRaw == "" {
		return model.PaymentResult{}, errors.New("kuveyt: AuthenticationResponse gerekli")
	}
	decoded, err := url.QueryUnescape(authRaw)
	if err != nil {
		decoded = authRaw
	}
	authMap, err := decodeXML([]byte(decoded))
	if err != nil {
		return model.PaymentResult{}, err
	}
	order := orderFromPayload(authMap)
	if order.ID == "" {
		order.ID = payloadVal(payload, "MerchantOrderId")
	}

	if !is3DAuthSuccess(payloadVal(authMap, "ResponseCode")) {
		return map3DAuthFailed(authMap, order), nil
	}

	md := payloadVal(authMap, "MD")
	if md == "" {
		return model.PaymentResult{}, errors.New("kuveyt: MD eksik")
	}

	vpos := vposSubMap(authMap, "VPosMessage")
	amount, _ := strconv.Atoi(firstNonEmpty(vpos["Amount"], payloadVal(authMap, "Amount")))
	install, _ := strconv.Atoi(firstNonEmpty(vpos["InstallmentCount"], "0"))
	curr := firstNonEmpty(vpos["CurrencyCode"], payloadVal(authMap, "CurrencyCode"))
	txSec := firstNonEmpty(vpos["TransactionSecurity"], "3")
	merchantOrder := firstNonEmpty(vpos["MerchantOrderId"], order.ID)

	c := g.cfg.Credentials
	hashIn := hashFieldsFrom(c, map[string]string{
		"MerchantOrderId": merchantOrder,
		"Amount":          fmt.Sprintf("%d", amount),
	})
	provHash := CreateHash(c.StoreKey, hashIn)

	provXML := encodeProvision(provisionParams{
		APIVersion:      apiVersion,
		HashData:        provHash,
		TxnType:         "Sale",
		TxSecurity:      txSec,
		Installment:     install,
		Amount:          amount,
		DisplayAmount:   amount,
		CurrencyCode:    curr,
		MerchantOrderID: merchantOrder,
		MD:              md,
		ClientIP:        firstNonEmpty(order.IP, "127.0.0.1"),
		MerchantID:      c.ClientID,
		CustomerID:      c.Password,
		UserName:        c.Username,
	})

	provURL := strings.TrimRight(g.cfg.Endpoints.PaymentAPI, "/") + "/ThreeDModelProvisionGate"
	provBody, err := g.http.postXML(ctx, provURL, provXML)
	if err != nil {
		return model.PaymentResult{}, err
	}
	payMap, err := decodeXML(provBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return map3DResult(authMap, payMap, order), nil
}

// Cancel ödeme iptali (SOAP SaleReversal).
func (g *Gateway) Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error) {
	return g.soapCancelRefund(ctx, req.Order, false, model.TxTypeCancel)
}

// Refund iade (SOAP DrawBack / PartialDrawback).
func (g *Gateway) Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error) {
	return g.soapCancelRefund(ctx, req.Order, req.Partial, model.TxTypeRefund)
}

// Status sipariş durumu (SOAP GetMerchantOrderDetail).
func (g *Gateway) Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	currency := req.Order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	start, end := defaultStatusDates()
	vpos := soapVPos{
		APIVersion:      apiVersion,
		TxnType:         "GetMerchantOrderDetail",
		TxSecurity:      "1",
		CardType:        "Visa",
		Installment:     "0",
		Amount:          0,
		DisplayAmount:   0,
		CurrencyCode:    mapCurrency(currency),
		MerchantOrderID: req.Order.ID,
	}
	vpos.HashData = CreateHash(c.StoreKey, hashFieldsFrom(c, map[string]string{
		"MerchantOrderId": req.Order.ID,
		"Amount":          "0",
	}))
	soapBody := encodeSOAPStatus(soapStatusRequest{
		CustomerID:      c.Password,
		MerchantID:      c.ClientID,
		MerchantOrderID: req.Order.ID,
		BankOrderID:     bankOrderID(req.Order),
		StartDate:       start,
		EndDate:         end,
		VPos:            vpos,
	})
	body, err := g.http.postSOAP(ctx, g.cfg.Endpoints.QueryAPI, soapActionFor("GetMerchantOrderDetail"), soapBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	raw, err := decodeSOAP(body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapSOAPStatus(raw, req.Order.ID), nil
}

func (g *Gateway) soapCancelRefund(ctx context.Context, order model.Order, partial bool, op string) (model.PaymentResult, error) {
	c := g.cfg.Credentials
	currency := order.Currency
	if currency == "" {
		currency = model.CurrencyTRY
	}
	amount := formatAmountKurus(order.Amount)
	txn := "SaleReversal"
	soapAction := "SaleReversal"
	if op == model.TxTypeRefund {
		txn = "DrawBack"
		soapAction = "DrawBack"
		if partial {
			txn = "PartialDrawback"
			soapAction = "PartialDrawback"
		}
	}
	vpos := soapVPos{
		APIVersion:      apiVersion,
		TxnType:         txn,
		TxSecurity:      "1",
		CardType:        "Visa",
		Installment:     "0",
		Amount:          amount,
		DisplayAmount:   amount,
		CancelAmount:    amount,
		CurrencyCode:    mapCurrency(currency),
		MerchantOrderID: order.ID,
	}
	vpos.HashData = CreateHash(c.StoreKey, hashFieldsFrom(c, map[string]string{
		"MerchantOrderId": order.ID,
		"Amount":          fmt.Sprintf("%d", amount),
	}))
	soapReq := soapCancelRequest{
		CustomerID:  c.Password,
		MerchantID:  c.ClientID,
		Amount:      amount,
		OrderID:     order.ID,
		BankOrderID: bankOrderID(order),
		RRN:         order.RefRetNum,
		Stan:        order.TransactionID,
		VPos:        vpos,
	}
	var soapBody []byte
	if op == model.TxTypeCancel {
		soapBody = encodeSOAPCancel(soapReq)
	} else {
		soapBody = encodeSOAPRefund(soapReq, partial)
	}
	body, err := g.http.postSOAP(ctx, g.cfg.Endpoints.QueryAPI, soapActionFor(soapAction), soapBody)
	if err != nil {
		return model.PaymentResult{}, err
	}
	raw, err := decodeSOAP(body)
	if err != nil {
		return model.PaymentResult{}, err
	}
	return mapSOAPCancelRefund(raw, order.ID), nil
}

func validateInit(req model.InitRequest) error {
	if req.Order.ID == "" {
		return errors.New("kuveyt: sipariş ID gerekli")
	}
	if req.Order.Amount <= 0 {
		return errors.New("kuveyt: tutar gerekli")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errors.New("kuveyt: success/fail URL gerekli")
	}
	return nil
}
