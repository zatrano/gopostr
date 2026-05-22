package garanti

import (
	"math"
	"strconv"

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
	return code
}

// formatAmountKurus float tutarı kuruş integer string'e (10.50 → "1050").
func formatAmountKurus(amount float64) string {
	return strconv.Itoa(int(math.Round(amount * 100)))
}

func parseAmountKurus(s string) float64 {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	return float64(n) / 100
}

func mapInstallment(n int) string {
	if n > 1 {
		return strconv.Itoa(n)
	}
	return ""
}
