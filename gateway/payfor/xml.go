package payfor

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type payforRequest struct {
	XMLName       xml.Name `xml:"PayforRequest"`
	MbrId         string   `xml:"MbrId,omitempty"`
	MerchantId    string   `xml:"MerchantId,omitempty"`
	UserCode      string   `xml:"UserCode,omitempty"`
	UserPass      string   `xml:"UserPass,omitempty"`
	OrderId       string   `xml:"OrderId,omitempty"`
	OrgOrderId    string   `xml:"OrgOrderId,omitempty"`
	SecureType    string   `xml:"SecureType,omitempty"`
	TxnType       string   `xml:"TxnType,omitempty"`
	PurchAmount   string   `xml:"PurchAmount,omitempty"`
	Currency      string   `xml:"Currency,omitempty"`
	Lang          string   `xml:"Lang,omitempty"`
	RequestGuid   string   `xml:"RequestGuid,omitempty"`
	InstallmentCount string `xml:"InstallmentCount,omitempty"`
}

type payforResponse struct {
	XMLName        xml.Name `xml:"PayforResponse"`
	ProcReturnCode string   `xml:"ProcReturnCode"`
	ErrMsg         string   `xml:"ErrMsg"`
	TransId        string   `xml:"TransId"`
	OrderId        string   `xml:"OrderId"`
	OrgOrderId     string   `xml:"OrgOrderId"`
	AuthCode       string   `xml:"AuthCode"`
	HostRefNum     string   `xml:"HostRefNum"`
	PurchAmount    string   `xml:"PurchAmount"`
	Currency       string   `xml:"Currency"`
	InstallmentCount string `xml:"InstallmentCount"`
	TxnType        string   `xml:"TxnType"`
	CardMask       string   `xml:"CardMask"`
	InsertDatetime string   `xml:"InsertDatetime"`
}

var xmlWhitespace = regexp.MustCompile(`\r\n\s*`)

func encodePayfor(req payforRequest) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("payfor: encode xml: %w", err)
	}
	return buf.Bytes(), nil
}

func decodePayfor(body []byte) (payforResponse, map[string]string, error) {
	normalized := xmlWhitespace.ReplaceAll(body, []byte(""))
	var resp payforResponse
	if err := xml.Unmarshal(normalized, &resp); err != nil {
		return resp, nil, fmt.Errorf("payfor: decode xml: %w", err)
	}
	return resp, flattenXML(normalized), nil
}

func flattenXML(body []byte) map[string]string {
	dec := xml.NewDecoder(bytes.NewReader(body))
	out := make(map[string]string)
	var stack []string
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
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
			if text != "" && len(stack) > 0 {
				out[stack[len(stack)-1]] = text
			}
		}
	}
	return out
}
