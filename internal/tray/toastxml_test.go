package tray

import (
	"encoding/xml"
	"strings"
	"testing"
)

type toastText struct {
	Placement string `xml:"placement,attr"`
	Value     string `xml:",chardata"`
}

type toastDoc struct {
	ActivationType string      `xml:"activationType,attr"`
	Launch         string      `xml:"launch,attr"`
	Texts          []toastText `xml:"visual>binding>text"`
	Image          struct {
		Placement string `xml:"placement,attr"`
		Src       string `xml:"src,attr"`
	} `xml:"visual>binding>image"`
	Audio struct {
		Silent string `xml:"silent,attr"`
	} `xml:"audio"`
	Actions []struct {
		Content   string `xml:"content,attr"`
		Arguments string `xml:"arguments,attr"`
		Type      string `xml:"activationType,attr"`
	} `xml:"actions>action"`
}

func parseToast(t *testing.T, out string) toastDoc {
	t.Helper()
	var doc toastDoc
	if err := xml.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("output is not well-formed XML: %v\n%s", err, out)
	}
	return doc
}

func TestBuildToastXMLEscapes(t *testing.T) {
	title := `A & B <script>"x" 'y'`
	body := "line1\nline2 </text><text>injected"
	out := BuildToastXML(ToastContent{Title: title, Body: body})
	doc := parseToast(t, out)
	if len(doc.Texts) != 2 {
		t.Fatalf("expected 2 text nodes, got %d: %v", len(doc.Texts), doc.Texts)
	}
	if doc.Texts[0].Value != title || doc.Texts[1].Value != body {
		t.Errorf("roundtrip mismatch: %+v", doc.Texts)
	}
	if strings.Contains(out, "<script>") || strings.Contains(out, "<text>injected") {
		t.Errorf("unescaped content in output: %s", out)
	}
	if doc.Audio.Silent != "true" {
		t.Error("toast should be silent; the app plays its own sound")
	}
}

func TestBuildToastXMLUnicodeAndInvalidChars(t *testing.T) {
	doc := parseToast(t, BuildToastXML(ToastContent{Title: "انقطع الاتصال 🔴", Body: "bad\x00char"}))
	if doc.Texts[0].Value != "انقطع الاتصال 🔴" {
		t.Errorf("unicode title mangled: %q", doc.Texts[0].Value)
	}
	if strings.ContainsRune(doc.Texts[1].Value, 0) {
		t.Errorf("NUL survived in body")
	}
}

func TestBuildToastXMLEmpty(t *testing.T) {
	parseToast(t, BuildToastXML(ToastContent{}))
}

func TestBuildToastXMLRich(t *testing.T) {
	launch := "http://127.0.0.1:8765/?session=abc&x=\"1\""
	doc := parseToast(t, BuildToastXML(ToastContent{
		Title:       "Down",
		Body:        "Failed: DNS",
		Attribution: "Internet Monitor · 14:32",
		IconPath:    `C:\Users\A B\AppData\Local\InternetMonitor\status-down.png`,
		ActionLabel: `Open "dashboard"`,
		ActionURL:   launch,
	}))
	if doc.ActivationType != "protocol" || doc.Launch != launch {
		t.Errorf("launch = %q %q", doc.ActivationType, doc.Launch)
	}
	if len(doc.Texts) != 3 || doc.Texts[2].Placement != "attribution" {
		t.Errorf("texts = %+v", doc.Texts)
	}
	if doc.Image.Placement != "appLogoOverride" || doc.Image.Src != "file:///C:/Users/A%20B/AppData/Local/InternetMonitor/status-down.png" {
		t.Errorf("image = %+v", doc.Image)
	}
	if len(doc.Actions) != 1 || doc.Actions[0].Content != `Open "dashboard"` || doc.Actions[0].Arguments != doc.Launch || doc.Actions[0].Type != "protocol" {
		t.Errorf("actions = %+v", doc.Actions)
	}
}

func TestBuildToastXMLDropsUnsafeActions(t *testing.T) {
	for _, bad := range []string{"javascript:alert(1)", "file:///C:/Windows/system32/calc.exe", "ms-settings:", "http://"} {
		out := BuildToastXML(ToastContent{Title: "t", ActionLabel: "Open", ActionURL: bad})
		doc := parseToast(t, out)
		if doc.Launch != "" || len(doc.Actions) != 0 {
			t.Errorf("%q produced an action: %s", bad, out)
		}
	}
}
