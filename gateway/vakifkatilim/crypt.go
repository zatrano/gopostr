package vakifkatilim

import (
	"crypto/sha1"
	"encoding/base64"
)

func sha1Base64(s string) string {
	sum := sha1.Sum([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// HashPassword StoreKey SHA-1 Base64 (HashPassword alanı).
func HashPassword(storeKey string) string {
	return sha1Base64(storeKey)
}

// CreateHash VPos HashData (Kuveyt ile aynı algoritma).
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
