package estpos

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
		ClientID:  "100200127",
		Username:  "testuser",
		Password:  "testpass",
		StoreKey:  "TEST_STORE_KEY",
		Lang:      model.LangTR,
	}
}

func TestInit_3DSecure(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "E-1", Amount: 10.50, Currency: model.CurrencyTRY,
			SuccessURL: "https://ok", FailURL: "https://fail",
		},
		PaymentModel: model.PaymentModel3DSecure,
		TxType:       model.TxTypePayAuth,
		Card: &model.CardInput{
			Number: "4111111111111111", ExpireMonth: "12", ExpireYear: "30", CVV: "123",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["hash"] == "" {
		t.Fatal("hash eksik")
	}
	if form.Inputs["storetype"] != "3d" {
		t.Fatalf("storetype: %s", form.Inputs["storetype"])
	}
	if form.Method != http.MethodPost {
		t.Fatalf("method: %s", form.Method)
	}
}

func TestHandleCallback_3DPay(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints, SkipHashCheck: true})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"storetype":      "3d_pay",
		"ProcReturnCode": procSuccess,
		"oid":            "E-2",
		"amount":         "25.00",
		"currency":       "949",
		"AuthCode":       "AUTH1",
		"HostRefNum":     "HOST1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("success=false: %+v", res)
	}
	if res.OrderID != "E-2" {
		t.Fatalf("order: %s", res.OrderID)
	}
}

func TestHandleCallback_3DSecure_provision(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "cc5_success.xml"))
	if err != nil {
		t.Skip("testdata yok:", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	g := New(Config{
		Credentials:   testCreds(),
		Endpoints:     Endpoints{PaymentAPI: srv.URL, Gateway3D: DefaultTestEndpoints.Gateway3D},
		SkipHashCheck: true,
	})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"storetype": "3d",
		"mdStatus":  "1",
		"oid":       "E-3",
		"amount":    "10.00",
		"currency":  "949",
		"md":        "mdtoken",
		"xid":       "xid1",
		"eci":       "05",
		"cavv":      "cavv1",
		"islemtipi": "Auth",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("success=false: %+v", res)
	}
	if res.TransactionID != "TX123" {
		t.Fatalf("trans id: %s", res.TransactionID)
	}
}

func TestDecodeResponse_fixture(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "cc5_success.xml"))
	if err != nil {
		t.Skip(err)
	}
	resp, raw, err := decodeResponse(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ProcReturnCode != procSuccess {
		t.Fatalf("proc: %s", resp.ProcReturnCode)
	}
	if resp.TransID != "TX123" {
		t.Fatalf("trans: %s, raw: %+v", resp.TransID, raw)
	}
}
