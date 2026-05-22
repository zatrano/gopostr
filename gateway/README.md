# Gateway paketleri

Her banka alt paketi aynı dosya düzenine sahiptir.

## Zorunlu dosyalar

| Dosya | Sorumluluk |
|-------|------------|
| `config.go` | `Config`, `Endpoints`, test URL'leri |
| `gateway.go` | `Gateway` implementasyonu (`Init`, `HandleCallback`, …) |
| `client.go` | HTTP istemcisi |
| `crypt.go` | Hash / MAC |
| `mapper.go` | `model.PaymentResult` eşlemesi |
| `currency.go` | Para birimi kodları |
| `util.go` | Ortak yardımcılar (`payloadVal`, `validateInit`, …) |
| `xml.go` **veya** `json.go` | Banka protokol serileştirme (biri) |
| `gateway_test.go` | Entegrasyon testleri |
| `crypt_test.go` | Kripto birim testleri |
| `testdata/` | Örnek banka yanıtları |

## Wire format

- XML API → `xml.go` (`estpos`, `garanti`, `payflex`, `payfor`, `posnet`, `interpos`, `kuveyt`, `vakifkatilim`, `payflexcpv4`)
- JSON API → `json.go` (`akbank`, `posnetv1`)

## Test

`gateway_test.go` içinde `testCreds()` tanımlıdır. Mümkün olduğunda `testdata/` fixture'ları `httptest` ile kullanılır.
