package garanti

import (
	"testing"

	"github.com/zatrano/gopostr/model"
)

func testCreds() model.BankCredentials {
	return model.BankCredentials{
		ClientID:   "7000679",
		Username:   "PROVAUT",
		Password:   "123qweASD/",
		StoreKey:   "12345678",
		TerminalID: "30691297",
		TestMode:   true,
	}
}

func TestCreate3DHash_deterministic(t *testing.T) {
	inputs := map[string]string{
		"terminalid":          "30691297",
		"orderid":             "ORDER-1",
		"txnamount":           "1050",
		"txncurrencycode":     "949",
		"successurl":          "https://ok",
		"errorurl":            "https://fail",
		"txntype":             "sales",
		"txninstallmentcount": "",
	}
	h1 := Create3DHash(credsFromModel(testCreds()), inputs)
	h2 := Create3DHash(credsFromModel(testCreds()), inputs)
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash: %s vs %s", h1, h2)
	}
}

func TestIs3DAuthSuccess(t *testing.T) {
	if !is3DAuthSuccess("1") {
		t.Fatal("1 başarılı")
	}
	if is3DAuthSuccess("7") {
		t.Fatal("7 başarısız")
	}
}
