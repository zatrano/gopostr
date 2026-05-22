package posnet

import (
	"context"
	"io"
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
		ClientID: "6700972650",
		Username: "12345",
		Password: "678901",
		StoreKey: "TEST_STORE_KEY",
		PosNetID: "12345",
	}
}

func TestFormatOrderID(t *testing.T) {
	got, err := formatOrderID("42")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Repeat("0", 18) + "42"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestInit_enrollment(t *testing.T) {
	xmlBody, err := os.ReadFile(filepath.Join("testdata", "enrollment_success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(xmlBody)
	}))
	defer srv.Close()

	g := New(Config{
		Credentials: testCreds(),
		Endpoints:   Endpoints{PaymentAPI: srv.URL, Gateway3D: "https://3d.example/"},
	})
	form, err := g.Init(context.Background(), model.InitRequest{
		Order: model.Order{
			ID: "42", Amount: 10.50, Currency: model.CurrencyTRY,
			SuccessURL: "https://ok", FailURL: "https://fail",
		},
		PaymentModel: model.PaymentModel3DSecure,
		TxType:       model.TxTypePayAuth,
		Card: &model.CardInput{
			Number: "4506347050031019", ExpireMonth: "12", ExpireYear: "26", CVV: "000",
			HolderName: "TEST USER",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if form.Inputs["posnetData"] != "ENC_DATA1" || form.Inputs["digest"] != "DIGEST_SIGN" {
		t.Fatalf("form: %+v", form.Inputs)
	}
}

func TestHandleCallback_3DSecure(t *testing.T) {
	resolveXML, _ := os.ReadFile(filepath.Join("testdata", "resolve_success.xml"))
	tranXML, _ := os.ReadFile(filepath.Join("testdata", "tran_success.xml"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body := string(b)
		switch {
		case strings.Contains(body, "oosResolveMerchantData"):
			_, _ = w.Write(resolveXML)
		case strings.Contains(body, "oosTranData"):
			_, _ = w.Write(tranXML)
		default:
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	g := New(Config{
		Credentials:   testCreds(),
		Endpoints:     Endpoints{PaymentAPI: srv.URL},
		SkipHashCheck: true,
	})
	res, err := g.HandleCallback(context.Background(), map[string]string{
		"BankPacket": "bank-pkt", "MerchantPacket": "merch-pkt", "Sign": "sig",
		"orderId": "42", "amount": "10.50", "currency": model.CurrencyTRY,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.AuthCode != "AUTH99" || res.HostRefNum != "HOSTLOG99" {
		t.Fatalf("sonuç: %+v", res)
	}
}

func TestCancel(t *testing.T) {
	xmlBody, _ := os.ReadFile(filepath.Join("testdata", "cancel_success.xml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(xmlBody)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{PaymentAPI: srv.URL}})
	res, err := g.Cancel(context.Background(), model.CancelRequest{
		Order: model.Order{ID: "42", RefRetNum: "HOSTREF"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("iptal: %+v", res)
	}
}

func TestStatus(t *testing.T) {
	xmlBody, _ := os.ReadFile(filepath.Join("testdata", "status_success.xml"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(xmlBody)
	}))
	defer srv.Close()

	g := New(Config{Credentials: testCreds(), Endpoints: Endpoints{PaymentAPI: srv.URL}})
	res, err := g.Status(context.Background(), model.StatusRequest{
		Order: model.Order{ID: "1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("durum: %+v", res)
	}
}

func TestInit_rejects3DPay(t *testing.T) {
	g := New(Config{Credentials: testCreds(), Endpoints: DefaultTestEndpoints})
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
