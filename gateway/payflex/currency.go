package payflex

import (
	"strconv"

	"github.com/zatrano/gopostr/model"
)

const procSuccess = "0000"

var currencyToCode = map[string]string{
	model.CurrencyTRY: "949",
	model.CurrencyUSD: "840",
	model.CurrencyEUR: "978",
	model.CurrencyGBP: "826",
	model.CurrencyJPY: "392",
	model.CurrencyRUB: "643",
}

var codeToCurrency = map[string]string{
	"949": model.CurrencyTRY,
	"840": model.CurrencyUSD,
	"978": model.CurrencyEUR,
}

var brandToCode = map[string]string{
	"visa":       "100",
	"mastercard": "200",
	"troy":       "300",
	"amex":       "400",
}

func mapCurrency(c string) string {
	if code, ok := currencyToCode[c]; ok {
		return code
	}
	return c
}

func parseCurrency(code string) string {
	if c, ok := codeToCurrency[code]; ok {
		return c
	}
	return code
}

func formatAmount(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

func mapInstallment(n int) string {
	if n > 1 {
		return strconv.Itoa(n)
	}
	return "0"
}

func mapBrand(brand string) string {
	if code, ok := brandToCode[brand]; ok {
		return code
	}
	return ""
}
