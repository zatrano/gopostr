package vakifkatilim

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func wrapContract(inner string) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="ISO-8859-1"?>`)
	b.WriteString("<VPosMessageContract>")
	b.WriteString(inner)
	b.WriteString("</VPosMessageContract>")
	return []byte(b.String())
}

func writeField(b *strings.Builder, tag, val string) {
	if val != "" {
		fmt.Fprintf(b, "<%s>%s</%s>", tag, xmlEscape(val), tag)
	}
}

type enrollmentFields struct {
	MerchantID, CustomerID, UserName, SubMerchantID string
	HashPassword, HashData, TxSecurity              string
	Installment, Amount, DisplayAmount              int
	FECCurrency, MerchantOrderID, OkURL, FailURL     string
	CardHolder, CardNumber, CardMonth, CardYear, CVV string
}

func encodeEnrollment(f enrollmentFields) []byte {
	var b strings.Builder
	writeField(&b, "MerchantId", f.MerchantID)
	writeField(&b, "CustomerId", f.CustomerID)
	writeField(&b, "UserName", f.UserName)
	writeField(&b, "SubMerchantId", f.SubMerchantID)
	writeField(&b, "APIVersion", apiVersion)
	writeField(&b, "HashPassword", f.HashPassword)
	writeField(&b, "HashData", f.HashData)
	writeField(&b, "TransactionSecurity", f.TxSecurity)
	writeField(&b, "InstallmentCount", fmt.Sprintf("%d", f.Installment))
	writeField(&b, "Amount", fmt.Sprintf("%d", f.Amount))
	writeField(&b, "DisplayAmount", fmt.Sprintf("%d", f.DisplayAmount))
	writeField(&b, "FECCurrencyCode", f.FECCurrency)
	writeField(&b, "MerchantOrderId", f.MerchantOrderID)
	writeField(&b, "OkUrl", f.OkURL)
	writeField(&b, "FailUrl", f.FailURL)
	if f.CardNumber != "" {
		writeField(&b, "CardHolderName", f.CardHolder)
		writeField(&b, "CardNumber", f.CardNumber)
		writeField(&b, "CardExpireDateMonth", f.CardMonth)
		writeField(&b, "CardExpireDateYear", f.CardYear)
		writeField(&b, "CardCVV2", f.CVV)
	}
	return wrapContract(b.String())
}

type provisionFields struct {
	MerchantID, CustomerID, UserName, SubMerchantID string
	OkURL, FailURL, HashData, MD                     string
	Installment, Amount                              int
	MerchantOrderID, TxSecurity                      string
}

func encodeProvision(f provisionFields) []byte {
	var b strings.Builder
	writeField(&b, "MerchantId", f.MerchantID)
	writeField(&b, "CustomerId", f.CustomerID)
	writeField(&b, "UserName", f.UserName)
	writeField(&b, "SubMerchantId", f.SubMerchantID)
	writeField(&b, "OkUrl", f.OkURL)
	writeField(&b, "FailUrl", f.FailURL)
	writeField(&b, "HashData", f.HashData)
	writeField(&b, "APIVersion", apiVersion)
	b.WriteString("<AdditionalData><AdditionalDataList><VPosAdditionalData>")
	writeField(&b, "Key", "MD")
	writeField(&b, "Data", f.MD)
	b.WriteString("</VPosAdditionalData></AdditionalDataList></AdditionalData>")
	writeField(&b, "InstallmentCount", fmt.Sprintf("%d", f.Installment))
	writeField(&b, "Amount", fmt.Sprintf("%d", f.Amount))
	writeField(&b, "MerchantOrderId", f.MerchantOrderID)
	writeField(&b, "TransactionSecurity", f.TxSecurity)
	return wrapContract(b.String())
}

func encodeCancelRefund(fields map[string]string) []byte {
	var b strings.Builder
	for _, key := range []string{
		"MerchantId", "CustomerId", "UserName", "SubMerchantId",
		"HashPassword", "MerchantOrderId", "OrderId", "PaymentType", "Amount", "HashData",
	} {
		writeField(&b, key, fields[key])
	}
	return wrapContract(b.String())
}

func encodeStatus(fields map[string]string) []byte {
	var b strings.Builder
	for _, key := range []string{
		"MerchantId", "CustomerId", "UserName", "SubMerchantId", "MerchantOrderId", "HashData",
	} {
		writeField(&b, key, fields[key])
	}
	return wrapContract(b.String())
}

func decodeXML(body []byte) (map[string]string, error) {
	body = bytes.ReplaceAll(body, []byte("&#x0;"), nil)
	body = bytes.ReplaceAll(body, []byte(` encoding="utf-16"`), nil)
	raw := flattenXML(body)
	if len(raw) == 0 {
		return nil, fmt.Errorf("vakifkatilim: boş veya geçersiz XML")
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
			out[stack[len(stack)-1]] = text
		}
	}
	return out
}
