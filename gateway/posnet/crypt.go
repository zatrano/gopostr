package posnet

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/zatrano/gopostr/model"
)

const hashSep = ";"

// CreateTransactionMAC oosTranData / resolve MAC üretir.
func CreateTransactionMAC(storeKey, terminalID, orderID string, amountKurus int, currency, mid string) string {
	sec := createSecurityData(storeKey, terminalID)
	parts := []string{orderID, fmt.Sprintf("%d", amountKurus), currency, mid, sec}
	return hashSHA256Base64(strings.Join(parts, hashSep))
}

// CheckResolveMAC oosResolveMerchantDataResponse mac doğrulaması.
func CheckResolveMAC(creds modelCredentials, data map[string]string) bool {
	sec := createSecurityData(creds.StoreKey, creds.TerminalID)
	parts := []string{
		payloadVal(data, "mdStatus"),
		payloadVal(data, "xid"),
		payloadVal(data, "amount"),
		payloadVal(data, "currency"),
		creds.MerchantID,
		sec,
	}
	expected := hashSHA256Base64(strings.Join(parts, hashSep))
	return payloadVal(data, "mac") == expected
}

func createSecurityData(storeKey, terminalID string) string {
	return hashSHA256Base64(storeKey + hashSep + terminalID)
}

func hashSHA256Base64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

type modelCredentials struct {
	MerchantID string
	TerminalID string
	StoreKey   string
}

func credsFromModel(c model.BankCredentials) modelCredentials {
	return modelCredentials{
		MerchantID: c.ClientID,
		TerminalID: c.Password,
		StoreKey:   c.StoreKey,
	}
}

func posNetID(c model.BankCredentials) string {
	if c.PosNetID != "" {
		return c.PosNetID
	}
	return c.Username
}

