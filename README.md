<div align="center">

<br>

[![Go 1.25+](https://img.shields.io/badge/Go-1.25%2B-7F77DD?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![MIT](https://img.shields.io/badge/Lisans-MIT-1D9E75?style=flat-square)](LICENSE)
[![Sıfır bağımlılık](https://img.shields.io/badge/Bağımlılık-Yok-185FA5?style=flat-square)]()

# gopostr

**Türkiye'nin lisanslı bankalarına yönelik sanal POS entegrasyonu için  
kurumsal düzeyde, saf Go kütüphanesi.**

Tek arayüz &nbsp;·&nbsp; Sıfır harici bağımlılık &nbsp;·&nbsp; Üretime hazır

```bash
go get github.com/zatrano/gopostr@latest
```

| 11 Gateway | 16+ Banka | 3D Secure · Pay · Host | 0 Bağımlılık |
|:---:|:---:|:---:|:---:|
| Lisanslı | BDDK onaylı | PCI uyumlu akış | Standart kütüphane |

</div>

---

## Neden gopostr?

| | gopostr |
|---|---|
| **Bağımlılık** | Yok — yalnızca Go standart kütüphanesi |
| **Banka desteği** | 11 gateway, 16+ banka |
| **Protokol** | XML, JSON, SOAP — banka ne istiyorsa |
| **3D modelleri** | 3D Secure · 3D Pay · 3D Host |
| **İşlemler** | Ödeme · İptal · İade · Durum sorgusu |
| **Test** | Her gateway için fixture tabanlı birim testleri |
| **Arayüz** | Tüm bankalar tek `gateway.Gateway` contract'ı |

---

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

| Factory | Desteklenen Bankalar | Protokol |
|---------|----------------------|----------|
| `estpos` | TEB, İş Bankası, Halkbank, Şekerbank, Ziraat Bankası, Akbank (eski) | XML — Payten/Asseco EstV3, SHA-512 |
| `garanti` | Garanti BBVA | XML |
| `akbank` | Akbank (yeni sanal POS API) | JSON |
| `payflex` | Vakıfbank, Ziraat Bankası | XML — PayFlex VPOS |
| `payflexcpv4` | Vakıfbank | XML — PayFlex CP v4 |
| `payfor` | QNB Finansbank, Enpara, Ziraat Katılım | XML — PayFor |
| `posnet` | Yapı Kredi Bankası | XML — PosNet |
| `posnetv1` | Albaraka Türk Katılım Bankası | JSON — PosNet V1 |
| `interpos` | Denizbank | InterPos |
| `kuveyt` | Kuveyt Türk Katılım Bankası | XML + SOAP |
| `vakifkatilim` | Vakıf Katılım Bankası | XML |

> **Not:** `posnet` (Yapı Kredi) ile `posnetv1` (Albaraka) farklı bankalar ve tamamen farklı API'lerdir; birbirinin yerine kullanılamaz. Yeni Akbank entegrasyonu için `estpos` değil `akbank` kullanın.

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

**Notlar:**
- `akbank`: ayrı `Status()` metodu yoktur; işlem sonucu `HandleCallback` ile doğrulanır
- `payflexcpv4`: iptal/iade/status desteklenmez; işlem sonucu yalnızca `HandleCallback` ile doğrulanır

Factory sabitleri: `factory.GatewayGaranti`, `factory.GatewayAkbank`, …

---

## Kurulum

Go **1.25+** gereklidir.

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

**Önerilen kullanım (PCI):** `3d` ve `3d_pay` modellerinde `InitRequest.Card` **nil** bırakın. Müşteri kartını bankanın 3D sayfasında girer; PAN/CVV sizin sunucunuza gelmez.

| Model | `Card` Init'te? | Nerede girilir? |
|-------|-----------------|-----------------|
| `3d` (3D Secure) | Hayır (önerilen) | Banka 3D sayfası |
| `3d_pay` (3D Pay) | Hayır (önerilen) | Banka 3D sayfası |
| `3d_host` (3D Host) | Hayır | Bankanın host ödeme sayfası |
| `3d_pay_hosting` | Hayır | Banka / hosting sayfası |

`model.CardInput` yalnızca **eski entegrasyonlar** veya bankanın teknik olarak zorunlu kıldığı form POST akışları içindir. Üretimde mümkün olduğunca kullanmayın.

**Banka özgü gereksinimler:** `posnet` (Yapı Kredi) enrollment API'si nedeniyle `Init` sırasında kart bilgisi ister; `estpos` `3d` modelinde kart yoksa hata döner. Bu teknik zorunluluklar ilgili banka API'sinden kaynaklanmaktadır.

---

## Gateway arayüzü

Tüm bankalar `gateway.Gateway` implement eder:

| Metot | Açıklama |
|-------|----------|
| `Init` | 3D işlemi başlatır; `model.FormData` döner |
| `HandleCallback` | Banka dönüşünü işler; `model.PaymentResult` döner |
| `Cancel` | İptal işlemi |
| `Refund` | İade işlemi |
| `Status` | Durum sorgusu (destekleyen bankalarda) |
| `Name` | Gateway adı |

Doğrudan banka paketini de kullanabilirsiniz:

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

İşlem türleri: `TxTypePayAuth`, `TxTypePayPreAuth`, `TxTypeCancel`, `TxTypeRefund`, `TxTypeRefundPartial`, `TxTypeStatus`

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

**Kuveyt Türk:** `Password` = CustomerId; iptal/iade için `Order.RecurringID` = banka sipariş no

**PayFlex CP v4:** `ClientID`, `Password`, `TerminalID`

---

## 3D işlem akışı

```
Uygulama → Init() → FormData → müşteri + banka 3D → HandleCallback() → PaymentResult
```

`model.FormData` alanları: `Gateway` (URL), `Method`, `Inputs`, isteğe bağlı `HTML`.

- **POST:** gizli alanlarla form veya `HTML` otomatik gönderim
- **GET:** `Gateway` + `Inputs` query string (ör. PayFlex CP `Ptkn`)

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

> Production ortamında callback hash doğrulamasını kapatmayın. `SkipHashCheck` yalnızca test içindir.

### Yapı Kredi (`posnet`) — callback'te state

Banka dönüşünde genelde şu alanlar gelir: `BankPacket`, `MerchantPacket`, `Sign` (bazen `bankData` / `merchantData`).

`HandleCallback` ayrıca **sipariş bağlamı** ister: `orderId`, `amount`, `currency`. Bu alanlar bankanın POST'unda çoğu zaman **bulunmaz**; `Init` sırasında oluşturduğunuz sipariş kaydından callback handler'da `payload` map'ine **eklemeniz** gerekir.

**gopostr state tutmaz;** bu sorumluluk uygulama katmanına aittir:

| Yöntem | Ne zaman tercih edilir? |
|--------|------------------------|
| **Redis** (önerilen) | Birden fazla API instance, kısa TTL (15–30 dk), key = `orderId` |
| **Oturum (session)** | Tek sunucu / sticky session |
| **Veritabanı** | Sipariş zaten `pending_payment` satırında; callback'te aynı satır okunur |

**Örnek akış (Fiber + Redis):**

```
1. POST /odeme/baslat  → siparişi DB'ye yaz, Init(orderId, amount, …)
2. Redis SET pos:3d:{orderId} → {amount, currency, gateway}  TTL 20m
3. Müşteriyi form.Gateway'e yönlendir
4. GET/POST /odeme/callback → banka alanlarını parse et
5. Redis GET pos:3d:{orderId} → payload'a orderId, amount, currency ekle
6. gw.HandleCallback(ctx, payload) → başarılıysa siparişi paid yap, Redis DEL
```

`BankPacket` / `MerchantPacket` / `Sign` Init'ten gelmez; yalnızca bankanın callback POST'undan gelir. Init sonrası saklanması gereken değerler: **sizin** `orderId`, `amount`, `currency` (ve isteğe bağlı `installment`, `txType`).

### Diğer gateway notları

| Gateway | Not |
|---------|-----|
| `payflexcpv4` | `TransactionId`, `PaymentToken`; `Rc != 0000` reddedilmiş işlem anlamına gelir |

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

Her banka paketi `gateway_test.go`, `crypt_test.go` ve `testdata/` fixture'ları ile test edilir; gerçek bankaya istek gitmez.

---

## Katkı

1. Fork ve branch oluşturun
2. `go test ./...` geçmeli
3. Yeni banka: `gateway/README.md` şablonuna uygun paket + `factory` kaydı ekleyin

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
