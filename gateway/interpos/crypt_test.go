package interpos

import "testing"

func TestCreate3DHash_deterministic(t *testing.T) {
	inputs := map[string]string{
		"ShopCode": "SHOP1", "OrderId": "ORD1", "PurchAmount": "10.00",
		"OkUrl": "https://ok", "FailUrl": "https://fail",
		"TxnType": "Auth", "InstallmentCount": "0", "Rnd": "abc",
	}
	h1 := Create3DHash("store", inputs)
	h2 := Create3DHash("store", inputs)
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash: %q", h1)
	}
}
