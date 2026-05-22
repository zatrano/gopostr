# gopostr

Türkiye **banka sanal POS** entegrasyonları için **saf Go** kütüphanesi.

- Yalnızca lisanslı bankalar
- Harici bağımlılık yok (Go 1.25+, standart kütüphane)
- Tek arayüz: `gateway.Gateway`
- 3D Secure, 3D Pay, 3D Host

```bash
go get github.com/zatrano/gopostr
```

## İçindekiler

1. [Proje yapısı](#proje-yapısı)
2. [Desteklenen bankalar](#desteklenen-bankalar)
3. [Kurulum](#kurulum)
4. [Hızlı başlangıç](#hızlı-başlangıç)
5. [Kart bilgisi ve `CardInput`](#kart-bilgisi-ve-cardinput)
6. [Gateway arayüzü](#gateway-arayüzü)
7. [3D modelleri](#3d-modelleri)
8. [Kimlik bilgileri](#kimlik-bilgileri)
9. [3D işlem akışı](#3d-işlem-akışı)
10. [Callback](#callback)
11. [İptal, iade ve sorgu](#iptal-iade-ve-sorgu)
12. [Testler](#testler)
13. [Katkı](#katkı)

---

## Proje yapısı

```
gopostr/
├── factory/              # factory.New("garanti", creds)
├── gateway/
│   ├── interface.go      # Gateway arayüzü
│   ├── README.md         # Banka paketi dosya şablonu
│   ├── estpos/           # Payten/Asseco (TEB, İş, Halk, Şeker, Ziraat, eski Akbank)
│   ├── garanti/
│   ├── akbank/           # Yeni Akbank JSON API (EstPOS değil)
│   ├── payflex/
│   ├── payflexcpv4/
│   ├── payfor/
│   ├── posnet/           # Yapı Kredi (XML)
│   ├── posnetv1/         # Albaraka (JSON) — posnet ile aynı değil
│   ├── interpos/
│   ├── kuveyt/
│   └── vakifkatilim/
├── model/                # Order, InitRequest, PaymentResult, sabitler
└── crypt/                # Ortak hash / rastgele string yardımcıları
```

Her banka alt paketi **aynı dosya düzenine** sahiptir:

| Dosya | Görev |
|-------|--------|
| `config.go` | `Config`, endpoint'ler, test URL'leri |
| `gateway.go` | `Init`, `HandleCallback`, `Cancel`, `Refund`, `Status` |
| `client.go` | HTTP istemcisi |
| `crypt.go` | Hash / MAC |
| `mapper.go` | `model.PaymentResult` eşlemesi |
| `currency.go` | Para birimi kodları |
| `util.go` | `payloadVal`, `validateInit`, yardımcılar |
| `xml.go` veya `json.go` | Banka protokolü (XML veya JSON, biri) |
| `gateway_test.go` | Akış testleri |
| `crypt_test.go` | Kripto birim testleri |
| `testdata/` | Örnek banka yanıtları |

Ayrıntılar: [`gateway/README.md`](gateway/README.md)

---

## Desteklenen bankalar

### Factory anahtarları

| Factory | Banka | Protokol |
|---------|-------|----------|
| `estpos` | TEB, İş Bankası, Halkbank, Şekerbank, Ziraat, eski Akbank (Payten/Asseco) | XML (EstV3) |
| `garanti` | Garanti BBVA | XML |
| `akbank` | Akbank (yeni sanal POS) | JSON |
| `payflex` | Vakıfbank, Ziraat Bankası | XML (PayFlex VPOS) |
| `payflexcpv4` | Vakıfbank | XML (PayFlex CP v4) |
| `payfor` | QNB Finansbank, Enpara, Ziraat Katılım | XML (PayFor) |
| `posnet` | Yapı Kredi | XML (PosNet) |
| `posnetv1` | Albaraka Türk | JSON (PosNet V1) |
| `interpos` | Denizbank | InterPos |
| `kuveyt` | Kuveyt Türk | XML + SOAP |
| `vakifkatilim` | Vakıf Katılım | XML |

`posnet` ve `posnetv1` **farklı bankalar ve farklı API**’lerdir; birbirinin yerine kullanılamaz.

### İşlem desteği

| Gateway | 3D Secure | 3D Pay | 3D Host | İptal | İade | Durum |
|---------|:---------:|:------:|:-------:|:-----:|:----:|:-----:|
| estpos | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| garanti | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| akbank | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| payflex | ✓ | — | — | ✓ | ✓ | ✓ |
| payflexcpv4 | ✓ | ✓ | ✓ | — | — | — |
| payfor | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| posnet | ✓ | — | — | ✓ | ✓ | ✓ |
| posnetv1 | ✓ | — | — | ✓ | ✓ | ✓ |
| interpos | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| kuveyt | ✓ | — | — | ✓ | ✓ | ✓ |
| vakifkatilim | ✓ | — | ✓ | ✓ | ✓ | ✓ |

- `estpos`: Payten/Asseco **EstV3** (SHA-512); yeni Akbank için `akbank` kullanın
- `akbank`: ayrı `Status()` yok
- `payflexcpv4`: iptal/iade/status yok; sonuç `HandleCallback` ile doğrulanır

Factory sabitleri: `factory.GatewayGaranti`, `factory.GatewayAkbank`, …

---

## Kurulum

Go **1.25+**

```bash
go get github.com/zatrano/gopostr@latest
```

```go
import (
    "github.com/zatrano/gopostr/factory"
    "github.com/zatrano/gopostr/model"
)
```

---

## Hızlı başlangıç

```go
creds := model.BankCredentials{
    ClientID: "MAĞAZA_NO",
    Username: "API_KULLANICI",
    Password: "API_ŞİFRE",
    StoreKey: "3D_HASH_ANAHTARI",
    Lang:     model.LangTR,
}

gw, err := factory.New(factory.GatewayGaranti, creds)

form, err := gw.Init(ctx, model.InitRequest{
    Order: model.Order{
        ID:          "SIPARIS-001",
        Amount:      149.90,
        Currency:    model.CurrencyTRY,
        Installment: 1,
        SuccessURL:  "https://magaza.com/odeme/basarili",
        FailURL:     "https://magaza.com/odeme/hata",
        IP:          "185.0.0.1",
    },
    PaymentModel: model.PaymentModel3DSecure,
    TxType:       model.TxTypePayAuth,
    // Card: nil — 3D Secure / 3D Pay'de kart bilgisi Init'e verilmez; müşteri bankada girer.
})
// form.Gateway, form.Method, form.Inputs → tarayıcıya yönlendir
```

---

## Kart bilgisi ve `CardInput`

**Önerilen kullanım (PCI):** `3d` ve `3d_pay` modellerinde `InitRequest.Card` **nil** bırakın. Müşteri kartını bankanın 3D sayfasında girer; PAN/CVV sizin sunucunuza gitmez.

| Model | `Card` Init'te? | Nerede girilir? |
|-------|-----------------|-----------------|
| `3d` (3D Secure) | Hayır (önerilen) | Banka 3D sayfası |
| `3d_pay` (3D Pay) | Hayır (önerilen) | Banka 3D sayfası |
| `3d_host` (3D Host) | Hayır | Bankanın host ödeme sayfası |
| `3d_pay_hosting` | Hayır | Banka / hosting sayfası |

`model.CardInput` yalnızca **eski entegrasyonlar** veya bankanın zorunlu kıldığı form POST akışları içindir (ör. bazı gateway'ler Init'te kart alanlarını HTML forma yazar). Üretimde mümkün olduğunca kullanmayın.

**Kütüphane notu:** `posnet` (Yapı Kredi) enrollment API'si nedeniyle `Init` sırasında kart bilgisi ister; `estpos` `3d` modelinde kart yoksa hata döner. Bu teknik zorunluluklar PHP referansından gelir; yine de kartı kendi sitenizde toplayıp sunucuya göndermek yerine banka sayfasına yönlendirme mümkünse tercih edilir.

---

## Gateway arayüzü

Tüm bankalar `gateway.Gateway` implement eder:

| Metot | Açıklama |
|-------|----------|
| `Init` | 3D işlemi başlatır; `model.FormData` döner |
| `HandleCallback` | Banka dönüşünü işler; `model.PaymentResult` |
| `Cancel` | İptal |
| `Refund` | İade |
| `Status` | Durum sorgusu (destekleyen bankalarda) |
| `Name` | Gateway adı |

Doğrudan banka paketi:

```go
import "github.com/zatrano/gopostr/gateway/garanti"

gw := garanti.New(garanti.Config{
    Credentials: creds,
    Endpoints:   garanti.DefaultTestEndpoints,
})
```

---

## 3D modelleri

| Sabit | Değer | Açıklama |
|-------|-------|----------|
| `PaymentModel3DSecure` | `3d` | Klasik 3D Secure |
| `PaymentModel3DPay` | `3d_pay` | 3D Pay |
| `PaymentModel3DHost` | `3d_host` | 3D Host |
| `PaymentModel3DPayHosting` | `3d_pay_hosting` | PayFor hosting |

İşlem türleri: `TxTypePayAuth`, `TxTypePayPreAuth`, `TxTypeCancel`, `TxTypeRefund`, `TxTypeRefundPartial`, `TxTypeStatus`.

Para birimi: `CurrencyTRY`, `CurrencyUSD`, `CurrencyEUR`, …

---

## Kimlik bilgileri

Ortak struct: `model.BankCredentials`

| Alan | Açıklama |
|------|----------|
| `ClientID` | Üye işyeri / mağaza no |
| `Username` / `Password` | API kullanıcısı |
| `StoreKey` | 3D hash / MAC anahtarı |
| `TerminalID` | Terminal no |
| `PosNetID` | PosNet üye no (`posnet`, `posnetv1`) |
| `Lang` | `tr` veya `en` |
| `MbrID` | PayFor kurum kodu |
| `MerchantType` / `SubMerchantID` | PayFlex alt bayi |
| `RefundUsername` / `RefundPassword` | Garanti iptal API |
| `TestMode` | Garanti test ortamı |

**Yapı Kredi (`posnet`):** `ClientID`, `Username`, `Password`, `StoreKey`, `PosNetID`

**Albaraka (`posnetv1`):** `ClientID`, `PosNetID`, `TerminalID`, `StoreKey`

**Kuveyt Türk:** `Password` = CustomerId; iptal/iade için `Order.RecurringID` banka sipariş no

**PayFlex CP v4:** `ClientID`, `Password`, `TerminalID`

---

## 3D işlem akışı

```
Uygulama → Init() → FormData → müşteri + banka 3D → HandleCallback() → PaymentResult
```

`model.FormData`: `Gateway` (URL), `Method`, `Inputs`, isteğe bağlı `HTML`.

- POST: gizli alanlarla form veya `HTML` otomatik gönderim
- GET: `Gateway` + `Inputs` query (ör. PayFlex CP `Ptkn`)

---

## Callback

```go
payload := map[string]string{}
// r.ParseForm() veya r.URL.Query() ile doldurun

result, err := gw.HandleCallback(ctx, payload)
if result.Success {
    // result.OrderID, result.AuthCode, result.HostRefNum
}
```

Production’da callback hash doğrulamasını kapatmayın (`SkipHashCheck` yalnızca test içindir).

### Yapı Kredi (`posnet`) — callback'te state

Banka dönüşünde genelde şu alanlar gelir: `BankPacket`, `MerchantPacket`, `Sign` (bazen `bankData` / `merchantData`).

`HandleCallback` ayrıca **sipariş bağlamı** ister: `orderId`, `amount`, `currency`. Bunlar bankanın POST'unda çoğu zaman **yoktur**; `Init` sırasında sizin oluşturduğunuz sipariş kaydından callback handler'da `payload` map'ine **eklemeniz** gerekir (bkz. `gateway/posnet` testleri).

**Nerede saklanır?** gopostr state tutmaz; uygulama katmanının sorumluluğu:

| Yöntem | Ne zaman? |
|--------|-----------|
| **Redis** (önerilen) | Birden fazla API instance, kısa TTL (15–30 dk), key = `orderId` |
| **Oturum (session)** | Tek sunucu / sticky session; Fiber `session` middleware |
| **Veritabanı** | Sipariş zaten `pending_payment` satırında; callback'te aynı satır okunur |

Örnek akış (Fiber + Redis):

```
1. POST /odeme/baslat → sipariş DB'ye yaz, Init(orderId, amount, …)
2. Redis SET pos:3d:{orderId} → {amount, currency, gateway}  TTL 20m
3. Müşteriyi form.Gateway'e yönlendir
4. GET/POST /odeme/callback → banka alanlarını parse et
5. Redis GET pos:3d:{orderId} → payload["orderId"], payload["amount"], payload["currency"] ile birleştir
6. gw.HandleCallback(ctx, payload) → başarılıysa siparişi DB'de paid yap, Redis DEL
```

`BankPacket` / `MerchantPacket` / `Sign` Init'ten gelmez; yalnızca bankanın callback POST'undan gelir. Init sonrası saklanması gerekenler: **sizin** `orderId`, `amount`, `currency` (ve isteğe bağlı `installment`, `txType`).

### Diğer gateway notları

| Gateway | Not |
|---------|-----|
| `payflexcpv4` | `TransactionId`, `PaymentToken`; `Rc != 0000` reddedilmiş işlem |

Çoğu bankada callback payload tek başına yeterlidir (`oid`, `amount` bankadan gelir). `posnet` bu kuralın istisnasıdır.

---

## İptal, iade ve sorgu

```go
gw.Cancel(ctx, model.CancelRequest{
    Order: model.Order{ID: "SIPARIS-001", RefRetNum: "HOST_REF"},
})
gw.Refund(ctx, model.RefundRequest{
    Order: model.Order{ID: "SIPARIS-001", Amount: 50},
    Partial: true,
})
gw.Status(ctx, model.StatusRequest{
    Order: model.Order{ID: "SIPARIS-001"},
})
```

---

## Testler

```bash
go test ./...
```

Her banka paketi `gateway_test.go`, `crypt_test.go` ve `testdata/` fixture’ları ile test edilir; gerçek bankaya istek gitmez.

---

## Katkı

1. Fork ve branch
2. `go test ./...` geçmeli
3. Yeni banka: `gateway/README.md` şablonuna uygun paket + `factory` kaydı

[GitHub Issues](https://github.com/zatrano/gopostr/issues)

---

## Lisans

MIT — [`LICENSE`](LICENSE)

---

<p align="center">
  <strong>© 2026 ZATRANO</strong><br>
  Yazar: <a href="https://github.com/serhankarakoc">serhankarakoc</a><br>
  <a href="https://github.com/zatrano/gopostr">github.com/zatrano/gopostr</a>
</p>
