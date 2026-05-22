package estpos

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

type cc5Request struct {
	XMLName                   xml.Name  `xml:"CC5Request"`
	Name                      string    `xml:"Name,omitempty"`
	Password                  string    `xml:"Password,omitempty"`
	ClientID                  string    `xml:"ClientId,omitempty"`
	Type                      string    `xml:"Type,omitempty"`
	OrderID                   string    `xml:"OrderId,omitempty"`
	Total                     string    `xml:"Total,omitempty"`
	Currency                  string    `xml:"Currency,omitempty"`
	Taksit                    string    `xml:"Taksit,omitempty"`
	IPAddress                 string    `xml:"IPAddress,omitempty"`
	Number                    string    `xml:"Number,omitempty"`
	PayerTxnID                string    `xml:"PayerTxnId,omitempty"`
	PayerSecurityLevel        string    `xml:"PayerSecurityLevel,omitempty"`
	PayerAuthenticationCode   string    `xml:"PayerAuthenticationCode,omitempty"`
	Mode                      string    `xml:"Mode,omitempty"`
	Extra                     *cc5Extra `xml:"Extra,omitempty"`
}

type cc5Extra struct {
	OrderStatus string `xml:"ORDERSTATUS,omitempty"`
}

type cc5Response struct {
	XMLName        xml.Name `xml:"CC5Response"`
	ProcReturnCode string   `xml:"ProcReturnCode"`
	ErrMsg         string   `xml:"ErrMsg"`
	OrderID        string   `xml:"OrderId"`
	TransID        string   `xml:"TransId"`
	AuthCode       string   `xml:"AuthCode"`
	HostRefNum     string   `xml:"HostRefNum"`
	GroupID        string   `xml:"GroupId"`
	Total          string   `xml:"Total"`
	Response       string   `xml:"Response"`
}

func encodeRequest(req cc5Request) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("estpos: encode xml: %w", err)
	}
	return buf.Bytes(), nil
}

func decodeResponse(body []byte) (cc5Response, map[string]string, error) {
	var resp cc5Response
	dec := xml.NewDecoder(bytes.NewReader(body))
	dec.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	if err := dec.Decode(&resp); err != nil {
		return resp, nil, fmt.Errorf("estpos: decode xml: %w", err)
	}
	return resp, flattenXML(body), nil
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
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			key := stack[len(stack)-1]
			val := string(bytes.TrimSpace(el))
			if val != "" {
				out[key] = val
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return out
}
