package garanti

import (
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

const procSuccess = "00"

var modelToSecure = map[string]string{
	model.PaymentModel3DSecure: "3D",
	model.PaymentModel3DPay:    "3D_PAY",
}

var secureToModel = map[string]string{
	"3D":     model.PaymentModel3DSecure,
	"3D_PAY": model.PaymentModel3DPay,
}

var txTypeToGaranti = map[string]string{
	model.TxTypePayAuth:       "sales",
	model.TxTypePayPreAuth:    "preauth",
	model.TxTypePayPostAuth:   "postauth",
	model.TxTypeCancel:        "void",
	model.TxTypeRefund:        "refund",
	model.TxTypeRefundPartial: "refund",
	model.TxTypeStatus:        "orderinq",
}

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

func map3DCommon(payload map[string]string) model.PaymentResult {
	md := payloadVal(payload, "mdstatus")
	proc := payloadVal(payload, "procreturncode")
	status := is3DAuthSuccess(md) && !strings.EqualFold(payloadVal(payload, "response"), "Error")

	result := model.PaymentResult{
		Success:      status,
		OrderID:      firstNonEmpty(payloadVal(payload, "orderid"), payloadVal(payload, "oid")),
		Amount:       parseAmountKurus(payloadVal(payload, "txnamount")),
		Currency:     parseCurrency(payloadVal(payload, "txncurrencycode")),
		ErrorCode:    "",
		ErrorMessage: payloadVal(payload, "errmsg"),
		RawResponse:  payloadToRaw(payload),
	}
	if strings.EqualFold(payloadVal(payload, "response"), "Error") {
		result.Success = false
		result.ErrorCode = proc
	}
	if !status && result.ErrorMessage == "" {
		result.ErrorMessage = payloadVal(payload, "mderrormessage")
	}
	return result
}

func map3DPayResult(payload map[string]string) model.PaymentResult {
	common := map3DCommon(payload)
	proc := payloadVal(payload, "procreturncode")
	if common.Success && proc == procSuccess {
		common.Success = true
		common.AuthCode = payloadVal(payload, "authcode")
		common.TransactionID = payloadVal(payload, "transid")
		common.HostRefNum = payloadVal(payload, "hostrefnum")
	} else if common.Success {
		common.Success = proc == procSuccess
		if !common.Success {
			common.ErrorCode = proc
		}
	}
	return common
}

func mapAPIResult(resp gvpsResponse, raw map[string]string, order model.Order) model.PaymentResult {
	code := resp.Transaction.Response.Code
	success := code == procSuccess
	result := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(resp.Order.OrderID, order.ID),
		AuthCode:      resp.Transaction.AuthCode,
		HostRefNum:    resp.Transaction.RetrefNum,
		Amount:        order.Amount,
		Currency:      order.Currency,
		ErrorCode:     resp.Transaction.Response.ReasonCode,
		ErrorMessage:  resp.Transaction.Response.ErrorMsg,
		RawResponse:   raw,
	}
	if result.ErrorMessage == "" {
		result.ErrorMessage = resp.Transaction.Response.Message
	}
	if success {
		result.ErrorCode = ""
		result.ErrorMessage = ""
	} else if result.ErrorCode == "" {
		result.ErrorCode = code
	}
	return result
}

func orderFromPayload(payload map[string]string) model.Order {
	inst := 0
	if s := payloadVal(payload, "txninstallmentcount"); s != "" {
		inst, _ = parseInt(s)
	}
	return model.Order{
		ID:          firstNonEmpty(payloadVal(payload, "orderid"), payloadVal(payload, "oid")),
		Amount:      parseAmountKurus(payloadVal(payload, "txnamount")),
		Currency:    parseCurrency(payloadVal(payload, "txncurrencycode")),
		Installment: inst,
		IP:          payloadVal(payload, "customeripaddress"),
	}
}

func paymentModelFromPayload(payload map[string]string) string {
	st := payloadVal(payload, "secure3dsecuritylevel")
	if m, ok := secureToModel[st]; ok {
		return m
	}
	return model.PaymentModel3DSecure
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
