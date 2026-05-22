package payflexcpv4

import (
	"crypto/sha1"
	"encoding/base64"
)

const hashVersion = "VBank3DPay2014"

func hashSHA1Base64(s string) string {
	sum := sha1.Sum([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// CreateEnrollmentHash kayıt isteği HashedData alanı.
func CreateEnrollmentHash(merchantID, amountCode, amount, merchantPassword string) string {
	parts := []string{merchantID, amountCode, amount, merchantPassword, "", hashVersion}
	return hashSHA1Base64(joinParts(parts))
}

func joinParts(parts []string) string {
	var s string
	for _, p := range parts {
		s += p
	}
	return s
}
