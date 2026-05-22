package posnet

import "testing"

func TestCreateTransactionMAC_deterministic(t *testing.T) {
	m1 := CreateTransactionMAC("key", "tid", "0000000000000000000042", 1050, "TL", "mid")
	m2 := CreateTransactionMAC("key", "tid", "0000000000000000000042", 1050, "TL", "mid")
	if m1 != m2 || m1 == "" {
		t.Fatalf("mac: %q", m1)
	}
}
