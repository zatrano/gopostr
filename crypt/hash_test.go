package crypt

import "testing"

func TestSHA512Base64(t *testing.T) {
	got := SHA512Base64([]byte("test"))
	if got == "" {
		t.Fatal("boş hash")
	}
}

func TestRandomString(t *testing.T) {
	s, err := RandomString(24)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 24 {
		t.Fatalf("uzunluk: %d", len(s))
	}
}
