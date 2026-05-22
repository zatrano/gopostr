package akbank

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"strings"

	"github.com/zatrano/gopostr/crypt"
)

// Create3DHash 3D form hash değeri (HMAC-SHA512, base64).
func Create3DHash(storeKey string, inputs map[string]string) string {
	parts := []string{
		payloadVal(inputs, "paymentModel"),
		payloadVal(inputs, "txnCode"),
		payloadVal(inputs, "merchantSafeId"),
		payloadVal(inputs, "terminalSafeId"),
		payloadVal(inputs, "orderId"),
		payloadVal(inputs, "lang"),
		payloadVal(inputs, "amount"),
		payloadVal(inputs, "ccbRewardAmount"),
		payloadVal(inputs, "pcbRewardAmount"),
		payloadVal(inputs, "xcbRewardAmount"),
		payloadVal(inputs, "currencyCode"),
		payloadVal(inputs, "installCount"),
		payloadVal(inputs, "okUrl"),
		payloadVal(inputs, "failUrl"),
		payloadVal(inputs, "emailAddress"),
		payloadVal(inputs, "subMerchantId"),
		payloadVal(inputs, "creditCard"),
		payloadVal(inputs, "expiredDate"),
		payloadVal(inputs, "cvv"),
		payloadVal(inputs, "randomNumber"),
		payloadVal(inputs, "requestDateTime"),
		payloadVal(inputs, "b2bIdentityNumber"),
	}
	return hmacSHA512Base64(storeKey, strings.Join(parts, ""))
}

// AuthHash JSON API gövdesi için auth-hash header.
func AuthHash(storeKey string, body []byte) string {
	return hmacSHA512Base64(storeKey, string(body))
}

// Check3DHash callback hash doğrulaması.
func Check3DHash(storeKey string, payload map[string]string) bool {
	expected := payloadVal(payload, "hash")
	if expected == "" {
		return false
	}
	params := payloadVal(payload, "hashParams")
	if params == "" {
		return false
	}
	var parts []string
	for _, name := range strings.Split(params, "+") {
		parts = append(parts, payloadVal(payload, name))
	}
	actual := hmacSHA512Base64(storeKey, strings.Join(parts, "+"))
	return actual == expected
}

func hmacSHA512Base64(key, data string) string {
	mac := hmac.New(sha512.New, []byte(key))
	_, _ = mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// RandomNumber 128 karakterlik hex rastgele değer.
func RandomNumber() (string, error) {
	return crypt.RandomString(128)
}

