package posnetv1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/zatrano/gopostr/model"
)

func testCreds() model.BankCredentials {
	return model.BankCredentials{
		ClientID:   "6700950031",
		Username:   "6700950031",
		PosNetID:   "6700950031",
		TerminalID: "6700950032",
		StoreKey:   "store-key",
	}
}

func TestInit_3DSecure(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "123", Amount: 10.5, Currency: model.CurrencyTRY,
			SuccessURL: "https://merchant/ok",
		},
		PaymentModel: model.PaymentModel3DSecure,
		TxType:       model.TxTypePayAuth,
		Card: &model.CardInput{
			Number: "4506347050031019", ExpireMonth: "12", ExpireYear: "26", CVV: "000",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["Mac"] == "" || form.Method != http.MethodPost || form.Inputs["OrderId"] != "00000000000000000123" {
		t.Fatalf("form: %+v", form.Inputs)
	}
}

func TestInit_rejects3DPay(t *testing.T) {
	g := New(Config{Credentials: testCreds()})
	_, err := g.Init(context.Background(), model.InitRequest{
		Order:        model.Order{ID: "1", Amount: 1, SuccessURL: "https://ok"},
		PaymentModel: model.PaymentModel3DPay,
		TxType:       model.TxTypePayAuth,
		Card:         &model.CardInput{Number: "4111", ExpireMonth: "12", ExpireYear: "26", CVV: "000"},
	})
	if err == nil {
		t.Fatal("expected error for 3d_pay")
	}
}

func TestHandleCallback_provision(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "sale_ok.json"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	g := New(Config{
		Credentials: testCreds(),
		Endpoints:   Endpoints{PaymentAPI: srv.URL},
		SkipHashCheck: true,
	})
	payload := map[string]string{
		"MdStatus": "1", "OrderId": "00000000000000000123", "Amount": "1050",
		"SecureTransactionId": "stx1", "CAVV": "cavv", "ECI": "05", "MD": "md1",
	}
	res, err := g.HandleCallback(context.Background(), payload)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.AuthCode != "AUTH1" {
		t.Fatalf("result: %+v", res)
	}
}
