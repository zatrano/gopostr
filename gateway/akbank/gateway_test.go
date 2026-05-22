package akbank

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/zatrano/gopostr/model"
)

func testCreds() model.BankCredentials {
	return model.BankCredentials{
		ClientID: "merchant-safe-id",
		Username: "terminal-safe-id",
		StoreKey:   "secret-key",
		Lang:     model.LangTR,
	}
}

func TestInit_3DSecure(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "AK-1", Amount: 25.50, Currency: model.CurrencyTRY,
			SuccessURL: "https://ok", FailURL: "https://fail", IP: "127.0.0.1",
		},
		PaymentModel: model.PaymentModel3DSecure,
		TxType:       model.TxTypePayAuth,
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["hash"] == "" {
		t.Fatal("hash eksik")
	}
	if form.Inputs["paymentModel"] != "3D" {
		t.Fatalf("model: %s", form.Inputs["paymentModel"])
	}
	if form.Method != http.MethodPost {
		t.Fatalf("method: %s", form.Method)
	}
}

func TestInit_3DHost_URL(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "AK-H", Amount: 1, Currency: model.CurrencyTRY,
			SuccessURL: "https://ok", FailURL: "https://fail", IP: "1.1.1.1",
		},
		PaymentModel: model.PaymentModel3DHost,
		TxType:       model.TxTypePayAuth,
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Gateway != DefaultTestEndpoints.Gateway3DHost {
		t.Fatalf("url: %s", form.Gateway)
	}
}

func TestHandleCallback_3DPay(t *testing.T) {
	g := New(Config{Credentials: testCreds(), SkipHashCheck: true})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"paymentModel": "3D_PAY", "orderId": "P1", "responseCode": procSuccess,
		"authCode": "A1", "rrn": "R1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("sonuç: %+v", res)
	}
}

func TestTestdata_paymentSuccess(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "payment_success.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 {
		t.Fatal("fixture boş")
	}
}

func TestStatus_notSupported(t *testing.T) {
	g := New(Config{Credentials: testCreds()})
	_, err := g.Status(context.Background(), model.StatusRequest{})
	if err == nil {
		t.Fatal("hata bekleniyordu")
	}
}
