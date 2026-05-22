package vakifkatilim

import "github.com/zatrano/gopostr/model"

const apiVersion = "1.0.0"

// Endpoints Vakıf Katılım sanal POS URL'leri.
type Endpoints struct {
	PaymentAPI    string
	Gateway3D     string
	Gateway3DHost string
}

// DefaultTestEndpoints test ortamı URL'leri.
var DefaultTestEndpoints = Endpoints{
	PaymentAPI:    "https://boa.vakifkatilim.com.tr/VirtualPOS.Gateway/Home",
	Gateway3D:     "https://boa.vakifkatilim.com.tr/VirtualPOS.Gateway/Home/ThreeDModelPayGate",
	Gateway3DHost: "https://boa.vakifkatilim.com.tr/VirtualPOS.Gateway/CommonPaymentPage/CommonPaymentPage",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
