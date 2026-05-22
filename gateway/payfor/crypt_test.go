package payfor

import (
	"crypto/sha1"
	"encoding/base64"
	"strings"
	"testing"
)

func TestCreate3DHash_deterministic(t *testing.T) {
	inputs := map[string]string{
		"MbrId": "5", "OrderId": "O1", "PurchAmount": "10.00",
		"OkUrl": "https://ok", "FailUrl": "https://fail",
		"TxnType": "Auth", "InstallmentCount": "0", "Rnd": "ABC",
	}
	h1 := Create3DHash("key", inputs)
	h2 := Create3DHash("key", inputs)
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash: %q %q", h1, h2)
	}
}

func TestCheck3DHash(t *testing.T) {
	creds := modelCredentials{MerchantID: "M1", StoreKey: "KEY", UserCode: "U1"}
	payload := map[string]string{
		"OrderId": "O1", "AuthCode": "A1", "ProcReturnCode": "00",
		"3DStatus": "1", "ResponseRnd": "RND",
	}
	var parts []string
	for _, p := range []string{creds.MerchantID, creds.StoreKey, payload["OrderId"], payload["AuthCode"],
		payload["ProcReturnCode"], payload["3DStatus"], payload["ResponseRnd"], creds.UserCode} {
		parts = append(parts, p)
	}
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	payload["ResponseHash"] = base64.StdEncoding.EncodeToString(sum[:])
	if !Check3DHash(creds, payload) {
		t.Fatal("hash doğrulaması başarısız")
	}
}
