package posnetv1

import (
	"github.com/zatrano/gopostr/model"
)

const procSuccess = "00"
const procStatusSuccess = "0000"

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

func map3DFailed(payload map[string]string, order model.Order) model.PaymentResult {
	return model.PaymentResult{
		Success:      false,
		OrderID:      firstNonEmpty(payloadVal(payload, "OrderId"), order.ID),
		ErrorCode:    payloadVal(payload, "MdStatus"),
		ErrorMessage: payloadVal(payload, "MdErrorMessage"),
		RawResponse:  payloadToRaw(payload),
	}
}

func mapProvision(auth, prov map[string]string, order model.Order) model.PaymentResult {
	code := payloadVal(prov, "ResponseCode")
	ok := code == procSuccess
	res := model.PaymentResult{
		Success:     ok,
		OrderID:     firstNonEmpty(order.ID, payloadVal(auth, "OrderId")),
		RawResponse: mergeRaw(auth, prov),
	}
	if ok {
		res.AuthCode = payloadVal(prov, "AuthCode")
		res.HostRefNum = payloadVal(prov, "ReferenceCode")
		res.Amount = order.Amount
		res.Currency = order.Currency
	} else {
		res.ErrorCode = code
		res.ErrorMessage = payloadVal(prov, "ResponseDescription")
	}
	return res
}

func mapCancelRefund(raw map[string]string) model.PaymentResult {
	code := payloadVal(raw, "ResponseCode")
	ok := code == procSuccess
	res := model.PaymentResult{
		Success:     ok,
		RawResponse: payloadToRaw(raw),
	}
	if !ok {
		res.ErrorCode = code
		res.ErrorMessage = payloadVal(raw, "ResponseDescription")
	}
	return res
}

func mapStatus(raw map[string]string, order model.Order) model.PaymentResult {
	code := payloadVal(raw, "ResponseCode")
	ok := code == procStatusSuccess
	res := model.PaymentResult{
		Success:     ok,
		OrderID:     firstNonEmpty(order.ID, payloadVal(raw, "OrderId")),
		RawResponse: payloadToRaw(raw),
	}
	if !ok {
		res.ErrorCode = code
		res.ErrorMessage = payloadVal(raw, "ResponseDescription")
	}
	return res
}
