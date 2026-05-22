package payfor

import "github.com/zatrano/gopostr/model"

// Endpoints PayFor (QNB Finansbank / Enpara) URL'leri.
type Endpoints struct {
	PaymentAPI    string
	Gateway3D     string
	Gateway3DHost string
}

// DefaultFinansTestEndpoints QNB Finansbank test.
var DefaultFinansTestEndpoints = Endpoints{
	PaymentAPI:    "https://vpostest.qnbfinansbank.com/Gateway/XMLGate.aspx",
	Gateway3D:     "https://vpostest.qnbfinansbank.com/Gateway/Default.aspx",
	Gateway3DHost: "https://vpostest.qnbfinansbank.com/Gateway/3DHost.aspx",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
