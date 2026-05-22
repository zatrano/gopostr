package interpos

import (
	"fmt"
	"regexp"
	"strings"
)

var delimSplit = regexp.MustCompile(`;;;|;;`)

// decodeResponse InterPos delimited key=value yanıtını parse eder.
func decodeResponse(body []byte) (map[string]string, error) {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return nil, fmt.Errorf("interpos: boş yanıt")
	}
	parts := delimSplit.Split(text, -1)
	out := make(map[string]string, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		out[kv[0]] = kv[1]
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("interpos: yanıt parse edilemedi")
	}
	return out, nil
}
