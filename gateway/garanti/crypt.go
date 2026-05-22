package garanti

import (
	"crypto/sha1"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/zatrano/gopostr/model"
)

// Create3DHash 3D form secure3dhash değerini üretir (SHA-512 hex upper).
func Create3DHash(creds modelCredentials, inputs map[string]string) string {
	txType := payloadVal(inputs, "txntype")
	sec := createSecurityData(creds, payloadVal(inputs, "terminalid"), txType)
	parts := []string{
		payloadVal(inputs, "terminalid"),
		payloadVal(inputs, "orderid"),
		payloadVal(inputs, "txnamount"),
		payloadVal(inputs, "txncurrencycode"),
		payloadVal(inputs, "successurl"),
		payloadVal(inputs, "errorurl"),
		txType,
		payloadVal(inputs, "txninstallmentcount"),
		creds.StoreKey,
		sec,
	}
	return hashUpperSHA512(strings.Join(parts, ""))
}

// Check3DHash callback hash doğrulaması (hashparams alanına göre).
func Check3DHash(storeKey string, payload map[string]string) bool {
	expected := strings.ToUpper(payloadVal(payload, "hash"))
	if expected == "" {
		return false
	}
	paramsKey := payloadVal(payload, "hashparams")
	if paramsKey == "" {
		return false
	}
	names := strings.Split(paramsKey, ":")
	var parts []string
	for _, name := range names {
		parts = append(parts, payloadVal(payload, name))
	}
	val := strings.Join(parts, "") + storeKey
	actual := hashUpperSHA512(val)
	return actual == expected
}

// CreateAPIHash GVPS API Terminal.HashData üretir.
func CreateAPIHash(creds modelCredentials, orderID, terminalID, cardNumber, amount, currency, txType string) string {
	sec := createSecurityData(creds, terminalID, txType)
	parts := []string{orderID, terminalID, cardNumber, amount, currency, sec}
	return hashUpperSHA512(strings.Join(parts, ""))
}

type modelCredentials struct {
	Password       string
	RefundPassword string
	StoreKey       string
}

func credsFromModel(c model.BankCredentials) modelCredentials {
	rp := c.RefundPassword
	if rp == "" {
		rp = c.Password
	}
	return modelCredentials{
		Password:       c.Password,
		RefundPassword: rp,
		StoreKey:       c.StoreKey,
	}
}

func createSecurityData(c modelCredentials, terminalID, txType string) string {
	pw := c.Password
	if txType == "void" || txType == "refund" {
		pw = c.RefundPassword
	}
	padded := fmt.Sprintf("%09s", terminalID)
	return hashUpperSHA1(pw + padded)
}

func hashUpperSHA512(s string) string {
	sum := sha512.Sum512([]byte(s))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func hashUpperSHA1(s string) string {
	sum := sha1.Sum([]byte(s))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}
