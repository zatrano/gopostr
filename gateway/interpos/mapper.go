package interpos

import (
	"strconv"

	"github.com/zatrano/gopostr/model"
)

func is3DAuthSuccess(status string) bool {
	switch status {
	case "1", "2", "3", "4":
		return true
	default:
		return false
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

func mapAPIResult(raw map[string]string, order model.Order) model.PaymentResult {
	proc := payloadVal(raw, "ProcReturnCode")
	success := proc == procSuccess
	res := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(payloadVal(raw, "OrderId"), order.ID),
		TransactionID: payloadVal(raw, "TransId"),
		AuthCode:      payloadVal(raw, "AuthCode"),
		Amount:        order.Amount,
		Currency:      order.Currency,
		ErrorCode:     payloadVal(raw, "ErrorCode"),
		ErrorMessage:  payloadVal(raw, "ErrorMessage"),
		RawResponse:   raw,
	}
	if !success && res.ErrorCode == "" {
		res.ErrorCode = proc
	}
	return res
}

func map3DFailed(payload map[string]string, order model.Order) model.PaymentResult {
	return model.PaymentResult{
		Success:      false,
		OrderID:      firstNonEmpty(payloadVal(payload, "OrderId"), order.ID),
		Amount:       firstAmount(payload, order),
		Currency:     firstNonEmpty(parseCurrency(payloadVal(payload, "Currency")), order.Currency),
		ErrorCode:    firstNonEmpty(payloadVal(payload, "ErrorCode"), payloadVal(payload, "ProcReturnCode")),
		ErrorMessage: payloadVal(payload, "ErrorMessage"),
		RawResponse:  payloadToRaw(payload),
	}
}

func map3DPayHostResult(payload map[string]string, order model.Order) model.PaymentResult {
	status := payloadVal(payload, "3DStatus")
	proc := payloadVal(payload, "ProcReturnCode")
	success := proc == procSuccess && is3DAuthSuccess(status)
	res := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(payloadVal(payload, "OrderId"), order.ID),
		TransactionID: payloadVal(payload, "TransId"),
		AuthCode:      payloadVal(payload, "AuthCode"),
		Amount:        firstAmount(payload, order),
		Currency:      firstNonEmpty(parseCurrency(payloadVal(payload, "Currency")), order.Currency),
		ErrorCode:     payloadVal(payload, "ErrorCode"),
		ErrorMessage:  payloadVal(payload, "ErrorMessage"),
		RawResponse:   payloadToRaw(payload),
	}
	if !success && res.ErrorCode == "" {
		res.ErrorCode = proc
	}
	return res
}

func map3DResult(authRaw, payRaw map[string]string, order model.Order) model.PaymentResult {
	if !is3DAuthSuccess(payloadVal(authRaw, "3DStatus")) {
		return map3DFailed(authRaw, order)
	}
	if payRaw == nil {
		return map3DFailed(authRaw, order)
	}
	res := mapAPIResult(payRaw, order)
	res.RawResponse = mergeRaw(authRaw, payRaw)
	return res
}

func orderFromPayload(p map[string]string) model.Order {
	inst, _ := strconv.Atoi(payloadVal(p, "InstallmentCount"))
	return model.Order{
		ID:          firstNonEmpty(payloadVal(p, "OrderId"), payloadVal(p, "oid")),
		Amount:      firstAmount(p, model.Order{}),
		Currency:    parseCurrency(payloadVal(p, "Currency")),
		Installment: inst,
	}
}

func firstAmount(p map[string]string, order model.Order) float64 {
	if order.Amount > 0 {
		return order.Amount
	}
	if v := payloadVal(p, "PurchAmount"); v != "" {
		return parsePurchAmount(v)
	}
	return 0
}

func paymentModelFromPayload(p map[string]string) string {
	if m, ok := secureTypeToModel[payloadVal(p, "SecureType")]; ok {
		return m
	}
	return model.PaymentModel3DSecure
}

func accountFields(c model.BankCredentials) map[string]string {
	return map[string]string{
		"ShopCode": c.ClientID,
		"UserCode": c.Username,
		"UserPass": c.Password,
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func mergeRaw(parts ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, p := range parts {
		for k, v := range p {
			out[k] = v
		}
	}
	return out
}
