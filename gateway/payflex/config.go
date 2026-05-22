package payflex

import "github.com/zatrano/gopostr/model"

// Endpoints PayFlex VPOS 7/24 (Vakıfbank / Ziraat).
type Endpoints struct {
	PaymentAPI string
	Gateway3D  string
	QueryAPI   string
}

// DefaultVakifTestEndpoints Vakıfbank test.
var DefaultVakifTestEndpoints = Endpoints{
	PaymentAPI: "https://onlineodemetest.vakifbank.com.tr:4443/VposService/v3/Vposreq.aspx",
	Gateway3D:  "https://3dsecuretest.vakifbank.com.tr:4443/MPIAPI/MPI_Enrollment.aspx",
	QueryAPI:   "https://onlineodemetest.vakifbank.com.tr:4443/UIService/Search.aspx",
}

// DefaultZiraatTestEndpoints Ziraat test.
var DefaultZiraatTestEndpoints = Endpoints{
	PaymentAPI: "https://preprod.payflex.com.tr/Ziraatbank/VposWeb/v3/Vposreq.aspx",
	Gateway3D:  "https://preprod.payflex.com.tr/ZiraatBank/MpiWeb/MPI_Enrollment.aspx",
	QueryAPI:   "https://sanalpos.ziraatbank.com.tr/v4/UIWebService/Search.aspx",
}

// Config gateway yapılandırması.
type Config struct {
	Credentials   model.BankCredentials
	Endpoints     Endpoints
	SkipHashCheck bool
}
