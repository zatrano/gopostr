package factory

import (
	"fmt"

	"github.com/zatrano/gopostr/gateway"
	"github.com/zatrano/gopostr/gateway/akbank"
	"github.com/zatrano/gopostr/gateway/estpos"
	"github.com/zatrano/gopostr/gateway/interpos"
	"github.com/zatrano/gopostr/gateway/kuveyt"
	"github.com/zatrano/gopostr/gateway/garanti"
	"github.com/zatrano/gopostr/gateway/payflex"
	"github.com/zatrano/gopostr/gateway/payflexcpv4"
	"github.com/zatrano/gopostr/gateway/payfor"
	"github.com/zatrano/gopostr/gateway/posnet"
	"github.com/zatrano/gopostr/gateway/posnetv1"
	"github.com/zatrano/gopostr/gateway/vakifkatilim"
	"github.com/zatrano/gopostr/model"
)

// Banka gateway ad sabitleri.
const (
	GatewayGaranti       = "garanti"
	GatewayAkbank        = "akbank"
	GatewayPayflex       = "payflex"
	GatewayPayfor        = "payfor"
	GatewayPosnet        = "posnet"
	GatewayInterpos      = "interpos"
	GatewayKuveyt        = "kuveyt"
	GatewayVakifkatilim  = "vakifkatilim"
	GatewayPosnetv1      = "posnetv1"
	GatewayPayflexcpv4   = "payflexcpv4"
	GatewayEstpos        = "estpos"
)

// New gateway adına göre instance oluşturur.
func New(name string, creds model.BankCredentials) (gateway.Gateway, error) {
	switch name {
	case GatewayGaranti:
		return garanti.New(garanti.Config{Credentials: creds}), nil
	case GatewayAkbank:
		return akbank.New(akbank.Config{Credentials: creds}), nil
	case GatewayPayflex:
		return payflex.New(payflex.Config{Credentials: creds}), nil
	case GatewayPayfor:
		return payfor.New(payfor.Config{Credentials: creds}), nil
	case GatewayPosnet:
		return posnet.New(posnet.Config{Credentials: creds}), nil
	case GatewayInterpos:
		return interpos.New(interpos.Config{Credentials: creds}), nil
	case GatewayKuveyt:
		return kuveyt.New(kuveyt.Config{Credentials: creds}), nil
	case GatewayVakifkatilim:
		return vakifkatilim.New(vakifkatilim.Config{Credentials: creds}), nil
	case GatewayPosnetv1:
		return posnetv1.New(posnetv1.Config{Credentials: creds}), nil
	case GatewayPayflexcpv4:
		return payflexcpv4.New(payflexcpv4.Config{Credentials: creds}), nil
	case GatewayEstpos:
		return estpos.New(estpos.Config{Credentials: creds}), nil
	default:
		return nil, fmt.Errorf("factory: bilinmeyen gateway: %s", name)
	}
}
