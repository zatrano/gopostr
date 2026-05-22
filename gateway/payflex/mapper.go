package payflex

import (
	"strings"

	"github.com/zatrano/gopostr/model"
)

var txTypeToPayflex = map[string]string{
	model.TxTypePayAuth:       "Sale",
	model.TxTypePayPreAuth:    "Auth",
	model.TxTypeCancel:        "Cancel",
	model.TxTypeRefund:        "Refund",
	model.TxTypeRefundPartial: "Refund",
}

func is3DAuthSuccess(status string) bool {
	return status == "Y"
}

func mapVposResult(resp vposResponse, raw map[string]string, order model.Order) model.PaymentResult {
	success := resp.ResultCode == procSuccess
	result := model.PaymentResult{
		Success:       success,
		OrderID:       firstNonEmpty(resp.OrderId, order.ID),
		TransactionID: resp.TransactionId,
		AuthCode:      resp.AuthCode,
		HostRefNum:    firstNonEmpty(resp.Rrn, resp.TransactionId),
		Amount:        order.Amount,
		Currency:      order.Currency,
		ErrorCode:     resp.ResultCode,
		ErrorMessage:  resp.ResultDetail,
		RawResponse:  raw,
	}
	if success {
		result.ErrorCode = ""
		result.ErrorMessage = ""
	}
	return result
}

func mapSearchResult(resp searchResponse, raw map[string]string, order model.Order) model.PaymentResult {
	code := resp.ResponseInfo.ResponseCode
	success := code == procSuccess
	result := model.PaymentResult{
		Success:      success,
		OrderID:      order.ID,
		ErrorCode:    code,
		ErrorMessage: resp.ResponseInfo.ResponseMessage,
		RawResponse:  raw,
	}
	if !success {
		return result
	}
	tx := resp.TransactionSearchResultInfo.TransactionSearchResultInfo
	txSuccess := tx.ResultCode == procSuccess
	result.Success = txSuccess
	result.OrderID = firstNonEmpty(tx.OrderId, order.ID)
	result.TransactionID = tx.TransactionId
	result.AuthCode = tx.AuthCode
	result.HostRefNum = tx.Rrn
	if !txSuccess {
		result.ErrorCode = tx.ResultCode
		result.ErrorMessage = tx.ResponseMessage
	} else {
		result.ErrorCode = ""
		result.ErrorMessage = ""
	}
	return result
}

func map3DFailed(payload map[string]string) model.PaymentResult {
	return model.PaymentResult{
		Success:      false,
		OrderID:      payloadVal(payload, "VerifyEnrollmentRequestId"),
		ErrorCode:    payloadVal(payload, "ErrorCode"),
		ErrorMessage: payloadVal(payload, "ErrorMessage"),
		RawResponse:  payloadToRaw(payload),
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

func payloadVal(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}
	kl := strings.ToLower(key)
	for k, v := range m {
		if strings.ToLower(k) == kl {
			return v
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
