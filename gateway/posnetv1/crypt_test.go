package posnetv1

import "testing"

func TestCreate3DFormHash_deterministic(t *testing.T) {
	inputs := map[string]string{
		"MerchantNo": "M1", "TerminalNo": "T1", "CardNo": "4111",
		"Cvv": "123", "ExpiredDate": "2612", "Amount": "1000",
	}
	h1 := Create3DFormHash("key", inputs)
	h2 := Create3DFormHash("key", inputs)
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash: %q", h1)
	}
}

func TestMACFromFieldList(t *testing.T) {
	fields := map[string]string{
		"MerchantNo": "M1", "TerminalNo": "T1", "OrderId": "ORD",
	}
	mac := MACFromFieldList("secret", "MerchantNo:TerminalNo:OrderId", fields)
	if mac == "" {
		t.Fatal("empty mac")
	}
}
