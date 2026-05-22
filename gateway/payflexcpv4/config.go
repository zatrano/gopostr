package payflexcpv4

import "github.com/zatrano/gopostr/model"

// Endpoints Vakıfbank PayFlex Common Payment v4.
type Endpoints struct {
	RegisterAPI string // RegisterTransaction
	VposAPI     string // VposTransaction (işlem doğrulama)
}

// DefaultTestEndpoints test ortamı.
var DefaultTestEndpoints = Endpoints{
	RegisterAPI: "https://cptest.vakifbank.com.tr/CommonPayment/api/RegisterTransaction",
	VposAPI:     "https://cptest.vakifbank.com.tr/CommonPayment/api/VposTransaction",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials model.BankCredentials
	Endpoints   Endpoints
}
