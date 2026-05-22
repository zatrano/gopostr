package payflex

import (
	"errors"
	"strconv"

	"github.com/zatrano/gopostr/model"
)

func validateInit(req model.InitRequest) error {
	if req.Order.ID == "" {
		return errors.New("payflex: sipariş ID zorunlu")
	}
	if req.Order.Amount <= 0 {
		return errors.New("payflex: geçersiz tutar")
	}
	if req.Order.SuccessURL == "" || req.Order.FailURL == "" {
		return errors.New("payflex: success/fail URL zorunlu")
	}
	if req.Card == nil || req.Card.Number == "" {
		return errors.New("payflex: enrollment için kart bilgisi zorunlu")
	}
	return nil
}

func cardFromPayload(p map[string]string) (model.CardInput, string, error) {
	card := model.CardInput{
		Number:     payloadVal(p, "pan"),
		CVV:        payloadVal(p, "cvv"),
		HolderName: payloadVal(p, "cardHoldersName"),
	}
	expiry := payloadVal(p, "expiry")
	if expiry == "" {
		expiry = cardExpiryYMLong(card)
	}
	if card.Number == "" {
		return card, "", errors.New("payflex: provizyon için payload.pan zorunlu")
	}
	if card.CVV == "" {
		return card, "", errors.New("payflex: provizyon için payload.cvv zorunlu")
	}
	if expiry == "" {
		return card, "", errors.New("payflex: provizyon için payload.expiry zorunlu (Ym)")
	}
	return card, expiry, nil
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

func cardExpiryYMLong(card model.CardInput) string {
	y := card.ExpireYear
	m := card.ExpireMonth
	if len(m) == 1 {
		m = "0" + m
	}
	if len(y) == 2 {
		y = "20" + y
	}
	return y + m
}

func parseAmount(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
