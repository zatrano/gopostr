package kuveyt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zatrano/gopostr/model"
)

func testCreds() model.BankCredentials {
	return model.BankCredentials{
		ClientID: "123456",
		Username: "apiuser",
		Password: "789012", // CustomerId
		StoreKey: "STOREKEY",
	}
}

func TestParseHTMLForm(t *testing.T) {
	html, _ := os.ReadFile(filepath.Join("testdata", "enrollment_form.html"))
	gw, inputs, err := parseHTMLForm(string(html))
	if err != nil {
		t.Fatal(err)
	}
	if gw != "https://acs.example/3d" || inputs["MD"] != "md-test-token" {
		t.Fatalf("gw=%s inputs=%v", gw, inputs)
	}
}

func TestInit_enrollment(t *testing.T) {
	formHTML, _ := os.ReadFile(filepath.Join("testdata", "enrollment_form.html"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(formHTML)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{Gateway3D: srv.URL}})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "ORD1", Amount: 100, Currency: model.CurrencyTRY,
			SuccessURL: "https://ok", FailURL: "https://fail",
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
	if form.Inputs["MD"] != "md-test-token" {
		t.Fatalf("form: %+v", form.Inputs)
	}
}

func TestHandleCallback_provision(t *testing.T) {
	authXML, _ := os.ReadFile(filepath.Join("testdata", "auth_success.xml"))
	provXML, _ := os.ReadFile(filepath.Join("testdata", "provision_success.xml"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "ThreeDModelProvisionGate") || strings.Contains(r.RequestURI, "Provision") {
			_, _ = w.Write(provXML)
			return
		}
		_, _ = w.Write(provXML)
	}))
	defer srv.Close()

	g := New(Config{
		Credentials: testCreds(),
		Endpoints:   Endpoints{PaymentAPI: srv.URL},
	})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"AuthenticationResponse": string(authXML),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.AuthCode != "AUTH-K1" {
		t.Fatalf("sonuç: %+v", res)
	}
}

func TestCancel_SOAP(t *testing.T) {
	soapXML, _ := os.ReadFile(filepath.Join("testdata", "cancel_success.xml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(soapXML)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{QueryAPI: srv.URL}})
	res, err := g.Cancel(context.Background(), model.CancelRequest{
		Order: model.Order{
			ID: "ORD-K1", Amount: 100, RefRetNum: "RRN", TransactionID: "STAN",
			RecurringID: "888",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("iptal: %+v", res)
	}
}
