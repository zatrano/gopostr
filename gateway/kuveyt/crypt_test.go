package kuveyt

import "testing"

func TestCreateHash_deterministic(t *testing.T) {
	in := map[string]string{
		"MerchantId": "1", "MerchantOrderId": "ORD1", "Amount": "100",
		"OkUrl": "https://ok", "FailUrl": "https://fail", "UserName": "user",
	}
	h1 := CreateHash("secret", in)
	h2 := CreateHash("secret", in)
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash: %q", h1)
	}
}
