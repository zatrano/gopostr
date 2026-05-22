package garanti

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/zatrano/gopostr/model"
)

func TestInit_3DSecure(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "G-1", Amount: 10.50, Currency: model.CurrencyTRY,
			SuccessURL: "https://ok", FailURL: "https://fail", IP: "127.0.0.1",
		},
		PaymentModel: model.PaymentModel3DSecure,
		TxType:       model.TxTypePayAuth,
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["secure3dhash"] == "" {
		t.Fatal("hash eksik")
	}
	if form.Inputs["secure3dsecuritylevel"] != "3D" {
		t.Fatalf("level: %s", form.Inputs["secure3dsecuritylevel"])
	}
	if form.Method != http.MethodPost {
		t.Fatalf("method: %s", form.Method)
	}
}

func TestHandleCallback_3DPay(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints, SkipHashCheck: true})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"orderid": "G-PAY", "txnamount": "2500", "txncurrencycode": "949",
		"secure3dsecuritylevel": "3D_PAY", "mdstatus": "1", "procreturncode": "00",
		"response": "Approved", "authcode": "A1", "transid": "T1", "hostrefnum": "H1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("sonuç: %+v", res)
	}
}

func TestHandleCallback_3DSecure_MockAPI(t *testing.T) {
	xmlBody, err := os.ReadFile(filepath.Join("testdata", "provision_success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(xmlBody)
	}))
	defer srv.Close()

	g := New(Config{
		Credentials: testCreds(),
		Endpoints:   Endpoints{PaymentAPI: srv.URL, Gateway3D: DefaultTestEndpoints.Gateway3D},
		SkipHashCheck: true,
	})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"orderid": "ORDER-G1", "txnamount": "1050", "txncurrencycode": "949",
		"txntype": "sales", "mdstatus": "1", "response": "Approved",
		"secure3dsecuritylevel": "3D", "cavv": "cavv", "eci": "05", "xid": "x", "md": "md",
		"customeripaddress": "127.0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("provision: %+v", res)
	}
	if res.AuthCode != "AUTH1" {
		t.Fatalf("auth: %s", res.AuthCode)
	}
}
