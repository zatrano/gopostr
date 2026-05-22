package garanti

import "github.com/zatrano/gopostr/model"

// Endpoints Garanti BBVA API URL'leri.
type Endpoints struct {
	PaymentAPI string
	Gateway3D  string
}

// DefaultTestEndpoints test ortamı.
var DefaultTestEndpoints = Endpoints{
	PaymentAPI: "https://sanalposprovtest.garantibbva.com.tr/VPServlet",
	Gateway3D:  "https://sanalposprovtest.garantibbva.com.tr/servlet/gt3dengine",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
