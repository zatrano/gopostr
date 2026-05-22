package vakifkatilim

import (
	"fmt"
	"regexp"
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

func accountFields(c model.BankCredentials) map[string]string {
	sub := c.SubMerchantID
	if sub == "" {
		sub = "0"
	}
	return map[string]string{
		"MerchantId":    c.ClientID,
		"CustomerId":    c.Password,
		"UserName":      c.Username,
		"SubMerchantId": sub,
	}
}

func hashInput(c model.BankCredentials, extra map[string]string) map[string]string {
	h := accountFields(c)
	for k, v := range extra {
		h[k] = v
	}
	return h
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func mergeRaw(parts ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, p := range parts {
		for k, v := range p {
			out[k] = v
		}
	}
	return out
}

func bankOrderID(order model.Order) string {
	if order.RecurringID != "" {
		return order.RecurringID
	}
	return "0"
}

var (
	formActionRe = regexp.MustCompile(`(?i)<form[^>]*action=["']([^"']+)["']`)
	inputRe      = regexp.MustCompile(`(?i)<input[^>]*>`)
	nameRe       = regexp.MustCompile(`(?i)name=["']([^"']+)["']`)
	valueRe      = regexp.MustCompile(`(?i)value=["']([^"']*)["']`)
)

func parseHTMLForm(html string) (gateway string, inputs map[string]string, err error) {
	if !strings.Contains(strings.ToLower(html), "<form") {
		return "", nil, fmt.Errorf("vakifkatilim: HTML form bulunamadı")
	}
	m := formActionRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return "", nil, fmt.Errorf("vakifkatilim: form action bulunamadı")
	}
	inputs = make(map[string]string)
	for _, tag := range inputRe.FindAllString(html, -1) {
		nm := nameRe.FindStringSubmatch(tag)
		if len(nm) < 2 {
			continue
		}
		name := nm[1]
		if name == "submit" || name == "submitBtn" {
			continue
		}
		val := ""
		if vm := valueRe.FindStringSubmatch(tag); len(vm) >= 2 {
			val = vm[1]
		}
		inputs[name] = val
	}
	return m[1], inputs, nil
}

func isHTMLResponse(body []byte) bool {
	s := strings.ToLower(string(body))
	return strings.Contains(s, "<html") || strings.Contains(s, "<form")
}
