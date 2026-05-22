package payflexcpv4

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zatrano/gopostr/model"
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

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func parseAmount(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func mapLang(lang string) string {
	if lang == model.LangEN {
		return "en-US"
	}
	return "tr-TR"
}

func cardExpireYear(card *model.CardInput) string {
	if card == nil {
		return ""
	}
	y := card.ExpireYear
	if len(y) == 2 {
		return "20" + y
	}
	return y
}

func validateInit(req model.InitRequest) error {
	if req.Order.ID == "" {
		return errMissing("sipariş id")
	}
	if req.Order.Amount <= 0 {
		return errMissing("tutar")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errMissing("success/fail URL")
	}
	switch req.PaymentModel {
	case model.PaymentModel3DSecure, model.PaymentModel3DPay:
		if req.Card == nil {
			return errMissing("kart bilgisi")
		}
	case model.PaymentModel3DHost:
		// kart banka sayfasında
	default:
		return fmt.Errorf("payflexcpv4: desteklenmeyen 3D modeli: %s", req.PaymentModel)
	}
	if req.TxType != model.TxTypePayAuth {
		return fmt.Errorf("payflexcpv4: yalnızca satış (pay_auth) desteklenir")
	}
	return nil
}

type cpError string

func (e cpError) Error() string { return "payflexcpv4: " + string(e) }

func errMissing(field string) error {
	return cpError(field + " zorunlu")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
