package crypt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const defaultNonceLen = 24

// Nonce kriptografik olarak güvenli rastgele hex string üretir.
func Nonce(byteLen int) (string, error) {
	if byteLen <= 0 {
		byteLen = defaultNonceLen / 2
	}
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypt: nonce: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// RandomString 0-9A-F rastgele string (varsayılan 24 karakter).
func RandomString(length int) (string, error) {
	if length <= 0 {
		length = defaultNonceLen
	}
	const chars = "0123456789ABCDEF"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypt: random string: %w", err)
	}
	out := make([]byte, length)
	for i := range out {
		out[i] = chars[int(b[i])%len(chars)]
	}
	return string(out), nil
}
