package tray

import "testing"

func TestValidateWebURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		ok   bool
	}{
		{"http", "http://127.0.0.1:8765", true},
		{"https", "https://example.com/path?q=1", true},
		{"uppercase scheme", "HTTPS://example.com", true},
		{"surrounding space", "  http://localhost:8765/  ", true},
		{"file", "file:///C:/Windows/System32/calc.exe", false},
		{"javascript", "javascript:alert(1)", false},
		{"ms-settings", "ms-settings:network", false},
		{"bare path", `C:\Windows\System32\calc.exe`, false},
		{"no host", "http://", false},
		{"empty", "", false},
		{"malformed", "http://[::1", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ValidateWebURL(c.in)
			if (err == nil) != c.ok {
				t.Fatalf("ValidateWebURL(%q) err=%v, want ok=%v", c.in, err, c.ok)
			}
		})
	}
}
