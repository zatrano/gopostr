package akbank

import (
	"encoding/json"
	"fmt"
	"time"
)

const apiVersion = "1.00"

type terminalBlock struct {
	MerchantSafeID string `json:"merchantSafeId"`
	TerminalSafeID string `json:"terminalSafeId"`
}

type orderBlock struct {
	OrderID string `json:"orderId,omitempty"`
}

type transactionBlock struct {
	Amount       string `json:"amount,omitempty"`
	CurrencyCode int    `json:"currencyCode,omitempty"`
	MotoInd      int    `json:"motoInd,omitempty"`
	InstallCount int    `json:"installCount,omitempty"`
}

type secureTxBlock struct {
	SecureID      string `json:"secureId"`
	SecureEcomInd string `json:"secureEcomInd"`
	SecureData    string `json:"secureData"`
	SecureMd      string `json:"secureMd"`
}

type customerBlock struct {
	IPAddress string `json:"ipAddress"`
}

type apiRequest struct {
	Version           string           `json:"version,omitempty"`
	TxnCode           string           `json:"txnCode"`
	RequestDateTime   string           `json:"requestDateTime"`
	RandomNumber      string           `json:"randomNumber"`
	Terminal          terminalBlock    `json:"terminal"`
	SubMerchant       *subMerchantBlock `json:"subMerchant,omitempty"`
	Order             orderBlock       `json:"order"`
	Transaction       *transactionBlock `json:"transaction,omitempty"`
	SecureTransaction *secureTxBlock   `json:"secureTransaction,omitempty"`
	Customer          *customerBlock   `json:"customer,omitempty"`
}

type subMerchantBlock struct {
	SubMerchantID string `json:"subMerchantId"`
}

type apiResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	HostMessage     string `json:"hostMessage"`
	TxnDateTime     string `json:"txnDateTime"`
	Order           struct {
		OrderID string `json:"orderId"`
	} `json:"order"`
	Transaction struct {
		AuthCode     string `json:"authCode"`
		Rrn          string `json:"rrn"`
		BatchNumber  string `json:"batchNumber"`
	} `json:"transaction"`
}

func encodeAPI(req apiRequest) ([]byte, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("akbank: encode json: %w", err)
	}
	return b, nil
}

func decodeAPI(body []byte) (apiResponse, map[string]string, error) {
	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp, nil, fmt.Errorf("akbank: decode json: %w", err)
	}
	raw := jsonToMap(body)
	return resp, raw, nil
}

func jsonToMap(body []byte) map[string]string {
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return nil
	}
	return flattenMap("", m)
}

func flattenMap(prefix string, m map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]interface{}:
			for sk, sv := range flattenMap(key, val) {
				out[sk] = sv
			}
		case string:
			out[key] = val
		case float64:
			out[key] = fmt.Sprintf("%v", val)
		default:
			if val != nil {
				out[key] = fmt.Sprintf("%v", val)
			}
		}
	}
	return out
}

func requestDateTime() string {
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		loc = time.FixedZone("TR", 3*3600)
	}
	return time.Now().In(loc).Format("2006-01-02T15:04:05") + ".000"
}
