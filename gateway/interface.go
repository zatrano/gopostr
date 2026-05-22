package gateway

import (
	"context"

	"github.com/zatrano/gopostr/model"
)

// Gateway banka sanal POS entegrasyonu. Init en az 3D Secure destekler.
type Gateway interface {
	// Init 3D işlemi başlatır; form alanları veya redirect URL üretir.
	Init(ctx context.Context, req model.InitRequest) (model.FormData, error)

	// HandleCallback banka callback'ini işler ve sonucu normalize eder.
	HandleCallback(ctx context.Context, payload map[string]string) (model.PaymentResult, error)

	// Cancel ödeme iptali yapar.
	Cancel(ctx context.Context, req model.CancelRequest) (model.PaymentResult, error)

	// Refund ödeme iadesi yapar.
	Refund(ctx context.Context, req model.RefundRequest) (model.PaymentResult, error)

	// Status sipariş durumunu sorgular.
	Status(ctx context.Context, req model.StatusRequest) (model.PaymentResult, error)

	// Name gateway tanımlayıcı adını döner.
	Name() string
}
