package payflexcpv4

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strings"
)

func decodeXML(body []byte) (map[string]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(body))
	dec.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	out := make(map[string]string)
	var stack []string
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("payflexcpv4: xml decode: %w", err)
		}
		switch el := tok.(type) {
		case xml.StartElement:
			stack = append(stack, el.Name.Local)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			text := strings.TrimSpace(string(el))
			if text == "" || len(stack) == 0 {
				continue
			}
			out[stack[len(stack)-1]] = text
		}
	}
	if strings.Contains(strings.ToLower(string(body)), "<html") {
		return out, fmt.Errorf("payflexcpv4: HTML yanıt alındı")
	}
	return out, nil
}

func encodeForm(fields map[string]string) string {
	vals := make(url.Values, len(fields))
	for k, v := range fields {
		if v != "" {
			vals.Set(k, v)
		}
	}
	return vals.Encode()
}
