package posnetv1

import "github.com/zatrano/gopostr/model"

const apiVersion = "V100"

// Endpoints Albaraka PosNet JSON API.
type Endpoints struct {
	PaymentAPI string
	Gateway3D  string
}

// DefaultTestEndpoints test ortamı.
var DefaultTestEndpoints = Endpoints{
	PaymentAPI: "https://epostest.albarakaturk.com.tr/ALBMerchantService/MerchantJSONAPI.svc",
	Gateway3D:  "https://epostest.albarakaturk.com.tr/ALBSecurePaymentUI/SecureProcess/SecureVerification.aspx",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
