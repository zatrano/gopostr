package posnet

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

const procSuccess = "1"

const (
	orderIDLen       = 20
	orderID3DPrefix  = "TDSC"
	orderIDTotalLen  = 24
)

var currencyToCode = map[string]string{
	model.CurrencyTRY: "TL",
	model.CurrencyUSD: "US",
	model.CurrencyEUR: "EU",
	model.CurrencyGBP: "GB",
	model.CurrencyJPY: "JP",
	model.CurrencyRUB: "RU",
}

var codeToCurrency = map[string]string{
	"TL": model.CurrencyTRY,
	"US": model.CurrencyUSD,
	"EU": model.CurrencyEUR,
}

var txTypeToPosnet = map[string]string{
	model.TxTypePayAuth:       "Sale",
	model.TxTypePayPreAuth:    "Auth",
	model.TxTypeCancel:        "reverse",
	model.TxTypeRefund:        "return",
	model.TxTypeRefundPartial: "return",
	model.TxTypeStatus:        "agreement",
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

func formatAmountKurus(amount float64) int {
	return int(amount*100 + 0.5)
}

func parseAmountKurus(s string) float64 {
	n, _ := strconv.Atoi(s)
	return float64(n) / 100
}

func mapInstallment(n int) string {
	if n > 1 {
		return fmt.Sprintf("%02d", n)
	}
	return "00"
}

func mapLang(lang string) string {
	if lang == model.LangEN {
		return "en"
	}
	return "tr"
}

func formatOrderID(id string) (string, error) {
	if len(id) > orderIDLen {
		return "", fmt.Errorf("posnet: sipariş ID en fazla %d karakter", orderIDLen)
	}
	return strings.Repeat("0", orderIDLen-len(id)) + id, nil
}

func prefixedOrderID(id string) (string, error) {
	formatted, err := formatOrderID(id)
	if err != nil {
		return "", err
	}
	prefixLen := orderIDTotalLen - len(orderID3DPrefix)
	if len(formatted) > prefixLen {
		formatted = formatted[len(formatted)-prefixLen:]
	}
	return orderID3DPrefix + formatted, nil
}

func cardExpiryYM(card *model.CardInput) string {
	if card == nil {
		return ""
	}
	y := card.ExpireYear
	if len(y) > 2 {
		y = y[len(y)-2:]
	}
	m := card.ExpireMonth
	if len(m) == 1 {
		m = "0" + m
	}
	return y + m
}
