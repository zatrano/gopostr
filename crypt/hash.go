package crypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"hash"
)

// SHA256Hex veriyi SHA-256 ile hashleyip hex string döner.
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// SHA512Hex veriyi SHA-512 ile hashleyip hex string döner.
func SHA512Hex(data []byte) string {
	sum := sha512.Sum512(data)
	return hex.EncodeToString(sum[:])
}

// SHA256Base64 veriyi SHA-256 ile hashleyip base64 döner (Payten EstV3 formatı).
func SHA256Base64(data []byte) string {
	sum := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(sum[:])
}

// SHA512Base64 veriyi SHA-512 ile hashleyip base64 döner (Payten EstV3 varsayılan).
func SHA512Base64(data []byte) string {
	sum := sha512.Sum512(data)
	return base64.StdEncoding.EncodeToString(sum[:])
}

// HMACSHA256 HMAC-SHA256 hesaplar.
func HMACSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

// HMACSHA512 HMAC-SHA512 hesaplar.
func HMACSHA512(key, data []byte) []byte {
	mac := hmac.New(sha512.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

// HashBase64 verilen hash fonksiyonu ile binary hash üretip base64 kodlar.
func HashBase64(h func() hash.Hash, data string) string {
	hasher := h()
	_, _ = hasher.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
