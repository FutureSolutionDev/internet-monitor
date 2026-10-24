package i18n

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func TestEveryLocaleHasMetaAndGoKeys(t *testing.T) {
	used := map[string]bool{}
	call := regexp.MustCompile(`i18n\.T\([^,]+,\s*"([^"]+)"`)
	for _, root := range []string{"../../internal", "../../cmd"} {
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return err
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			for _, m := range call.FindAllStringSubmatch(string(data), -1) {
				used[m[1]] = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(used) < 10 {
		t.Fatalf("found only %d i18n keys in Go code; the scan is broken", len(used))
	}
	for code, l := range locales {
		if l.Meta.Name == "" || (l.Meta.Dir != "ltr" && l.Meta.Dir != "rtl") || l.Meta.Intl == "" {
			t.Errorf("%s: incomplete meta %+v", code, l.Meta)
		}
		for key := range used {
			if strings.TrimSpace(l.Messages[key]) == "" {
				t.Errorf("%s: missing %q used by Go code", code, key)
			}
		}
	}
}
