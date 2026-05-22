package vakifkatilim

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
		ClientID: "100000", Username: "user", Password: "200000", StoreKey: "KEY",
	}
}

func TestInit_3DHost(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order:        model.Order{ID: "H1", Amount: 50, Currency: model.CurrencyTRY, SuccessURL: "https://ok", FailURL: "https://fail"},
		PaymentModel: model.PaymentModel3DHost,
		TxType:       model.TxTypePayAuth,
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["HashPassword"] == "" || form.Gateway == "" {
		t.Fatalf("form: %+v", form)
	}
}

func TestInit_3DSecure_enrollment(t *testing.T) {
	html, _ := os.ReadFile(filepath.Join("testdata", "enrollment_form.html"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(html)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{Gateway3D: srv.URL}})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{ID: "ORD1", Amount: 50, SuccessURL: "https://ok", FailURL: "https://fail"},
		PaymentModel: model.PaymentModel3DSecure, TxType: model.TxTypePayAuth,
		Card: &model.CardInput{Number: "4506347050031019", ExpireMonth: "12", ExpireYear: "26", CVV: "000"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["MD"] != "md-vk-1" {
		t.Fatalf("form: %+v", form.Inputs)
	}
}

func TestHandleCallback_3DSecure(t *testing.T) {
	prov, _ := os.ReadFile(filepath.Join("testdata", "provision_success.xml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "ThreeDModelProvisionGate") {
			_, _ = w.Write(prov)
		}
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{PaymentAPI: srv.URL}})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"ResponseCode": "00", "MD": "md-vk-1", "MerchantOrderId": "ORD-VK-1", "Amount": "5000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.AuthCode != "PROV-VK" {
		t.Fatalf("sonuç: %+v", res)
	}
}

func TestHandleCallback_3DHost(t *testing.T) {
	g := New(Config{Credentials: testCreds()})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"ResponseCode": "00", "PaymentType": "1", "MerchantOrderId": "H1",
		"ProvisionNumber": "P1", "Stan": "S1", "RRN": "R1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("sonuç: %+v", res)
	}
}
