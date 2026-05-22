package model

// Order işlem veya sorguda kullanılan sipariş bilgileri.
type Order struct {
	ID              string
	Amount          float64
	Currency        string
	Installment     int
	IP              string
	SuccessURL      string
	FailURL         string
	PreAuthAmount   *float64
	RecurringID     string
	RecurringInstNo int
	// RefRetNum iptal/iade için host referans numarası (Garanti vb.).
	RefRetNum string
	// TransactionID PayFlex iptal/iade/sorgu referans işlem numarası.
	TransactionID string
}

// InitRequest 3D işlem başlatma isteğidir.
type InitRequest struct {
	Order        Order
	PaymentModel string // PaymentModel3DSecure, PaymentModel3DPay, ...
	TxType       string // TxTypePayAuth, TxTypePayPreAuth
	// Card çoğu 3D akışta nil olmalıdır (kart banka sayfasında girilir).
	// Yalnızca bankanın zorunlu kıldığı eski form-POST entegrasyonlarında kullanılır; bkz. README «Kart bilgisi».
	Card *CardInput
}

// CardInput isteğe bağlı; üretimde 3D Secure/Pay/Host için genelde nil bırakılır (PCI).
type CardInput struct {
	Number      string
	ExpireMonth string
	ExpireYear  string
	CVV         string
	// HolderName PayFlex provizyon için kart sahibi adı.
	HolderName string
	// Brand visa, mastercard, troy, amex (PayFlex BrandName kodu için).
	Brand string
}

// CallbackRequest banka dönüşü işlenirken kullanılan bağlam.
// HandleCallback yalnızca payload alsa da, sipariş meta verisi için bu struct tanımlıdır.
type CallbackRequest struct {
	Payload      map[string]string
	Order        Order
	PaymentModel string
	TxType       string
}

// CancelRequest iptal işlemi isteğidir.
type CancelRequest struct {
	Order Order
}

// RefundRequest iade işlemi isteğidir.
type RefundRequest struct {
	Order    Order
	Partial  bool
}

// StatusRequest durum sorgulama isteğidir.
type StatusRequest struct {
	Order Order
}
