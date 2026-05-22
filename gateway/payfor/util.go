package payfor

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
		return errors.New("payfor: sipariş ID zorunlu")
	}
	if req.Order.Amount <= 0 {
		return errors.New("payfor: geçersiz tutar")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errors.New("payfor: success/fail URL zorunlu")
	}
	return nil
}
