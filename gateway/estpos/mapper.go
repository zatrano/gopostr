package estpos

import (
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

const procSuccess = "00"

var modelToStoreType = map[string]string{
	model.PaymentModel3DSecure:     "3d",
	model.PaymentModel3DPay:        "3d_pay",
	model.PaymentModel3DHost:       "3d_host",
	model.PaymentModel3DPayHosting: "3d_pay_hosting",
}

var storeTypeToModel = map[string]string{
	"3d":               model.PaymentModel3DSecure,
	"3d_pay":           model.PaymentModel3DPay,
	"3d_host":          model.PaymentModel3DHost,
	"3d_pay_hosting":   model.PaymentModel3DPayHosting,
}

var txTypeToEst = map[string]string{
	model.TxTypePayAuth:       "Auth",
	model.TxTypePayPreAuth:    "PreAuth",
	model.TxTypeCancel:        "Void",
	model.TxTypeRefund:        "Credit",
	model.TxTypeRefundPartial: "Credit",
}

func is3DAuthSuccess(mdStatus string) bool {
	return mdStatus == "1"
}

func paymentModelFromPayload(p map[string]string) string {
	st := strings.ToLower(payloadVal(p, "storetype"))
	if m, ok := storeTypeToModel[st]; ok {
		return m
	}
	return model.PaymentModel3DSecure
}

func is3DPayOrHost(modelName string) bool {
	switch modelName {
	case model.PaymentModel3DPay, model.PaymentModel3DHost, model.PaymentModel3DPayHosting:
		return true
	default:
		return false
	}
}

func mapAPIResult(resp cc5Response, raw map[string]string, order model.Order) model.PaymentResult {
	success := resp.ProcReturnCode == procSuccess
	result := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(resp.OrderID, order.ID),
		TransactionID: resp.TransID,
		AuthCode:      resp.AuthCode,
		HostRefNum:    firstNonEmpty(resp.HostRefNum, resp.TransID),
		Amount:        order.Amount,
		Currency:      order.Currency,
		ErrorCode:     resp.ProcReturnCode,
		ErrorMessage:  resp.ErrMsg,
		RawResponse:   raw,
	}
	if success {
		result.ErrorCode = ""
		result.ErrorMessage = ""
	}
	return result
}

func map3DPayHostResult(payload map[string]string, order model.Order) model.PaymentResult {
	proc := payloadVal(payload, "ProcReturnCode")
	success := proc == procSuccess
	result := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(payloadVal(payload, "oid"), order.ID),
		TransactionID: payloadVal(payload, "TransId"),
		AuthCode:      payloadVal(payload, "AuthCode"),
		HostRefNum:    payloadVal(payload, "HostRefNum"),
		Amount:        parseAmount(payloadVal(payload, "amount"), order.Amount),
		Currency:      parseCurrency(payloadVal(payload, "currency")),
		ErrorCode:     proc,
		ErrorMessage:  firstNonEmpty(payloadVal(payload, "ErrMsg"), payloadVal(payload, "mdErrorMsg")),
		RawResponse:   payloadToRaw(payload),
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
		OrderID:      firstNonEmpty(payloadVal(payload, "oid"), payloadVal(payload, "OrderId")),
		ErrorCode:    firstNonEmpty(payloadVal(payload, "ProcReturnCode"), payloadVal(payload, "ErrCode")),
		ErrorMessage: firstNonEmpty(payloadVal(payload, "ErrMsg"), payloadVal(payload, "mdErrorMsg")),
		RawResponse:  payloadToRaw(payload),
	}
}

func mapStatusResult(resp cc5Response, raw map[string]string, order model.Order) model.PaymentResult {
	success := resp.ProcReturnCode == procSuccess
	result := model.PaymentResult{
		Success:      success,
		OrderID:      firstNonEmpty(resp.OrderID, order.ID),
		AuthCode:     resp.AuthCode,
		HostRefNum:   resp.HostRefNum,
		ErrorCode:    resp.ProcReturnCode,
		ErrorMessage: resp.ErrMsg,
		RawResponse:  raw,
	}
	if success {
		result.ErrorCode = ""
		result.ErrorMessage = ""
		if resp.Total != "" {
			result.Amount, _ = strconv.ParseFloat(resp.Total, 64)
		}
	}
	return result
}

func payloadToRaw(p map[string]string) map[string]string {
	out := make(map[string]string, len(p))
	for k, v := range p {
		out[k] = v
	}
	return out
}
