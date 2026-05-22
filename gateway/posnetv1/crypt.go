package posnetv1

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

func hashSHA256Base64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// Create3DFormHash 3D form Mac alanı (SHA-256, base64).
func Create3DFormHash(storeKey string, inputs map[string]string) string {
	parts := []string{
		payloadVal(inputs, "MerchantNo"),
		payloadVal(inputs, "TerminalNo"),
		payloadVal(inputs, "CardNo"),
		payloadVal(inputs, "Cvv"),
		payloadVal(inputs, "ExpiredDate"),
		payloadVal(inputs, "Amount"),
		storeKey,
	}
	return hashSHA256Base64(strings.Join(parts, ""))
}

// MACFromFieldList bankanın MACParams alan listesine göre MAC üretir (değerler bitişik + storeKey).
func MACFromFieldList(storeKey, macFieldList string, fields map[string]string) string {
	var parts []string
	for _, name := range strings.Split(macFieldList, ":") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		parts = append(parts, macField(fields, name))
	}
	return hashSHA256Base64(strings.Join(parts, "") + storeKey)
}

func macField(fields map[string]string, key string) string {
	if v := payloadVal(fields, key); v != "" {
		return v
	}
	// ExpireDate form MacParams'ta; banka ExpiredDate alanı kullanır.
	if key == "ExpireDate" {
		return payloadVal(fields, "ExpiredDate")
	}
	return ""
}

// CreateProvisionMAC Sale/Auth provizyon MAC.
func CreateProvisionMAC(storeKey string, merchantNo, terminalNo string, tds map[string]string) string {
	parts := []string{
		merchantNo,
		terminalNo,
		payloadVal(tds, "SecureTransactionId"),
		payloadVal(tds, "CavvData"),
		payloadVal(tds, "Eci"),
		payloadVal(tds, "MdStatus"),
		storeKey,
	}
	return hashSHA256Base64(strings.Join(parts, ""))
}

// Check3DCallbackHash callback Mac doğrulaması.
func Check3DCallbackHash(storeKey string, payload map[string]string) bool {
	macFieldList := payloadVal(payload, "MacParams")
	if macFieldList == "" {
		macFieldList = payloadVal(payload, "MACParams")
	}
	if macFieldList == "" {
		return false
	}
	expected := payloadVal(payload, "Mac")
	if expected == "" {
		expected = payloadVal(payload, "MAC")
	}
	actual := MACFromFieldList(storeKey, macFieldList, payload)
	return actual == expected
}

func flattenForMAC(data map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range data {
		switch x := v.(type) {
		case map[string]interface{}:
			for sk, sv := range x {
				out[sk] = fmt.Sprint(sv)
			}
		case nil:
			continue
		default:
			out[k] = fmt.Sprint(v)
		}
	}
	return out
}
