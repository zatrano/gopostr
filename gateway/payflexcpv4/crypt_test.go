package payflexcpv4

import "testing"

func TestCreateEnrollmentHash_deterministic(t *testing.T) {
	h := CreateEnrollmentHash("M1", "949", "10.50", "pass")
	h2 := CreateEnrollmentHash("M1", "949", "10.50", "pass")
	if h == "" || h != h2 {
		t.Fatalf("hash: %q", h)
	}
}
