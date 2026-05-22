package payflexcpv4

import (
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

func mapCurrency(c string) string {
	switch c {
	case model.CurrencyUSD:
		return "840"
	case model.CurrencyEUR:
		return "978"
	case model.CurrencyGBP:
		return "826"
	case model.CurrencyJPY:
		return "392"
	case model.CurrencyRUB:
		return "643"
	default:
		return "949"
	}
}

func mapBrand(brand string) string {
	switch strings.ToLower(brand) {
	case "visa":
		return "100"
	case "mastercard", "master":
		return "200"
	case "troy":
		return "300"
	case "amex", "american express":
		return "400"
	default:
		return ""
	}
}

func mapInstallment(n int) string {
	if n > 1 {
		return strconv.Itoa(n)
	}
	return "0"
}
