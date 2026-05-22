package payfor

import (
	"strconv"

	"github.com/zatrano/gopostr/model"
)

const procSuccess = "00"

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

var modelToSecure = map[string]string{
	model.PaymentModel3DSecure: "3DModel",
	model.PaymentModel3DPay:    "3DPay",
	model.PaymentModel3DHost:   "3DHost",
}

var secureToModel = map[string]string{
	"3DModel": model.PaymentModel3DSecure,
	"3DPay":   model.PaymentModel3DPay,
	"3DHost":  model.PaymentModel3DHost,
}

var txTypeToPayfor = map[string]string{
	model.TxTypePayAuth:       "Auth",
	model.TxTypePayPreAuth:    "PreAuth",
	model.TxTypeCancel:        "Void",
	model.TxTypeRefund:        "Refund",
	model.TxTypeRefundPartial: "Refund",
	model.TxTypeStatus:        "OrderInquiry",
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

func mapLang(lang string) string {
	if lang == model.LangEN {
		return "en"
	}
	return "tr"
}

func cardExpiryMY(card *model.CardInput) string {
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
	return m + y
}
