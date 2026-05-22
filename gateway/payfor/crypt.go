package payfor

import (
	"crypto/sha1"
	"encoding/base64"
	"strings"

	"github.com/zatrano/gopostr/model"
)

// Create3DHash 3D form Hash (SHA-1, base64).
func Create3DHash(storeKey string, inputs map[string]string) string {
	parts := []string{
		payloadVal(inputs, "MbrId"),
		payloadVal(inputs, "OrderId"),
		payloadVal(inputs, "PurchAmount"),
		payloadVal(inputs, "OkUrl"),
		payloadVal(inputs, "FailUrl"),
		payloadVal(inputs, "TxnType"),
		payloadVal(inputs, "InstallmentCount"),
		payloadVal(inputs, "Rnd"),
		storeKey,
	}
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// Check3DHash callback ResponseHash doğrulaması.
func Check3DHash(creds modelCredentials, payload map[string]string) bool {
	parts := []string{
		creds.MerchantID,
		creds.StoreKey,
		payloadVal(payload, "OrderId"),
		payloadVal(payload, "AuthCode"),
		payloadVal(payload, "ProcReturnCode"),
		payloadVal(payload, "3DStatus"),
		payloadVal(payload, "ResponseRnd"),
		creds.UserCode,
	}
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	expected := base64.StdEncoding.EncodeToString(sum[:])
	return payloadVal(payload, "ResponseHash") == expected
}

type modelCredentials struct {
	MerchantID string
	StoreKey   string
	UserCode   string
}

func credsFromModel(c model.BankCredentials) modelCredentials {
	return modelCredentials{
		MerchantID: c.ClientID,
		StoreKey:   c.StoreKey,
		UserCode:   c.Username,
	}
}

