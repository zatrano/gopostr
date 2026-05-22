package payflexcpv4

import (
	"github.com/zatrano/gopostr/model"
)

const procSuccess = "0000"

func payloadToRaw(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func mergeRaw(a, b map[string]string) map[string]string {
	out := payloadToRaw(a)
	for k, v := range b {
		out[k] = v
	}
	return out
}

func mapPaymentResult(raw map[string]string, order model.Order) model.PaymentResult {
	code := firstNonEmpty(payloadVal(raw, "Rc"), payloadVal(raw, "ResultCode"))
	errCode := payloadVal(raw, "ErrorCode")
	ok := code == procSuccess && errCode == ""
	if errCode != "" {
		ok = false
		code = errCode
	}
	res := model.PaymentResult{
		Success:     ok,
		OrderID:     firstNonEmpty(payloadVal(raw, "OrderID"), order.ID),
		RawResponse: payloadToRaw(raw),
	}
	if ok {
		res.TransactionID = payloadVal(raw, "TransactionId")
		res.AuthCode = payloadVal(raw, "AuthCode")
		res.HostRefNum = payloadVal(raw, "TransactionId")
		res.Amount = parseAmount(payloadVal(raw, "Amount"))
		if res.Amount == 0 {
			res.Amount = order.Amount
		}
	} else {
		res.ErrorCode = code
		res.ErrorMessage = firstNonEmpty(
			payloadVal(raw, "Message"),
			payloadVal(raw, "ResponseMessage"),
		)
	}
	return res
}

func mapEarlyDecline(payload map[string]string, order model.Order) model.PaymentResult {
	return model.PaymentResult{
		Success:      false,
		OrderID:      firstNonEmpty(payloadVal(payload, "OrderID"), order.ID),
		ErrorCode:    payloadVal(payload, "Rc"),
		ErrorMessage: payloadVal(payload, "Message"),
		RawResponse:  payloadToRaw(payload),
	}
}
