package akbank

import "testing"

func TestCreate3DHash_deterministic(t *testing.T) {
	inputs := map[string]string{
		"paymentModel": "3D", "txnCode": "3000",
		"merchantSafeId": "M1", "terminalSafeId": "T1",
		"orderId": "O1", "lang": "TR", "amount": "10.00",
		"currencyCode": "949", "installCount": "1",
		"okUrl": "https://ok", "failUrl": "https://fail",
		"randomNumber": "ABC", "requestDateTime": "2024-01-01T00:00:00.000",
	}
	h1 := Create3DHash("secret", inputs)
	h2 := Create3DHash("secret", inputs)
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash: %q %q", h1, h2)
	}
}

func TestAuthHash(t *testing.T) {
	body := []byte(`{"txnCode":"1000"}`)
	h := AuthHash("key", body)
	if h == "" {
		t.Fatal("boş auth hash")
	}
}
