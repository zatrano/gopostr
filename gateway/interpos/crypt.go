package interpos

import (
	"crypto/sha1"
	"encoding/base64"
	"strings"
)

func sha1Base64(s string) string {
	sum := sha1.Sum([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// Create3DHash 3D form Hash (SHA-1, base64, alanlar bitişik).
func Create3DHash(storeKey string, inputs map[string]string) string {
	parts := []string{
		payloadVal(inputs, "ShopCode"),
		payloadVal(inputs, "OrderId"),
		payloadVal(inputs, "PurchAmount"),
		payloadVal(inputs, "OkUrl"),
		payloadVal(inputs, "FailUrl"),
		payloadVal(inputs, "TxnType"),
		payloadVal(inputs, "InstallmentCount"),
		payloadVal(inputs, "Rnd"),
		storeKey,
	}
	return sha1Base64(strings.Join(parts, ""))
}

// Check3DHash callback HASH / HASHPARAMS doğrulaması.
func Check3DHash(storeKey string, payload map[string]string) bool {
	hashParams := payloadVal(payload, "HASHPARAMS")
	if hashParams == "" {
		return false
	}
	names := strings.Split(hashParams, ":")
	vals := make([]string, len(names))
	for i, n := range names {
		vals[i] = payloadVal(payload, strings.TrimSpace(n))
	}
	expected := sha1Base64(strings.Join(vals, ":") + storeKey)
	return payloadVal(payload, "HASH") == expected
}
