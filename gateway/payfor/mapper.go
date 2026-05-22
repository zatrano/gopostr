package payfor

import (
	"strconv"

	"github.com/zatrano/gopostr/model"
)

func is3DAuthSuccess(mdStatus string) bool {
	return mdStatus == "1"
}

func paymentModelFromPayload(p map[string]string) string {
	st := payloadVal(p, "SecureType")
	if m, ok := secureToModel[st]; ok {
		return m
	}
	return model.PaymentModel3DSecure
}

func is3DPayOrHost(modelName string) bool {
	return modelName == model.PaymentModel3DPay || modelName == model.PaymentModel3DHost
}

func mapAPIResult(resp payforResponse, raw map[string]string, order model.Order) model.PaymentResult {
	success := resp.ProcReturnCode == procSuccess
	result := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(resp.TransId, resp.OrderId, order.ID),
		TransactionID: resp.TransId,
		AuthCode:      resp.AuthCode,
		HostRefNum:    firstNonEmpty(resp.HostRefNum, resp.TransId),
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
		OrderID:       payloadVal(payload, "OrderId"),
		TransactionID: payloadVal(payload, "TransId"),
		AuthCode:      payloadVal(payload, "AuthCode"),
		HostRefNum:    payloadVal(payload, "HostRefNum"),
		Amount:        parseAmount(payloadVal(payload, "PurchAmount"), order.Amount),
		Currency:      parseCurrency(payloadVal(payload, "Currency")),
		ErrorCode:     proc,
		ErrorMessage:  payloadVal(payload, "ErrMsg"),
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
		OrderID:      payloadVal(payload, "OrderId"),
		ErrorCode:    payloadVal(payload, "ProcReturnCode"),
		ErrorMessage: payloadVal(payload, "ErrMsg"),
		RawResponse:  payloadToRaw(payload),
	}
}

func mapStatusResult(resp payforResponse, raw map[string]string, order model.Order) model.PaymentResult {
	success := resp.ProcReturnCode == procSuccess
	result := model.PaymentResult{
		Success:      success,
		OrderID:      firstNonEmpty(resp.OrderId, order.ID),
		AuthCode:     resp.AuthCode,
		HostRefNum:   resp.HostRefNum,
		ErrorCode:    resp.ProcReturnCode,
		ErrorMessage: resp.ErrMsg,
		RawResponse:  raw,
	}
	if success {
		result.ErrorCode = ""
		result.ErrorMessage = ""
		if resp.PurchAmount != "" {
			result.Amount, _ = strconv.ParseFloat(resp.PurchAmount, 64)
		}
		result.Currency = parseCurrency(resp.Currency)
	}
	return result
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

func parseAmount(s string, fallback float64) float64 {
	if s == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fallback
	}
	return f
}
