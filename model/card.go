package model

// CreditCard kart bilgisi modeli. InitRequest.Card ile kullanılır.
type CreditCard struct {
	Number      string
	HolderName  string
	ExpireMonth string
	ExpireYear  string
	CVV         string
}
