package estpos

import "github.com/zatrano/gopostr/model"

// Endpoints EstPOS (Payten/Asseco) API URL'leri.
type Endpoints struct {
	PaymentAPI string
	Gateway3D  string
}

// DefaultTestEndpoints Asseco entegrasyon test ortamı.
var DefaultTestEndpoints = Endpoints{
	PaymentAPI: "https://entegrasyon.asseco-see.com.tr/fim/api",
	Gateway3D:  "https://entegrasyon.asseco-see.com.tr/fim/est3Dgate",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
