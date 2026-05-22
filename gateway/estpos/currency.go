package estpos

import (
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

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
	"826": model.CurrencyGBP,
	"392": model.CurrencyJPY,
	"643": model.CurrencyRUB,
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
	if code == "*" {
		return model.CurrencyTRY
	}
	return code
}

func formatAmount(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

func parseAmount(s string, fallback float64) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fallback
	}
	return v
}

func mapInstallment(n int) string {
	if n > 1 {
		return strconv.Itoa(n)
	}
	return ""
}

func mapLang(lang string) string {
	if lang == model.LangEN {
		return "en"
	}
	return "tr"
}
