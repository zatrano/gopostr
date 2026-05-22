package posnetv1

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
)

const (
	orderIDLength      = 20
	orderIDPrefix3D    = "TDS_"
	orderIDTotalLength = 24
)

func payloadVal(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		return v
	}
	kl := strings.ToLower(key)
	for k, v := range m {
		if strings.ToLower(k) == kl {
			return v
		}
	}
	return ""
}

func formatAmountKurus(amount float64) int {
	return int(math.Round(amount * 100))
}

func parseAmountKurus(s string) float64 {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return float64(n) / 100
}

func mapInstallment(n int) string {
	if n > 1 {
		return strconv.Itoa(n)
	}
	return "0"
}

func mapLang(lang string) string {
	if lang == model.LangEN {
		return "EN"
	}
	return "TR"
}

func terminalNo(c model.BankCredentials) string {
	if c.TerminalID != "" {
		return c.TerminalID
	}
	return c.Password
}

func posNetID(c model.BankCredentials) string {
	if c.PosNetID != "" {
		return c.PosNetID
	}
	return c.Username
}

func formatOrderID(id string) (string, error) {
	if len(id) > orderIDLength {
		return "", fmt.Errorf("posnetv1: sipariş ID en fazla %d karakter olabilir", orderIDLength)
	}
	return strings.Repeat("0", orderIDLength-len(id)) + id, nil
}

func prefixedOrderID(id, paymentModel string) (string, error) {
	prefix := ""
	if paymentModel == model.PaymentModel3DSecure || paymentModel == "" {
		prefix = orderIDPrefix3D
	}
	padLen := orderIDTotalLength - len(prefix)
	formatted, err := formatOrderIDWithLen(id, padLen)
	if err != nil {
		return "", err
	}
	return prefix + formatted, nil
}

func formatOrderIDWithLen(id string, padLen int) (string, error) {
	if len(id) > padLen {
		return "", fmt.Errorf("posnetv1: sipariş ID en fazla %d karakter olabilir", padLen)
	}
	return strings.Repeat("0", padLen-len(id)) + id, nil
}

func cardExpiryYM(card *model.CardInput) string {
	if card == nil {
		return ""
	}
	m := card.ExpireMonth
	if len(m) == 1 {
		m = "0" + m
	}
	y := card.ExpireYear
	if len(y) == 4 {
		y = y[2:]
	}
	return y + m
}

func txTypeToAPI(tx string) (string, error) {
	switch tx {
	case model.TxTypePayAuth:
		return "Sale", nil
	case model.TxTypePayPreAuth:
		return "Auth", nil
	case model.TxTypeCancel:
		return "Reverse", nil
	case model.TxTypeRefund, model.TxTypeRefundPartial:
		return "Return", nil
	case model.TxTypeStatus:
		return "TransactionInquiry", nil
	default:
		return "", fmt.Errorf("posnetv1: desteklenmeyen işlem: %s", tx)
	}
}

func joinAPI(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

func validateInit(req model.InitRequest) error {
	if req.PaymentModel != model.PaymentModel3DSecure {
		return fmt.Errorf("posnetv1: yalnızca 3D Secure desteklenir")
	}
	if req.TxType != model.TxTypePayAuth && req.TxType != model.TxTypePayPreAuth {
		return fmt.Errorf("posnetv1: desteklenmeyen işlem türü: %s", req.TxType)
	}
	if req.Order.ID == "" {
		return errMissing("sipariş id")
	}
	if req.Order.Amount <= 0 {
		return errMissing("tutar")
	}
	if req.Order.SuccessURL == "" {
		return errMissing("success URL")
	}
	if req.Card == nil {
		return errMissing("kart bilgisi")
	}
	return nil
}

type posnetv1Error string

func (e posnetv1Error) Error() string { return "posnetv1: " + string(e) }

func errMissing(field string) error {
	return posnetv1Error(field + " zorunlu")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func is3DAuthSuccess(mdStatus string) bool {
	return mdStatus == "1"
}
