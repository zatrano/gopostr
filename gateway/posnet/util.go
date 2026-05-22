package posnet

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
		return errors.New("posnet: sipariş ID gerekli")
	}
	if req.Order.Amount <= 0 {
		return errors.New("posnet: tutar gerekli")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errors.New("posnet: success/fail URL gerekli")
	}
	if req.Card == nil {
		return errors.New("posnet: kart bilgisi zorunlu")
	}
	if req.TxType != model.TxTypePayAuth && req.TxType != model.TxTypePayPreAuth {
		return errors.New("posnet: yalnızca satış veya ön provizyon desteklenir")
	}
	return nil
}
