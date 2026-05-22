package kuveyt

import (
	"strconv"

	"github.com/zatrano/gopostr/model"
)

const procSuccess = "00"

var currencyToCode = map[string]string{
	model.CurrencyTRY: "0949",
	model.CurrencyUSD: "0840",
	model.CurrencyEUR: "0978",
}

var codeToCurrency = map[string]string{
	"0949": model.CurrencyTRY,
	"949":  model.CurrencyTRY,
	"0840": model.CurrencyUSD,
	"840":  model.CurrencyUSD,
	"0978": model.CurrencyEUR,
	"978":  model.CurrencyEUR,
}

var cardBrandToType = map[string]string{
	"visa":       "Visa",
	"mastercard": "MasterCard",
	"troy":       "Troy",
}

func mapCurrency(c string) string {
	if code, ok := currencyToCode[c]; ok {
		return code
	}
	return c
}

func parseCurrency(code string) string {
	if len(code) == 3 {
		code = "0" + code
	}
	if c, ok := codeToCurrency[code]; ok {
		return c
	}
	return code
}

func formatAmountKurus(amount float64) int {
	return int(amount*100 + 0.5)
}

func parseAmountKurus(n int) float64 {
	return float64(n) / 100
}

func mapInstallment(n int) string {
	if n > 1 {
		return strconv.Itoa(n)
	}
	return "0"
}

func mapCardType(brand string) string {
	if t, ok := cardBrandToType[brand]; ok {
		return t
	}
	return "Visa"
}

func cardExpiryParts(card *model.CardInput) (month, year string) {
	if card == nil {
		return "", ""
	}
	m := card.ExpireMonth
	if len(m) == 1 {
		m = "0" + m
	}
	y := card.ExpireYear
	if len(y) > 2 {
		y = y[len(y)-2:]
	}
	return m, y
}
