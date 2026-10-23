package tray

import (
	"bytes"
	"encoding/xml"
	"strings"
)

func escapeXML(s string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return ""
	}
	return buf.String()
}

func BuildToastXML(title, body string) string {
	var sb strings.Builder
	sb.WriteString(`<toast><visual><binding template="ToastGeneric"><text>`)
	sb.WriteString(escapeXML(title))
	sb.WriteString(`</text><text>`)
	sb.WriteString(escapeXML(body))
	sb.WriteString(`</text></binding></visual></toast>`)
	return sb.String()
}
