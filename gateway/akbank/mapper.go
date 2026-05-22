package akbank

import (
	"strconv"

	"github.com/zatrano/gopostr/model"
)

var modelToPayment = map[string]string{
	model.PaymentModel3DSecure:     "3D",
	model.PaymentModel3DPay:        "3D_PAY",
	model.PaymentModel3DHost:       "3D_PAY_HOSTING",
	model.PaymentModel3DPayHosting: "3D_PAY_HOSTING",
}

var paymentToModel = map[string]string{
	"3D":             model.PaymentModel3DSecure,
	"3D_PAY":         model.PaymentModel3DPay,
	"3D_PAY_HOSTING": model.PaymentModel3DHost,
}

var txCodeByModel = map[string]map[string]string{
	model.TxTypePayAuth: {
		model.PaymentModel3DSecure: "3000",
		model.PaymentModel3DPay:    "3000",
		model.PaymentModel3DHost:   "3000",
	},
	model.TxTypePayPreAuth: {
		model.PaymentModel3DSecure: "3004",
		model.PaymentModel3DPay:    "3004",
		model.PaymentModel3DHost:   "3004",
	},
}

const (
	txCodeSale    = "1000"
	txCodePreAuth = "1004"
	txCodeCancel  = "1003"
	txCodeRefund  = "1002"
)

func is3DAuthSuccess(responseCode string) bool {
	return responseCode == procSuccess
}

func paymentModelFromPayload(p map[string]string) string {
	pm := payloadVal(p, "paymentModel")
	if m, ok := paymentToModel[pm]; ok {
		return m
	}
	return model.PaymentModel3DSecure
}

func mapAPIResult(resp apiResponse, raw map[string]string, order model.Order) model.PaymentResult {
	success := resp.ResponseCode == procSuccess
	result := model.PaymentResult{
		Success:      success,
		OrderID:      firstNonEmpty(resp.Order.OrderID, order.ID),
		AuthCode:     resp.Transaction.AuthCode,
		HostRefNum:   resp.Transaction.Rrn,
		Amount:       order.Amount,
		Currency:     order.Currency,
		ErrorCode:    resp.ResponseCode,
		ErrorMessage: firstNonEmpty(resp.HostMessage, resp.ResponseMessage),
		RawResponse:  raw,
	}
	if success {
		result.ErrorCode = ""
		result.ErrorMessage = ""
	}
	return result
}

func map3DPayHostResult(payload map[string]string, order model.Order) model.PaymentResult {
	code := payloadVal(payload, "responseCode")
	success := is3DAuthSuccess(code)
	result := model.PaymentResult{
		Success:      success,
		OrderID:      firstNonEmpty(payloadVal(payload, "orderId"), order.ID),
		Amount:       order.Amount,
		Currency:     order.Currency,
		AuthCode:     payloadVal(payload, "authCode"),
		HostRefNum:   payloadVal(payload, "rrn"),
		ErrorCode:    code,
		ErrorMessage: payloadVal(payload, "responseMessage"),
		RawResponse:  payloadToRaw(payload),
	}
	if success {
		result.ErrorCode = ""
		result.ErrorMessage = ""
	}
	return result
}

func map3DFailed(payload map[string]string) model.PaymentResult {
	return model.PaymentResult{
		Success:      false,
		OrderID:      payloadVal(payload, "orderId"),
		ErrorCode:    payloadVal(payload, "responseCode"),
		ErrorMessage: payloadVal(payload, "responseMessage"),
		RawResponse:  payloadToRaw(payload),
	}
}

func orderFromPayload(p map[string]string) model.Order {
	amount := 0.0
	if s := payloadVal(p, "amount"); s != "" {
		amount, _ = strconv.ParseFloat(s, 64)
	}
	inst := 0
	if s := payloadVal(p, "installCount"); s != "" {
		inst, _ = strconv.Atoi(s)
	}
	return model.Order{
		ID:          payloadVal(p, "orderId"),
		Amount:      amount,
		Currency:    parseCurrency(payloadVal(p, "currencyCode")),
		Installment: inst,
	}
}

func payloadToRaw(p map[string]string) map[string]string {
	if p == nil {
		return nil
	}
	out := make(map[string]string, len(p))
	for k, v := range p {
		out[k] = v
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func is3DPayOrHost(modelName string) bool {
	return modelName == model.PaymentModel3DPay ||
		modelName == model.PaymentModel3DHost ||
		modelName == model.PaymentModel3DPayHosting
}

func txnCodeProvision(txType string) string {
	if txType == model.TxTypePayPreAuth {
		return txCodePreAuth
	}
	return txCodeSale
}
