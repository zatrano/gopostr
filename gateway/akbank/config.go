package akbank

import "github.com/zatrano/gopostr/model"

// Endpoints Akbank sanal POS URL'leri.
type Endpoints struct {
	PaymentAPI  string
	Gateway3D   string
	Gateway3DHost string
}

// DefaultTestEndpoints prep ortamı.
var DefaultTestEndpoints = Endpoints{
	PaymentAPI:    "https://apipre.akbank.com/api/v1/payment/virtualpos",
	Gateway3D:     "https://virtualpospaymentgatewaypre.akbank.com/securepay",
	Gateway3DHost: "https://virtualpospaymentgatewaypre.akbank.com/payhosting",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
