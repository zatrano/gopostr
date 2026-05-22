package payflex

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

type enrollmentResponse struct {
	XMLName      xml.Name `xml:"IPaySecure"`
	Message      enrollmentMessage `xml:"Message"`
	ErrorMessage string `xml:"ErrorMessage"`
	MessageErrorCode string `xml:"MessageErrorCode"`
}

type enrollmentMessage struct {
	VERes enrollmentVERes `xml:"VERes"`
}

type enrollmentVERes struct {
	Status string `xml:"Status"`
	PaReq  string `xml:"PaReq"`
	TermUrl string `xml:"TermUrl"`
	MD     string `xml:"MD"`
	ACSUrl string `xml:"ACSUrl"`
}

type vposRequest struct {
	XMLName                 xml.Name `xml:"VposRequest"`
	MerchantId              string   `xml:"MerchantId"`
	Password                string   `xml:"Password"`
	TerminalNo              string   `xml:"TerminalNo"`
	TransactionType         string   `xml:"TransactionType"`
	TransactionId           string   `xml:"TransactionId,omitempty"`
	CurrencyAmount          string   `xml:"CurrencyAmount,omitempty"`
	CurrencyCode            string   `xml:"CurrencyCode,omitempty"`
	ECI                     string   `xml:"ECI,omitempty"`
	CAVV                    string   `xml:"CAVV,omitempty"`
	MpiTransactionId        string   `xml:"MpiTransactionId,omitempty"`
	OrderId                 string   `xml:"OrderId,omitempty"`
	ClientIp                string   `xml:"ClientIp,omitempty"`
	TransactionDeviceSource string   `xml:"TransactionDeviceSource,omitempty"`
	CardHoldersName         string   `xml:"CardHoldersName,omitempty"`
	Cvv                     string   `xml:"Cvv,omitempty"`
	Pan                     string   `xml:"Pan,omitempty"`
	Expiry                  string   `xml:"Expiry,omitempty"`
	NumberOfInstallments    string   `xml:"NumberOfInstallments,omitempty"`
	ReferenceTransactionId  string   `xml:"ReferenceTransactionId,omitempty"`
}

type vposResponse struct {
	XMLName         xml.Name `xml:"VposResponse"`
	ResultCode      string   `xml:"ResultCode"`
	ResultDetail    string   `xml:"ResultDetail"`
	TransactionId   string   `xml:"TransactionId"`
	OrderId         string   `xml:"OrderId"`
	AuthCode        string   `xml:"AuthCode"`
	Rrn             string   `xml:"Rrn"`
	TransactionType string   `xml:"TransactionType"`
	CurrencyCode    string   `xml:"CurrencyCode"`
}

type searchRequest struct {
	XMLName             xml.Name `xml:"SearchRequest"`
	MerchantCriteria    merchantCriteria `xml:"MerchantCriteria"`
	TransactionCriteria transactionCriteria `xml:"TransactionCriteria"`
}

type merchantCriteria struct {
	HostMerchantId   string `xml:"HostMerchantId"`
	MerchantPassword string `xml:"MerchantPassword"`
}

type transactionCriteria struct {
	TransactionId string `xml:"TransactionId"`
	OrderId       string `xml:"OrderId"`
	AuthCode      string `xml:"AuthCode"`
}

type searchResponse struct {
	XMLName     xml.Name `xml:"SearchResponse"`
	ResponseInfo struct {
		ResponseCode    string `xml:"ResponseCode"`
		ResponseMessage string `xml:"ResponseMessage"`
	} `xml:"ResponseInfo"`
	TransactionSearchResultInfo struct {
		TransactionSearchResultInfo struct {
			OrderId         string `xml:"OrderId"`
			TransactionId   string `xml:"TransactionId"`
			AuthCode        string `xml:"AuthCode"`
			Rrn             string `xml:"Rrn"`
			ResultCode      string `xml:"ResultCode"`
			ResponseMessage string `xml:"ResponseMessage"`
		} `xml:"TransactionSearchResultInfo"`
	} `xml:"TransactionSearchResultInfo"`
}

func encodeVpos(req vposRequest) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("payflex: encode vpos: %w", err)
	}
	return buf.Bytes(), nil
}

func encodeSearch(req searchRequest) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("payflex: encode search: %w", err)
	}
	return buf.Bytes(), nil
}

func decodeEnrollment(body []byte) (enrollmentVERes, string, error) {
	var resp enrollmentResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return enrollmentVERes{}, "", fmt.Errorf("payflex: decode enrollment: %w", err)
	}
	if resp.ErrorMessage != "" {
		return enrollmentVERes{}, resp.ErrorMessage, fmt.Errorf("payflex: enrollment: %s", resp.ErrorMessage)
	}
	return resp.Message.VERes, "", nil
}

func decodeVpos(body []byte) (vposResponse, map[string]string, error) {
	var resp vposResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return resp, nil, fmt.Errorf("payflex: decode vpos: %w", err)
	}
	return resp, flattenXML(body), nil
}

func decodeSearch(body []byte) (searchResponse, map[string]string, error) {
	var resp searchResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return resp, nil, fmt.Errorf("payflex: decode search: %w", err)
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
			if text != "" && len(stack) > 0 {
				out[stack[len(stack)-1]] = text
			}
		}
	}
	return out
}
