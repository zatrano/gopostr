package kuveyt

import "github.com/zatrano/gopostr/model"

const apiVersion = "TDV2.0.0"

// Endpoints Kuveyt Türk sanal POS URL'leri.
type Endpoints struct {
	PaymentAPI string
	Gateway3D  string
	QueryAPI   string
}

// DefaultTestEndpoints test ortamı.
var DefaultTestEndpoints = Endpoints{
	PaymentAPI: "https://boatest.kuveytturk.com.tr/boa.virtualpos.services/Home",
	Gateway3D:  "https://boatest.kuveytturk.com.tr/boa.virtualpos.services/Home/ThreeDModelPayGate",
	QueryAPI:   "https://boatest.kuveytturk.com.tr/BOA.Integration.WCFService/BOA.Integration.VirtualPos/VirtualPosService.svc/Basic",
}

// DefaultProdEndpoints canlı ortam.
var DefaultProdEndpoints = Endpoints{
	PaymentAPI: "https://sanalpos.kuveytturk.com.tr/ServiceGateWay/Home",
	Gateway3D:  "https://sanalpos.kuveytturk.com.tr/ServiceGateWay/Home/ThreeDModelPayGate",
	QueryAPI:   "https://sanalpos.kuveytturk.com.tr/BOA.Integration.WCFService/BOA.Integration.VirtualPos/VirtualPosService.svc/Basic",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
