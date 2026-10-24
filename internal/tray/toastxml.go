package tray

import (
	"bytes"
	"encoding/xml"
	"net/url"
	"path/filepath"
	"strings"
)

type ToastContent struct {
	Title       string
	Body        string
	Attribution string
	IconPath    string
	ActionLabel string
	ActionURL   string
}

func escapeXML(s string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return ""
	}
	return buf.String()
}

func escapeAttr(s string) string {
	return strings.NewReplacer(`"`, "&quot;", "'", "&apos;").Replace(escapeXML(s))
}

func fileURI(path string) string {
	u := url.URL{Scheme: "file", Path: "/" + strings.TrimPrefix(filepath.ToSlash(path), "/")}
	return u.String()
}

func BuildToastXML(c ToastContent) string {
	action := ""
	if c.ActionURL != "" {
		if valid, err := ValidateWebURL(c.ActionURL); err == nil {
			action = valid
		}
	}
	var sb strings.Builder
	if action != "" {
		sb.WriteString(`<toast activationType="protocol" launch="` + escapeAttr(action) + `">`)
	} else {
		sb.WriteString(`<toast>`)
	}
	sb.WriteString(`<visual><binding template="ToastGeneric">`)
	sb.WriteString(`<text>` + escapeXML(c.Title) + `</text>`)
	sb.WriteString(`<text>` + escapeXML(c.Body) + `</text>`)
	if c.Attribution != "" {
		sb.WriteString(`<text placement="attribution">` + escapeXML(c.Attribution) + `</text>`)
	}
	if c.IconPath != "" {
		sb.WriteString(`<image placement="appLogoOverride" hint-crop="circle" src="` + escapeAttr(fileURI(c.IconPath)) + `"/>`)
	}
	sb.WriteString(`</binding></visual><audio silent="true"/>`)
	if action != "" && c.ActionLabel != "" {
		sb.WriteString(`<actions><action activationType="protocol" content="` + escapeAttr(c.ActionLabel) + `" arguments="` + escapeAttr(action) + `"/></actions>`)
	}
	sb.WriteString(`</toast>`)
	return sb.String()
}
