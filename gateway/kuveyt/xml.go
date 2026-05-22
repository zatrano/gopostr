package kuveyt

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"
)

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func wrapKuveytMessage(inner string) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="ISO-8859-1"?>`)
	b.WriteString("<KuveytTurkVPosMessage>")
	b.WriteString(inner)
	b.WriteString("</KuveytTurkVPosMessage>")
	return []byte(b.String())
}

type enrollmentParams struct {
	APIVersion, HashData, TxnType, TxSecurity string
	Installment, Amount, DisplayAmount        int
	CurrencyCode, MerchantOrderID             string
	OkURL, FailURL, ClientIP                  string
	MerchantID, CustomerID, UserName          string
	CardHolder, CardType, CardNumber          string
	CardExpMonth, CardExpYear, CardCVV        string
}

func encodeEnrollment(p enrollmentParams) []byte {
	var b strings.Builder
	write := func(tag, val string) {
		if val != "" {
			fmt.Fprintf(&b, "<%s>%s</%s>", tag, xmlEscape(val), tag)
		}
	}
	write("APIVersion", p.APIVersion)
	write("HashData", p.HashData)
	write("TransactionType", p.TxnType)
	write("TransactionSecurity", p.TxSecurity)
	write("InstallmentCount", fmt.Sprintf("%d", p.Installment))
	write("Amount", fmt.Sprintf("%d", p.Amount))
	write("DisplayAmount", fmt.Sprintf("%d", p.DisplayAmount))
	write("CurrencyCode", p.CurrencyCode)
	write("MerchantOrderId", p.MerchantOrderID)
	write("OkUrl", p.OkURL)
	write("FailUrl", p.FailURL)
	write("MerchantId", p.MerchantID)
	write("CustomerId", p.CustomerID)
	write("UserName", p.UserName)
	b.WriteString("<DeviceData><ClientIP>" + xmlEscape(p.ClientIP) + "</ClientIP></DeviceData>")
	if p.CardNumber != "" {
		write("CardHolderName", p.CardHolder)
		write("CardType", p.CardType)
		write("CardNumber", p.CardNumber)
		write("CardExpireDateMonth", p.CardExpMonth)
		write("CardExpireDateYear", p.CardExpYear)
		write("CardCVV2", p.CardCVV)
	}
	return wrapKuveytMessage(b.String())
}

type provisionParams struct {
	APIVersion, HashData, TxnType, TxSecurity string
	Installment, Amount, DisplayAmount        int
	CurrencyCode, MerchantOrderID, MD         string
	ClientIP, MerchantID, CustomerID, UserName string
}

func encodeProvision(p provisionParams) []byte {
	var b strings.Builder
	write := func(tag, val string) {
		if val != "" {
			fmt.Fprintf(&b, "<%s>%s</%s>", tag, xmlEscape(val), tag)
		}
	}
	write("APIVersion", p.APIVersion)
	write("HashData", p.HashData)
	write("CustomerIPAddress", p.ClientIP)
	b.WriteString("<KuveytTurkVPosAdditionalData><AdditionalData>")
	write("Key", "MD")
	write("Data", p.MD)
	b.WriteString("</AdditionalData></KuveytTurkVPosAdditionalData>")
	write("TransactionType", p.TxnType)
	write("InstallmentCount", fmt.Sprintf("%d", p.Installment))
	write("Amount", fmt.Sprintf("%d", p.Amount))
	write("DisplayAmount", fmt.Sprintf("%d", p.DisplayAmount))
	write("CurrencyCode", p.CurrencyCode)
	write("MerchantOrderId", p.MerchantOrderID)
	write("TransactionSecurity", p.TxSecurity)
	write("MerchantId", p.MerchantID)
	write("CustomerId", p.CustomerID)
	write("UserName", p.UserName)
	write("MOTO", "0")
	return wrapKuveytMessage(b.String())
}

func decodeXML(body []byte) (map[string]string, error) {
	raw := flattenXML(body)
	if len(raw) == 0 {
		return nil, fmt.Errorf("kuveyt: boş veya geçersiz XML")
	}
	return raw, nil
}

func flattenXML(body []byte) map[string]string {
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
			if text == "" || len(stack) == 0 {
				continue
			}
			key := stack[len(stack)-1]
			out[key] = text
		}
	}
	return out
}

func vposSubMap(raw map[string]string, prefix string) map[string]string {
	sub := make(map[string]string)
	p := prefix + "."
	for k, v := range raw {
		if strings.HasPrefix(k, p) {
			sub[strings.TrimPrefix(k, p)] = v
		}
	}
	if len(sub) > 0 {
		return sub
	}
	keys := []string{
		"InstallmentCount", "Amount", "CurrencyCode", "MerchantOrderId",
		"TransactionSecurity", "CardNumber", "BatchID", "TransactionType",
	}
	for _, k := range keys {
		if v := payloadVal(raw, k); v != "" {
			sub[k] = v
		}
	}
	return sub
}

const soapNS = `http://schemas.xmlsoap.org/soap/envelope/`
const serNS = `http://boa.net/BOA.Integration.VirtualPos/Service`

type soapVPos struct {
	APIVersion, HashData, TxnType, TxSecurity, CardType string
	Installment                                         string
	Amount, DisplayAmount, CancelAmount                 int
	CurrencyCode, MerchantOrderID                       string
}

func encodeSOAPCancel(req soapCancelRequest) []byte {
	req.VPos.TxnType = "SaleReversal"
	return encodeSOAPAction("SaleReversal", buildCancelRefundBody(req))
}

func encodeSOAPRefund(req soapCancelRequest, partial bool) []byte {
	action := "DrawBack"
	if partial {
		action = "PartialDrawback"
	}
	req.VPos.TxnType = action
	return encodeSOAPAction(action, buildCancelRefundBody(req))
}

type soapCancelRequest struct {
	CustomerID, MerchantID string
	Amount                 int
	OrderID                string
	BankOrderID            string
	RRN, Stan, Provision   string
	VPos                   soapVPos
}

func buildCancelRefundBody(req soapCancelRequest) string {
	v := req.VPos
	var b strings.Builder
	fmt.Fprintf(&b, `<ser:IsFromExternalNetwork>true</ser:IsFromExternalNetwork>`)
	fmt.Fprintf(&b, `<ser:BusinessKey>0</ser:BusinessKey>`)
	fmt.Fprintf(&b, `<ser:ResourceId>0</ser:ResourceId>`)
	fmt.Fprintf(&b, `<ser:ActionId>0</ser:ActionId>`)
	fmt.Fprintf(&b, `<ser:LanguageId>0</ser:LanguageId>`)
	fmt.Fprintf(&b, `<ser:CustomerId>%s</ser:CustomerId>`, xmlEscape(req.CustomerID))
	fmt.Fprintf(&b, `<ser:MailOrTelephoneOrder>true</ser:MailOrTelephoneOrder>`)
	fmt.Fprintf(&b, `<ser:Amount>%d</ser:Amount>`, req.Amount)
	fmt.Fprintf(&b, `<ser:MerchantId>%s</ser:MerchantId>`, xmlEscape(req.MerchantID))
	fmt.Fprintf(&b, `<ser:OrderId>%s</ser:OrderId>`, xmlEscape(req.BankOrderID))
	fmt.Fprintf(&b, `<ser:RRN>%s</ser:RRN>`, xmlEscape(req.RRN))
	fmt.Fprintf(&b, `<ser:Stan>%s</ser:Stan>`, xmlEscape(req.Stan))
	fmt.Fprintf(&b, `<ser:ProvisionNumber>%s</ser:ProvisionNumber>`, xmlEscape(req.Provision))
	b.WriteString(`<ser:VPosMessage>`)
	appendVPosFields(&b, v)
	b.WriteString(`</ser:VPosMessage>`)
	return b.String()
}

func encodeSOAPStatus(req soapStatusRequest) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, `<ser:IsFromExternalNetwork>true</ser:IsFromExternalNetwork>`)
	fmt.Fprintf(&b, `<ser:BusinessKey>0</ser:BusinessKey>`)
	fmt.Fprintf(&b, `<ser:ResourceId>0</ser:ResourceId>`)
	fmt.Fprintf(&b, `<ser:ActionId>0</ser:ActionId>`)
	fmt.Fprintf(&b, `<ser:LanguageId>0</ser:LanguageId>`)
	fmt.Fprintf(&b, `<ser:CustomerId>%s</ser:CustomerId>`, xmlEscape(req.CustomerID))
	fmt.Fprintf(&b, `<ser:MailOrTelephoneOrder>true</ser:MailOrTelephoneOrder>`)
	fmt.Fprintf(&b, `<ser:Amount>0</ser:Amount>`)
	fmt.Fprintf(&b, `<ser:MerchantId>%s</ser:MerchantId>`, xmlEscape(req.MerchantID))
	fmt.Fprintf(&b, `<ser:MerchantOrderId>%s</ser:MerchantOrderId>`, xmlEscape(req.MerchantOrderID))
	fmt.Fprintf(&b, `<ser:OrderId>%s</ser:OrderId>`, xmlEscape(req.BankOrderID))
	fmt.Fprintf(&b, `<ser:StartDate>%s</ser:StartDate>`, req.StartDate)
	fmt.Fprintf(&b, `<ser:EndDate>%s</ser:EndDate>`, req.EndDate)
	fmt.Fprintf(&b, `<ser:TransactionType>0</ser:TransactionType>`)
	b.WriteString(`<ser:VPosMessage>`)
	appendVPosFields(&b, req.VPos)
	b.WriteString(`</ser:VPosMessage>`)
	return encodeSOAPAction("GetMerchantOrderDetail", b.String())
}

type soapStatusRequest struct {
	CustomerID, MerchantID, MerchantOrderID, BankOrderID string
	StartDate, EndDate                                   string
	VPos                                                 soapVPos
}

func appendVPosFields(b *strings.Builder, v soapVPos) {
	fmt.Fprintf(b, `<ser:APIVersion>%s</ser:APIVersion>`, apiVersion)
	fmt.Fprintf(b, `<ser:InstallmentMaturityCommisionFlag>0</ser:InstallmentMaturityCommisionFlag>`)
	fmt.Fprintf(b, `<ser:HashData>%s</ser:HashData>`, xmlEscape(v.HashData))
	fmt.Fprintf(b, `<ser:SubMerchantId>0</ser:SubMerchantId>`)
	fmt.Fprintf(b, `<ser:CardType>%s</ser:CardType>`, xmlEscape(v.CardType))
	fmt.Fprintf(b, `<ser:BatchID>0</ser:BatchID>`)
	fmt.Fprintf(b, `<ser:TransactionType>%s</ser:TransactionType>`, xmlEscape(v.TxnType))
	fmt.Fprintf(b, `<ser:InstallmentCount>%s</ser:InstallmentCount>`, v.Installment)
	fmt.Fprintf(b, `<ser:Amount>%d</ser:Amount>`, v.Amount)
	fmt.Fprintf(b, `<ser:DisplayAmount>%d</ser:DisplayAmount>`, v.DisplayAmount)
	fmt.Fprintf(b, `<ser:CancelAmount>%d</ser:CancelAmount>`, v.CancelAmount)
	fmt.Fprintf(b, `<ser:MerchantOrderId>%s</ser:MerchantOrderId>`, xmlEscape(v.MerchantOrderID))
	fmt.Fprintf(b, `<ser:CurrencyCode>%s</ser:CurrencyCode>`, xmlEscape(v.CurrencyCode))
	fmt.Fprintf(b, `<ser:FECAmount>0</ser:FECAmount>`)
	fmt.Fprintf(b, `<ser:QeryId>0</ser:QeryId>`)
	fmt.Fprintf(b, `<ser:DebtId>0</ser:DebtId>`)
	fmt.Fprintf(b, `<ser:SurchargeAmount>0</ser:SurchargeAmount>`)
	fmt.Fprintf(b, `<ser:SGKDebtAmount>0</ser:SGKDebtAmount>`)
	fmt.Fprintf(b, `<ser:TransactionSecurity>%s</ser:TransactionSecurity>`, xmlEscape(v.TxSecurity))
}

func encodeSOAPAction(action, inner string) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<soapenv:Envelope xmlns:soapenv="` + soapNS + `" xmlns:ser="` + serNS + `">`)
	b.WriteString(`<soapenv:Body><ser:` + action + `><ser:request>`)
	b.WriteString(inner)
	b.WriteString(`</ser:request></ser:` + action + `></soapenv:Body></soapenv:Envelope>`)
	return []byte(b.String())
}

func decodeSOAP(body []byte) (map[string]string, error) {
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
			return nil, fmt.Errorf("kuveyt: soap decode: %w", err)
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
			key := strings.Join(stack, ".")
			out[key] = text
		}
	}
	if strings.Contains(string(body), "Fault") {
		msg := payloadVal(out, "faultstring")
		if msg == "" {
			msg = "SOAP fault"
		}
		return out, fmt.Errorf("kuveyt: %s", msg)
	}
	return out, nil
}

func soapActionFor(txType string) string {
	return "http://boa.net/BOA.Integration.VirtualPos/Service/IVirtualPosService/" + txType
}

func defaultStatusDates() (start, end string) {
	now := time.Now()
	start = now.AddDate(0, -12, 0).Format("2006-01-02T15:04:05")
	end = now.Format("2006-01-02T15:04:05")
	return start, end
}
