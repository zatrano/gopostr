package payflexcpv4

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
		Password:   "pass",
		TerminalID: "term1",
	}
}

func TestInit_3DSecure(t *testing.T) {
	body, _ := os.ReadFile(filepath.Join("testdata", "register_ok.xml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{RegisterAPI: srv.URL}})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "ORD1", Amount: 10.5, SuccessURL: "https://ok", FailURL: "https://fail",
		},
		PaymentModel: model.PaymentModel3DSecure,
		TxType:       model.TxTypePayAuth,
		Card: &model.CardInput{
			Number: "4506347050031019", ExpireMonth: "12", ExpireYear: "26", CVV: "000", Brand: "visa",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["Ptkn"] != "TOKEN123" || form.Method != http.MethodGet {
		t.Fatalf("form: %+v", form)
	}
}

func TestInit_3DHost(t *testing.T) {
	body, _ := os.ReadFile(filepath.Join("testdata", "register_ok.xml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{RegisterAPI: srv.URL}})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "ORD2", Amount: 99, SuccessURL: "https://ok", FailURL: "https://fail",
		},
		PaymentModel: model.PaymentModel3DHost,
		TxType:       model.TxTypePayAuth,
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Gateway == "" {
		t.Fatal("empty gateway")
	}
}

func TestHandleCallback(t *testing.T) {
	vpos, _ := os.ReadFile(filepath.Join("testdata", "vpos_ok.xml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(vpos)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{VposAPI: srv.URL}})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"TransactionId": "TX1", "PaymentToken": "TOKEN123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.AuthCode != "AUTH1" {
		t.Fatalf("result: %+v", res)
	}
}
