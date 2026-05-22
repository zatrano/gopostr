package vakifkatilim

import (
	"strconv"

	"github.com/zatrano/gopostr/model"
)

func isAuthSuccess(code string) bool {
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
	if amt := payloadVal(raw, "Amount"); amt != "" {
		if n, err := strconv.Atoi(amt); err == nil {
			res.Amount = parseAmountKurus(n)
		}
	}
	if success {
		res.ErrorCode = ""
	}
	return res
}

func mapAuthFailed(raw map[string]string, order model.Order) model.PaymentResult {
	return model.PaymentResult{
		Success:      false,
		OrderID:      firstNonEmpty(payloadVal(raw, "MerchantOrderId"), order.ID),
		ErrorCode:    payloadVal(raw, "ResponseCode"),
		ErrorMessage: payloadVal(raw, "ResponseMessage"),
		RawResponse:  raw,
	}
}

func map3DResult(authRaw, payRaw map[string]string, order model.Order) model.PaymentResult {
	if !isAuthSuccess(payloadVal(authRaw, "ResponseCode")) {
		return mapAuthFailed(authRaw, order)
	}
	if payRaw == nil {
		return mapAuthFailed(authRaw, order)
	}
	res := mapPaymentResult(payRaw, order)
	res.RawResponse = mergeRaw(authRaw, payRaw)
	return res
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
		Currency: parseCurrency(payloadVal(p, "FECCurrencyCode")),
	}
}
