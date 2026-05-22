package estpos

import (
	"sort"
	"strings"

	"github.com/zatrano/gopostr/crypt"
)

const hashSeparator = "|"

var hashExcludeKeys = map[string]struct{}{
	"hash": {}, "encoding": {}, "nationalidno": {},
}

// Create3DHash EstV3 3D form hash (SHA-512, base64).
func Create3DHash(storeKey string, inputs map[string]string) string {
	keys := make([]string, 0, len(inputs))
	for k := range inputs {
		if _, skip := hashExcludeKeys[strings.ToLower(k)]; skip {
			continue
		}
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) < 0
	})

	values := make([]string, 0, len(keys)+1)
	for _, k := range keys {
		values = append(values, inputs[k])
	}
	values = append(values, storeKey)

	escaped := make([]string, len(values))
	for i, v := range values {
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, hashSeparator, `\`+hashSeparator)
		escaped[i] = v
	}
	return crypt.SHA512Base64([]byte(strings.Join(escaped, hashSeparator)))
}

// Check3DHash callback HASH doğrulaması (EstV3).
func Check3DHash(storeKey string, payload map[string]string) bool {
	expected := Create3DHash(storeKey, payload)
	return strings.EqualFold(payloadVal(payload, "HASH"), expected)
}
