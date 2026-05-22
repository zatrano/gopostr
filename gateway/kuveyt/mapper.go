package kuveyt

import (
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

func is3DAuthSuccess(code string) bool {
	return code == procSuccess
}

func mapPaymentResult(raw map[string]string, order model.Order) model.PaymentResult {
	proc := payloadVal(raw, "ResponseCode")
	success := proc == procSuccess
	res := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(payloadVal(raw, "MerchantOrderId"), order.ID),
		TransactionID: payloadVal(raw, "Stan"),
		AuthCode:      payloadVal(raw, "ProvisionNumber"),
		HostRefNum:    payloadVal(raw, "RRN"),
		Amount:        order.Amount,
		Currency:      order.Currency,
		ErrorCode:     proc,
		ErrorMessage:  payloadVal(raw, "ResponseMessage"),
		RawResponse:   raw,
	}
	if success {
		if amt := payloadVal(raw, "Amount"); amt != "" {
			if n, err := strconv.Atoi(amt); err == nil {
				res.Amount = parseAmountKurus(n)
			}
		}
		if cc := payloadVal(raw, "CurrencyCode"); cc != "" {
			res.Currency = parseCurrency(cc)
		}
		res.ErrorCode = ""
	}
	return res
}

func map3DAuthFailed(raw map[string]string, order model.Order) model.PaymentResult {
	return model.PaymentResult{
		Success:      false,
		OrderID:      firstNonEmpty(payloadVal(raw, "MerchantOrderId"), order.ID),
		ErrorCode:    payloadVal(raw, "ResponseCode"),
		ErrorMessage: payloadVal(raw, "ResponseMessage"),
		RawResponse:  raw,
	}
}

func map3DResult(authRaw, payRaw map[string]string, order model.Order) model.PaymentResult {
	if !is3DAuthSuccess(payloadVal(authRaw, "ResponseCode")) {
		return map3DAuthFailed(authRaw, order)
	}
	if payRaw == nil {
		return map3DAuthFailed(authRaw, order)
	}
	res := mapPaymentResult(payRaw, order)
	res.RawResponse = mergeRaw(authRaw, payRaw)
	return res
}

func mapSOAPCancelRefund(raw map[string]string, orderID string) model.PaymentResult {
	proc := findNestedSuffix(raw, "Value.ResponseCode")
	if proc == "" {
		proc = findNested(raw, "ResponseCode")
	}
	success := proc == procSuccess
	res := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(findNested(raw, "MerchantOrderId"), orderID),
		TransactionID: findNested(raw, "Stan"),
		AuthCode:      findNested(raw, "ProvisionNumber"),
		HostRefNum:    findNested(raw, "RRN"),
		ErrorCode:     proc,
		ErrorMessage:  findNested(raw, "ResponseMessage"),
		RawResponse:   raw,
	}
	if success {
		res.ErrorCode = ""
	}
	return res
}

func mapSOAPStatus(raw map[string]string, orderID string) model.PaymentResult {
	proc := findNested(raw, "ResponseCode")
	if proc == "" {
		proc = findNested(raw, "ErrorCode")
	}
	success := proc == procSuccess
	res := model.PaymentResult{
		Success:      success,
		OrderID:      firstNonEmpty(findNested(raw, "MerchantOrderId"), orderID),
		ErrorCode:    proc,
		ErrorMessage: findNested(raw, "ErrorMessage"),
		RawResponse:  raw,
	}
	if success {
		res.AuthCode = findNested(raw, "ProvNumber")
		res.HostRefNum = findNested(raw, "RRN")
		res.TransactionID = findNested(raw, "Stan")
		res.ErrorCode = ""
	}
	return res
}

func findNestedSuffix(m map[string]string, suffix string) string {
	for k, v := range m {
		if strings.HasSuffix(k, suffix) {
			return v
		}
	}
	return ""
}

func findNested(m map[string]string, leaf string) string {
	if v := payloadVal(m, leaf); v != "" {
		return v
	}
	suffix := "." + leaf
	for k, v := range m {
		if strings.HasSuffix(k, suffix) || k == leaf {
			return v
		}
	}
	return ""
}

func orderFromPayload(p map[string]string) model.Order {
	amt := 0
	if v := payloadVal(p, "Amount"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			amt = n
		}
	}
	return model.Order{
		ID:       payloadVal(p, "MerchantOrderId"),
		Amount:   parseAmountKurus(amt),
		Currency: parseCurrency(payloadVal(p, "CurrencyCode")),
		IP:       payloadVal(p, "CustomerIPAddress"),
	}
}
