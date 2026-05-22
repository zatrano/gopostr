package posnet

import (
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

func is3DAuthSuccess(mdStatus string) bool {
	switch mdStatus {
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

func orderFromPayload(p map[string]string) model.Order {
	currency := firstNonEmpty(
		payloadVal(p, "currency"),
		payloadVal(p, "Currency"),
		parseCurrency(payloadVal(p, "currencyCode")),
	)
	if currency == "" {
		currency = model.CurrencyTRY
	}
	amount := parseAmountFromPayload(p)
	return model.Order{
		ID:       firstNonEmpty(payloadVal(p, "orderId"), payloadVal(p, "OrderId"), payloadVal(p, "oid"), payloadVal(p, "XID")),
		Amount:   amount,
		Currency: currency,
	}
}

func parseAmountFromPayload(p map[string]string) float64 {
	for _, k := range []string{"amount", "Amount", "PurchAmount", "purchAmount"} {
		if v := payloadVal(p, k); v != "" {
			if strings.Contains(v, ".") {
				return parseAmountKurusFromDecimal(v)
			}
			return parseAmountKurus(v)
		}
	}
	return 0
}

func parseAmountKurusFromDecimal(s string) float64 {
	s = strings.ReplaceAll(s, ",", ".")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func mapAPIResult(raw map[string]string, order model.Order) model.PaymentResult {
	proc := payloadVal(raw, "approved")
	errCode := payloadVal(raw, "respCode")
	success := proc == procSuccess && errCode == ""

	res := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(order.ID, payloadVal(raw, "orderID"), payloadVal(raw, "xid")),
		TransactionID: payloadVal(raw, "hostlogkey"),
		AuthCode:      payloadVal(raw, "authCode"),
		HostRefNum:    payloadVal(raw, "hostlogkey"),
		Amount:        order.Amount,
		Currency:      order.Currency,
		ErrorCode:     errCode,
		ErrorMessage:  payloadVal(raw, "respText"),
		RawResponse:   raw,
	}
	if order.Amount == 0 && payloadVal(raw, "amount") != "" {
		res.Amount = parseAmountKurus(payloadVal(raw, "amount"))
	}
	if res.Currency == "" && payloadVal(raw, "currency") != "" {
		res.Currency = parseCurrency(payloadVal(raw, "currency"))
	}
	return res
}

func map3DFailed(payload, raw map[string]string, order model.Order) model.PaymentResult {
	return model.PaymentResult{
		Success:      false,
		OrderID:      order.ID,
		Amount:       order.Amount,
		Currency:     order.Currency,
		ErrorCode:    firstNonEmpty(payloadVal(raw, "respCode"), payloadVal(payload, "respCode")),
		ErrorMessage: firstNonEmpty(payloadVal(raw, "respText"), payloadVal(payload, "respText")),
		RawResponse:  mergeRaw(payload, raw),
	}
}

func map3DResult(resolveRaw, payRaw map[string]string, order model.Order) model.PaymentResult {
	md := payloadVal(resolveRaw, "mdStatus")
	if !is3DAuthSuccess(md) {
		return map3DFailed(nil, resolveRaw, order)
	}
	if payRaw == nil {
		return map3DFailed(nil, resolveRaw, order)
	}
	res := mapAPIResult(payRaw, order)
	if res.OrderID == "" {
		res.OrderID = firstNonEmpty(payloadVal(resolveRaw, "xid"), order.ID)
	}
	res.RawResponse = mergeRaw(resolveRaw, payRaw)
	return res
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
