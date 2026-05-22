package interpos

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
		ClientID: "SHOP1",
		Username: "user",
		Password: "pass",
		StoreKey: "STOREKEY",
	}
}

func TestDecodeResponse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "provision_success.txt"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := decodeResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if raw["ProcReturnCode"] != procSuccess || raw["TransId"] != "TX-IP-1" {
		t.Fatalf("raw: %+v", raw)
	}
}

func TestInit_3DSecure(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "ORD1", Amount: 99.9, Currency: model.CurrencyTRY,
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
	if form.Inputs["Hash"] == "" || form.Inputs["SecureType"] != "3DModel" {
		t.Fatalf("form: %+v", form.Inputs)
	}
}

func TestHandleCallback_3DSecure(t *testing.T) {
	provBody, _ := os.ReadFile(filepath.Join("testdata", "provision_success.txt"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(provBody)
	}))
	defer srv.Close()

	g := New(Config{
		Credentials:   testCreds(),
		Endpoints:     Endpoints{PaymentAPI: srv.URL},
		SkipHashCheck: true,
	})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"3DStatus": "1", "OrderId": "ORD-IP-1", "PurchAmount": "99.9", "Currency": "949",
		"MD": "md-token", "PayerTxnId": "ptx", "Eci": "05", "PayerAuthenticationCode": "cavv",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.TransactionID != "TX-IP-1" {
		t.Fatalf("sonuç: %+v", res)
	}
}

func TestHandleCallback_3DPay(t *testing.T) {
	g := New(Config{Credentials: testCreds(), SkipHashCheck: true})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"3DStatus": "1", "SecureType": "3DPay", "OrderId": "P1",
		"ProcReturnCode": procSuccess, "AuthCode": "A1", "TransId": "T1",
		"PurchAmount": "50", "Currency": "949",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("sonuç: %+v", res)
	}
}

func TestCancel(t *testing.T) {
	body := []byte("ProcReturnCode=00;;OrderId=ORD1;;TransId=TX-C;;AuthCode=;;ErrorCode=;;ErrorMessage=")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{PaymentAPI: srv.URL}})
	res, err := g.Cancel(context.Background(), model.CancelRequest{Order: model.Order{ID: "ORD1"}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("iptal: %+v", res)
	}
}
