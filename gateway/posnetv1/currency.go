package posnetv1

import "github.com/zatrano/gopostr/model"

func mapCurrency(c string) string {
	switch c {
	case model.CurrencyUSD:
		return "US"
	case model.CurrencyEUR:
		return "EU"
	case model.CurrencyGBP:
		return "GB"
	case model.CurrencyJPY:
		return "JP"
	case model.CurrencyRUB:
		return "RU"
	default:
		return "TL"
	}
}
