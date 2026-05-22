package model

// 3D işlem modelleri.
const (
	PaymentModel3DSecure      = "3d"
	PaymentModel3DPay         = "3d_pay"
	PaymentModel3DHost        = "3d_host"
	PaymentModel3DPayHosting  = "3d_pay_hosting"
)

// İşlem türleri.
const (
	TxTypePayAuth      = "pay"
	TxTypePayPreAuth   = "pre"
	TxTypePayPostAuth  = "post"
	TxTypeCancel       = "cancel"
	TxTypeRefund       = "refund"
	TxTypeRefundPartial = "refund_partial"
	TxTypeStatus       = "status"
)

// Para birimleri (ISO 4217).
const (
	CurrencyTRY = "TRY"
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
	CurrencyGBP = "GBP"
	CurrencyJPY = "JPY"
	CurrencyRUB = "RUB"
)

// Dil kodları.
const (
	LangTR = "tr"
	LangEN = "en"
)
