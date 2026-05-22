package model

// BankCredentials sanal POS API kimlik bilgilerini tutar.
type BankCredentials struct {
	// ClientID banka tarafından verilen üye işyeri numarası.
	ClientID string
	// Username API kullanıcı adı.
	Username string
	// Password API şifresi.
	Password string
	// StoreKey 3D Secure hash doğrulaması için gizli anahtar.
	StoreKey string
	// Lang varsayılan dil (tr/en).
	Lang string
	// TerminalID Garanti BBVA terminal numarası.
	TerminalID string
	// RefundUsername iptal/iade API kullanıcısı (Garanti; boşsa Username kullanılır).
	RefundUsername string
	// RefundPassword iptal/iade API şifresi (Garanti; boşsa Password kullanılır).
	RefundPassword string
	// TestMode true ise TEST modu (Garanti).
	TestMode bool
	// SubMerchantID Akbank alt üye işyeri numarası (opsiyonel).
	SubMerchantID string
	// MerchantType PayFlex üye işyeri tipi (0 standart, 1 ana bayi, 2 alt bayi).
	MerchantType int
	// MbrID PayFor kurum kodu (Finansbank: 5, Ziraat Katılım: 12).
	MbrID string
	// PosNetID Yapı Kredi PosNet üye işyeri no (Username ile aynı alan da kullanılabilir).
	PosNetID string
}
