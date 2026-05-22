package posnet

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const xmlEncoding = "ISO-8859-9"

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func writeXML(buf *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(buf, "<%s>%s</%s>", name, xmlEscape(value), name)
}

func wrapPosnetRequest(inner string) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="` + xmlEncoding + `"?>`)
	b.WriteString("<posnetRequest>")
	b.WriteString(inner)
	b.WriteString("</posnetRequest>")
	return []byte(b.String())
}

func encodeEnrollment(mid, tid, posnetID string, oos oosRequestFields) []byte {
	var b strings.Builder
	writeXML(&b, "mid", mid)
	writeXML(&b, "tid", tid)
	b.WriteString("<oosRequestData>")
	writeXML(&b, "posnetid", posnetID)
	writeXML(&b, "ccno", oos.CCNo)
	writeXML(&b, "expDate", oos.ExpDate)
	writeXML(&b, "cvc", oos.CVC)
	writeXML(&b, "amount", fmt.Sprintf("%d", oos.AmountKurus))
	writeXML(&b, "currencyCode", oos.CurrencyCode)
	writeXML(&b, "installment", oos.Installment)
	writeXML(&b, "XID", oos.XID)
	writeXML(&b, "cardHolderName", oos.CardHolderName)
	writeXML(&b, "tranType", oos.TranType)
	b.WriteString("</oosRequestData>")
	return wrapPosnetRequest(b.String())
}

type oosRequestFields struct {
	CCNo, ExpDate, CVC, XID, CardHolderName, TranType, CurrencyCode, Installment string
	AmountKurus                                                                  int
}

func encodeResolve(mid, tid, mac string, bank, merchant, sign string) []byte {
	var b strings.Builder
	writeXML(&b, "mid", mid)
	writeXML(&b, "tid", tid)
	b.WriteString("<oosResolveMerchantData>")
	writeXML(&b, "bankData", bank)
	writeXML(&b, "merchantData", merchant)
	writeXML(&b, "sign", sign)
	writeXML(&b, "mac", mac)
	b.WriteString("</oosResolveMerchantData>")
	return wrapPosnetRequest(b.String())
}

func encodeTran(mid, tid, mac string, bank, merchant, sign string) []byte {
	var b strings.Builder
	writeXML(&b, "mid", mid)
	writeXML(&b, "tid", tid)
	b.WriteString("<oosTranData>")
	writeXML(&b, "bankData", bank)
	writeXML(&b, "merchantData", merchant)
	writeXML(&b, "sign", sign)
	writeXML(&b, "wpAmount", "0")
	writeXML(&b, "mac", mac)
	b.WriteString("</oosTranData>")
	return wrapPosnetRequest(b.String())
}

func encodeCancel(mid, tid, hostLogKey, orderID, authCode string) []byte {
	var b strings.Builder
	writeXML(&b, "mid", mid)
	writeXML(&b, "tid", tid)
	writeXML(&b, "tranDateRequired", "1")
	b.WriteString("<reverse>")
	writeXML(&b, "transaction", "sale")
	if authCode != "" {
		writeXML(&b, "authCode", authCode)
	}
	if hostLogKey != "" {
		writeXML(&b, "hostLogKey", hostLogKey)
	} else {
		writeXML(&b, "orderID", orderID)
	}
	b.WriteString("</reverse>")
	return wrapPosnetRequest(b.String())
}

func encodeRefund(mid, tid string, amountKurus int, currency, hostLogKey, orderID string) []byte {
	var b strings.Builder
	writeXML(&b, "mid", mid)
	writeXML(&b, "tid", tid)
	writeXML(&b, "tranDateRequired", "1")
	b.WriteString("<return>")
	writeXML(&b, "amount", fmt.Sprintf("%d", amountKurus))
	writeXML(&b, "currencyCode", currency)
	if hostLogKey != "" {
		writeXML(&b, "hostLogKey", hostLogKey)
	} else {
		writeXML(&b, "orderID", orderID)
	}
	b.WriteString("</return>")
	return wrapPosnetRequest(b.String())
}

func encodeStatus(mid, tid, orderID string) []byte {
	var b strings.Builder
	writeXML(&b, "mid", mid)
	writeXML(&b, "tid", tid)
	b.WriteString("<agreement>")
	writeXML(&b, "orderID", orderID)
	b.WriteString("</agreement>")
	return wrapPosnetRequest(b.String())
}

func decodeResponse(body []byte) (map[string]string, error) {
	raw := flattenXML(body)
	if len(raw) == 0 {
		return nil, fmt.Errorf("posnet: boş yanıt")
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

func approvedOK(raw map[string]string) bool {
	return payloadVal(raw, "approved") == procSuccess
}

func resolveMACFields(raw map[string]string) map[string]string {
	return map[string]string{
		"mdStatus": payloadVal(raw, "mdStatus"),
		"xid":      payloadVal(raw, "xid"),
		"amount":   payloadVal(raw, "amount"),
		"currency": payloadVal(raw, "currency"),
		"mac":      payloadVal(raw, "mac"),
	}
}
