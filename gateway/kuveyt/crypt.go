package kuveyt

import (
	"crypto/sha1"
	"encoding/base64"
)

// CreateHash KuveytTurk HashData (SHA-1 Base64).
// hashInput: MerchantId, MerchantOrderId, Amount, OkUrl, FailUrl, UserName alanlarını içerir.
func CreateHash(storeKey string, hashInput map[string]string) string {
	hashedPass := sha1Base64(storeKey)
	parts := []string{
		payloadVal(hashInput, "MerchantId"),
		payloadVal(hashInput, "MerchantOrderId"),
		payloadVal(hashInput, "Amount"),
		payloadVal(hashInput, "OkUrl"),
		payloadVal(hashInput, "FailUrl"),
		payloadVal(hashInput, "UserName"),
		hashedPass,
	}
	var b []byte
	for _, p := range parts {
		b = append(b, p...)
	}
	return sha1Base64(string(b))
}

func sha1Base64(s string) string {
	sum := sha1.Sum([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}
