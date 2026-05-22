package interpos

import "github.com/zatrano/gopostr/model"

// Endpoints Denizbank InterPos URL'leri.
type Endpoints struct {
	PaymentAPI    string
	Gateway3D     string
	Gateway3DHost string
}

// DefaultTestEndpoints test ortamı.
var DefaultTestEndpoints = Endpoints{
	PaymentAPI:    "https://test.inter-vpos.com.tr/mpi/Default.aspx",
	Gateway3D:     "https://test.inter-vpos.com.tr/mpi/Default.aspx",
	Gateway3DHost: "https://test.inter-vpos.com.tr/mpi/3DHost.aspx",
}

// DefaultProdEndpoints canlı ortam.
var DefaultProdEndpoints = Endpoints{
	PaymentAPI:    "https://inter-vpos.com.tr/mpi/Default.aspx",
	Gateway3D:     "https://inter-vpos.com.tr/mpi/Default.aspx",
	Gateway3DHost: "https://inter-vpos.com.tr/mpi/3DHost.aspx",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
