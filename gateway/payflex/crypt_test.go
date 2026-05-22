package payflex

import "testing"

func TestNewEnrollmentRequestID(t *testing.T) {
	id, err := newEnrollmentRequestID()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 24 {
		t.Fatalf("uzunluk: %d", len(id))
	}
	id2, _ := newEnrollmentRequestID()
	if id == id2 {
		t.Fatal("rastgele ID tekrar etmemeli")
	}
}
