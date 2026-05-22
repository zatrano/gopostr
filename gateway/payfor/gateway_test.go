package payfor

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
		ClientID: "merchant1",
		Username: "user1",
		Password: "pass1",
		StoreKey: "storekey",
		MbrID:    "5",
		Lang:     model.LangTR,
	}
}

func TestInit_3DSecure(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultFinansTestEndpoints})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "PF-1", Amount: 50, Currency: model.CurrencyTRY,
			SuccessURL: "https://ok", FailURL: "https://fail",
		},
		PaymentModel: model.PaymentModel3DSecure,
		TxType:       model.TxTypePayAuth,
		Card: &model.CardInput{
			Number: "4111111111111111", ExpireMonth: "05", ExpireYear: "25", CVV: "123",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["Hash"] == "" {
		t.Fatal("hash eksik")
	}
	if form.Inputs["SecureType"] != "3DModel" {
		t.Fatalf("secure: %s", form.Inputs["SecureType"])
	}
	if form.Method != http.MethodPost {
		t.Fatalf("method: %s", form.Method)
	}
}

func TestHandleCallback_3DPay(t *testing.T) {
	g := New(Config{Credentials: testCreds()})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"SecureType": "3DPay", "OrderId": "P1", "ProcReturnCode": procSuccess,
		"AuthCode": "A1", "HostRefNum": "H1", "PurchAmount": "50.00", "Currency": "949",
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
		_, _ = w.Write(xmlBody)
	}))
	defer srv.Close()

	g := New(Config{
		Credentials: testCreds(),
		Endpoints:   Endpoints{PaymentAPI: srv.URL},
		SkipHashCheck: true,
	})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"SecureType": "3DModel", "3DStatus": "1", "OrderId": "PF-1",
		"RequestGuid": "guid-1", "ProcReturnCode": "00",
		"AuthCode": "AUTH1", "HostRefNum": "HOST1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("provision: %+v", res)
	}
}
