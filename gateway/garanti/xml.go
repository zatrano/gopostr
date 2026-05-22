package garanti

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

const apiVersion = "512"

type gvpsRequest struct {
	XMLName     xml.Name         `xml:"GVPSRequest"`
	Mode        string           `xml:"Mode"`
	Version     string           `xml:"Version"`
	Terminal    gvpsTerminal     `xml:"Terminal"`
	Customer    *gvpsCustomer    `xml:"Customer,omitempty"`
	Order       gvpsOrder        `xml:"Order"`
	Transaction gvpsTransaction  `xml:"Transaction"`
}

type gvpsTerminal struct {
	ProvUserID string `xml:"ProvUserID"`
	UserID     string `xml:"UserID"`
	HashData   string `xml:"HashData"`
	ID         string `xml:"ID"`
	MerchantID string `xml:"MerchantID"`
}

type gvpsCustomer struct {
	IPAddress string `xml:"IPAddress"`
}

type gvpsOrder struct {
	OrderID string `xml:"OrderID"`
}

type gvpsTransaction struct {
	Type                  string         `xml:"Type"`
	InstallmentCnt        string         `xml:"InstallmentCnt,omitempty"`
	Amount                string         `xml:"Amount,omitempty"`
	CurrencyCode          string         `xml:"CurrencyCode,omitempty"`
	CardholderPresentCode string         `xml:"CardholderPresentCode,omitempty"`
	MotoInd               string         `xml:"MotoInd,omitempty"`
	OriginalRetrefNum     string         `xml:"OriginalRetrefNum,omitempty"`
	Secure3D              *gvpsSecure3D   `xml:"Secure3D,omitempty"`
	Response              *gvpsTxResponse  `xml:"Response,omitempty"`
	AuthCode              string         `xml:"AuthCode,omitempty"`
	RetrefNum             string         `xml:"RetrefNum,omitempty"`
}

type gvpsSecure3D struct {
	AuthenticationCode string `xml:"AuthenticationCode"`
	SecurityLevel    string `xml:"SecurityLevel"`
	TxnID            string `xml:"TxnID"`
	Md               string `xml:"Md"`
}

type gvpsResponse struct {
	XMLName     xml.Name    `xml:"GVPSResponse"`
	Order       gvpsRespOrder `xml:"Order"`
	Transaction gvpsRespTx    `xml:"Transaction"`
}

type gvpsRespOrder struct {
	OrderID string `xml:"OrderID"`
}

type gvpsRespTx struct {
	Response  gvpsTxResponse `xml:"Response"`
	AuthCode  string         `xml:"AuthCode"`
	RetrefNum string         `xml:"RetrefNum"`
}

type gvpsTxResponse struct {
	Code     string `xml:"Code"`
	ReasonCode string `xml:"ReasonCode"`
	ErrorMsg string `xml:"ErrorMsg"`
	Message  string `xml:"Message"`
}

func encodeGVPS(req gvpsRequest) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("garanti: encode xml: %w", err)
	}
	return buf.Bytes(), nil
}

func decodeGVPS(body []byte) (gvpsResponse, map[string]string, error) {
	var resp gvpsResponse
	dec := xml.NewDecoder(bytes.NewReader(body))
	if err := dec.Decode(&resp); err != nil {
		return resp, nil, fmt.Errorf("garanti: decode xml: %w", err)
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
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			text := string(bytes.TrimSpace(el))
			if text == "" || len(stack) == 0 {
				continue
			}
			out[stack[len(stack)-1]] = text
		}
	}
	return out
}
