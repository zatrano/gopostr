package akbank

import (
	"strconv"

	"github.com/zatrano/gopostr/model"
)

const procSuccess = "VPS-0000"

var currencyToCode = map[string]int{
	model.CurrencyTRY: 949,
	model.CurrencyUSD: 840,
	model.CurrencyEUR: 978,
	model.CurrencyJPY: 392,
	model.CurrencyRUB: 643,
}

var codeToCurrency = map[int]string{
	949: model.CurrencyTRY,
	840: model.CurrencyUSD,
	978: model.CurrencyEUR,
	392: model.CurrencyJPY,
	643: model.CurrencyRUB,
}

func mapCurrency(c string) int {
	if code, ok := currencyToCode[c]; ok {
		return code
	}
	if n, err := strconv.Atoi(c); err == nil {
		return n
	}
	return 949
}

func parseCurrency(code string) string {
	if n, err := strconv.Atoi(code); err == nil {
		if c, ok := codeToCurrency[n]; ok {
			return c
		}
	}
	return code
}

func formatAmount(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

func mapInstallment(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

func mapLang(lang string) string {
	if lang == model.LangEN {
		return "EN"
	}
	return "TR"
}
