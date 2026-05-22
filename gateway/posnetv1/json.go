package posnetv1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

type serviceResponseData struct {
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
}

type apiResponse struct {
	ServiceResponseData serviceResponseData `json:"ServiceResponseData"`
	AuthCode            string              `json:"AuthCode"`
	ReferenceCode       string              `json:"ReferenceCode"`
}

func encodeJSON(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("posnetv1: json encode: %w", err)
	}
	return b, nil
}

func decodeAPI(body []byte) (apiResponse, map[string]string, error) {
	var resp apiResponse
	if len(body) == 0 {
		return resp, nil, nil
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp, nil, fmt.Errorf("posnetv1: json decode: %w", err)
	}
	raw := map[string]string{
		"ResponseCode":        resp.ServiceResponseData.ResponseCode,
		"ResponseDescription": resp.ServiceResponseData.ResponseDescription,
		"AuthCode":            resp.AuthCode,
		"ReferenceCode":       resp.ReferenceCode,
	}
	return resp, raw, nil
}

func buildProvisionRequest(creds merchantFields, order model.Order, txType string, tds map[string]string) (map[string]interface{}, error) {
	inst := mapInstallment(order.Installment)
	req := map[string]interface{}{
		"ApiType":               "JSON",
		"ApiVersion":            apiVersion,
		"MerchantNo":            creds.MerchantNo,
		"TerminalNo":            creds.TerminalNo,
		"PaymentInstrumentType": "CARD",
		"IsEncrypted":           "N",
		"IsTDSecureMerchant":    "Y",
		"IsMailOrder":           "N",
		"ThreeDSecureData": map[string]interface{}{
			"SecureTransactionId": payloadVal(tds, "SecureTransactionId"),
			"CavvData":            firstNonEmpty(payloadVal(tds, "CavvData"), payloadVal(tds, "CAVV")),
			"Eci":                 firstNonEmpty(payloadVal(tds, "Eci"), payloadVal(tds, "ECI")),
			"MdStatus":            atoiDefault(payloadVal(tds, "MdStatus"), 0),
			"MD":                  firstNonEmpty(payloadVal(tds, "MD"), payloadVal(tds, "Md")),
		},
		"MACParams":        "MerchantNo:TerminalNo:SecureTransactionId:CavvData:Eci:MdStatus",
		"Amount":           formatAmountKurus(order.Amount),
		"CurrencyCode":     mapCurrency(order.Currency),
		"PointAmount":      0,
		"OrderId":          creds.OrderID,
		"InstallmentCount": inst,
		"InstallmentType":  "N",
	}
	if order.Installment > 1 {
		req["InstallmentType"] = "Y"
	}
	fields := flattenForMAC(req)
	req["MAC"] = MACFromFieldList(creds.StoreKey, "MerchantNo:TerminalNo:SecureTransactionId:CavvData:Eci:MdStatus", fields)
	return req, nil
}

func buildCancelRefundRequest(creds merchantFields, order model.Order, txAPI string) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"ApiType":         "JSON",
		"ApiVersion":      apiVersion,
		"MerchantNo":      creds.MerchantNo,
		"TerminalNo":      creds.TerminalNo,
		"MACParams":       "MerchantNo:TerminalNo:ReferenceCode:OrderId",
		"CipheredData":    nil,
		"DealerData":      nil,
		"IsEncrypted":     nil,
		"ReferenceCode":   nil,
		"OrderId":         nil,
		"TransactionType": txAPI,
	}
	if order.RefRetNum != "" {
		req["ReferenceCode"] = order.RefRetNum
	} else {
		oid, err := prefixedOrderID(order.ID, model.PaymentModel3DSecure)
		if err != nil {
			return nil, err
		}
		req["OrderId"] = oid
	}
	fields := flattenForMAC(req)
	req["MAC"] = MACFromFieldList(creds.StoreKey, "MerchantNo:TerminalNo:ReferenceCode:OrderId", fields)
	return req, nil
}

func buildStatusRequest(creds merchantFields, orderID string) (map[string]interface{}, error) {
	oid, err := prefixedOrderID(orderID, model.PaymentModel3DSecure)
	if err != nil {
		return nil, err
	}
	req := map[string]interface{}{
		"ApiType":     "JSON",
		"ApiVersion":  apiVersion,
		"MerchantNo":  creds.MerchantNo,
		"TerminalNo":  creds.TerminalNo,
		"MACParams":   "MerchantNo:TerminalNo",
		"IsEncrypted": "N",
		"OrderId":     oid,
	}
	fields := flattenForMAC(req)
	req["MAC"] = MACFromFieldList(creds.StoreKey, "MerchantNo:TerminalNo", fields)
	return req, nil
}

type merchantFields struct {
	MerchantNo, TerminalNo, StoreKey, OrderID string
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}
