package model

// FormData 3D yönlendirme formu veya redirect bilgisini taşır.
type FormData struct {
	// Gateway bankanın 3D gateway URL'si.
	Gateway string
	// Method HTTP metodu (POST veya GET).
	Method string
	// Inputs form alanları (hash dahil).
	Inputs map[string]string
	// HTML isteğe bağlı önceden üretilmiş form HTML'i.
	HTML string
}

// PaymentResult tüm banka gateway'lerinin normalize ettiği işlem yanıtıdır.
type PaymentResult struct {
	Success       bool
	OrderID       string
	TransactionID string
	AuthCode      string
	HostRefNum    string
	Amount        float64
	Currency      string
	ErrorCode     string
	ErrorMessage  string
	RawResponse   map[string]string
}
