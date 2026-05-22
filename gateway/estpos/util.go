package estpos

import (
	"errors"
	"strings"

	"github.com/zatrano/gopostr/model"
)

func payloadVal(m map[string]string, key string) string {
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

func validateInit(req model.InitRequest) error {
	if req.Order.ID == "" {
		return errors.New("estpos: sipariş ID zorunlu")
	}
	if req.Order.Amount <= 0 {
		return errors.New("estpos: geçersiz tutar")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errors.New("estpos: success/fail URL zorunlu")
	}
	return nil
}

func cardExpMonth(card *model.CardInput) string {
	if card == nil {
		return ""
	}
	m := card.ExpireMonth
	if len(m) == 1 {
		m = "0" + m
	}
	return m
}

func cardExpYear(card *model.CardInput) string {
	if card == nil {
		return ""
	}
	y := card.ExpireYear
	if len(y) > 2 {
		y = y[len(y)-2:]
	}
	return y
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
