package posnet

import "github.com/zatrano/gopostr/model"

// Endpoints Yapı Kredi PosNet URL'leri.
type Endpoints struct {
	PaymentAPI string
	Gateway3D  string
}

// DefaultTestEndpoints test ortamı.
var DefaultTestEndpoints = Endpoints{
	PaymentAPI: "https://setmpos.ykb.com/PosnetWebService/XML",
	Gateway3D:  "https://setmpos.ykb.com/3DSWebService/YKBPaymentService",
}

// DefaultProdEndpoints canlı ortam.
var DefaultProdEndpoints = Endpoints{
	PaymentAPI: "https://posnet.yapikredi.com.tr/PosnetWebService/XML",
	Gateway3D:  "https://posnet.yapikredi.com.tr/3DSWebService/YKBPaymentService",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
