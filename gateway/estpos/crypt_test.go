package estpos

import "testing"

func TestCreate3DHash_ver3(t *testing.T) {
	storeKey := "TEST_STORE_KEY"
	inputs := map[string]string{
		"amount":    "10.50",
		"clientid":  "100200127",
		"currency":  "949",
		"failUrl":   "https://merchant.test/fail",
		"hash":      "ignored",
		"islemtipi": "Auth",
		"lang":      "tr",
		"oid":       "ORDER-1",
		"okUrl":     "https://merchant.test/ok",
		"rnd":       "abc123",
		"storetype": "3d",
		"taksit":    "",
	}
	h1 := Create3DHash(storeKey, inputs)
	h2 := Create3DHash(storeKey, inputs)
	if h1 == "" {
		t.Fatal("hash boş")
	}
	if h1 != h2 {
		t.Fatalf("hash deterministik değil: %s vs %s", h1, h2)
	}
}

func TestCheck3DHash_excludesHashField(t *testing.T) {
	storeKey := "KEY"
	payload := map[string]string{
		"amount": "1.00",
		"HASH":   "",
	}
	payload["HASH"] = Create3DHash(storeKey, payload)
	if !Check3DHash(storeKey, payload) {
		t.Fatal("hash doğrulaması başarısız")
	}
}
