package tray

import (
	"encoding/xml"
	"strings"
	"testing"
)

type toastDoc struct {
	Texts []string `xml:"visual>binding>text"`
}

func TestBuildToastXMLEscapes(t *testing.T) {
	title := `A & B <script>"x" 'y'`
	body := "line1\nline2 </text><text>injected"
	out := BuildToastXML(title, body)

	var doc toastDoc
	if err := xml.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("output is not well-formed XML: %v\n%s", err, out)
	}
	if len(doc.Texts) != 2 {
		t.Fatalf("expected 2 text nodes, got %d: %v", len(doc.Texts), doc.Texts)
	}
	if doc.Texts[0] != title {
		t.Errorf("title roundtrip mismatch: %q", doc.Texts[0])
	}
	if doc.Texts[1] != body {
		t.Errorf("body roundtrip mismatch: %q", doc.Texts[1])
	}
	if strings.Contains(out, "<script>") || strings.Contains(out, "<text>injected") {
		t.Errorf("unescaped content in output: %s", out)
	}
}

func TestBuildToastXMLUnicodeAndInvalidChars(t *testing.T) {
	out := BuildToastXML("انقطع الاتصال 🔴", "bad\x00char")
	var doc toastDoc
	if err := xml.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("output is not well-formed XML: %v\n%s", err, out)
	}
	if doc.Texts[0] != "انقطع الاتصال 🔴" {
		t.Errorf("unicode title mangled: %q", doc.Texts[0])
	}
	if strings.ContainsRune(doc.Texts[1], 0) {
		t.Errorf("NUL survived in body")
	}
}

func TestBuildToastXMLEmpty(t *testing.T) {
	var doc toastDoc
	if err := xml.Unmarshal([]byte(BuildToastXML("", "")), &doc); err != nil {
		t.Fatalf("empty input produced invalid XML: %v", err)
	}
}
