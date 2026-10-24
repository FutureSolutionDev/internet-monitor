package i18n

import (
	"testing"
	"testing/fstest"
)

func TestShippedLocalesLoad(t *testing.T) {
	for _, code := range []string{"en", "ar"} {
		if !Has(code) {
			t.Fatalf("locale %q not embedded", code)
		}
	}
	if T("ar", "app_name", nil) == T("en", "app_name", nil) {
		t.Error("Arabic app name falls back to English")
	}
}

func TestTFallsBackAndFillsVars(t *testing.T) {
	if got := T("xx", "app_name", nil); got != "Internet Monitor" {
		t.Errorf("unknown language: %q", got)
	}
	if got := T("en", "no.such.key", nil); got != "no.such.key" {
		t.Errorf("missing key: %q", got)
	}
	if got := T("en", "version_current", Vars{"v": "1.2.3"}); got != "Version 1.2.3" {
		t.Errorf("vars: %q", got)
	}
	if got := T("en", "version_current", Vars{}); got != "Version {v}" {
		t.Errorf("unknown placeholder should stay: %q", got)
	}
}

func TestNormalize(t *testing.T) {
	if Normalize("ar") != "ar" || Normalize("") != "en" || Normalize("../x") != "en" {
		t.Error("Normalize accepted an unknown language")
	}
}

func TestLoadRejectsBrokenFiles(t *testing.T) {
	if _, err := load(fstest.MapFS{"locales/en.json": {Data: []byte("{")}}); err == nil {
		t.Error("broken JSON accepted")
	}
	if _, err := load(fstest.MapFS{"locales/ar.json": {Data: []byte(`{"meta":{},"messages":{}}`)}}); err == nil {
		t.Error("missing English accepted")
	}
}
