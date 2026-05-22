package interpos

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
		return errors.New("interpos: sipariş ID gerekli")
	}
	if req.Order.Amount <= 0 {
		return errors.New("interpos: tutar gerekli")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errors.New("interpos: success/fail URL gerekli")
	}
	return nil
}
