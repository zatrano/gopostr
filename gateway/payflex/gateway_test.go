package payflex

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
		ClientID:   "merchant1",
		Password:   "pass1",
		TerminalID: "term1",
	}
}

func TestInit_enrollment(t *testing.T) {
	enrollXML, err := os.ReadFile(filepath.Join("testdata", "enrollment_success.xml"))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(enrollXML)
	}))
	defer srv.Close()

	g := New(Config{
		Credentials: testCreds(),
		Endpoints: Endpoints{
			Gateway3D:  srv.URL,
			PaymentAPI: srv.URL,
		},
	})

	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "ORD-1", Amount: 100, Currency: model.CurrencyTRY,
			SuccessURL: "https://ok", FailURL: "https://fail",
		},
		PaymentModel: model.PaymentModel3DSecure,
		TxType:       model.TxTypePayAuth,
		Card: &model.CardInput{
			Number: "4111111111111111", ExpireMonth: "05", ExpireYear: "25", CVV: "123",
			Brand: "visa",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["PaReq"] != "pareq-test" {
		t.Fatalf("inputs: %+v", form.Inputs)
	}
	if form.Gateway != "https://acs.test/3d" {
		t.Fatalf("gateway: %s", form.Gateway)
	}
}

func TestHandleCallback_provision(t *testing.T) {
	vposXML, err := os.ReadFile(filepath.Join("testdata", "vpos_success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(vposXML)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{PaymentAPI: srv.URL}})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"Status": "Y", "Eci": "05", "Cavv": "cavv", "VerifyEnrollmentRequestId": "ORD-1",
		"OrderId": "ORD-1", "PurchAmount": "100.00", "PurchCurrency": "949",
		"pan": "4111111111111111", "cvv": "123", "expiry": "202505", "cardHoldersName": "TEST",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.TransactionID != "TX-PF-1" {
		t.Fatalf("sonuç: %+v", res)
	}
}

func TestInit_rejects3DPay(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultVakifTestEndpoints})
	_, err := g.Init(context.Background(), model.InitRequest{
		Order:        model.Order{ID: "1", Amount: 1, SuccessURL: "a", FailURL: "b"},
		PaymentModel: model.PaymentModel3DPay,
		TxType:       model.TxTypePayAuth,
		Card:         &model.CardInput{Number: "4111111111111111", ExpireMonth: "01", ExpireYear: "30", CVV: "123"},
	})
	if err == nil {
		t.Fatal("3D Pay desteklenmemeli")
	}
}
