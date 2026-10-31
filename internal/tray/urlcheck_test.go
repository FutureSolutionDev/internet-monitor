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

func TestValidateExternalURL(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"https://github.com/Khaymah/internet-monitor/issues", true},
		{"https://wa.me/201069017890", true},
		{"https://discord.gg/7Vhx6fw4qj", true},
		{"https://chat.whatsapp.com/Jt5KHILgpCMLp0dtx7hGd0", true},
		{"https://paypal.me/example", true},
		{"https://www.paypal.com/donate/?hosted_button_id=X", true},
		{"https://ipn.eg/S/example/instapay/abc", true},
		{"https://vf.eg/vfcash?id=mt&qrId=example", true},
		{"http://vf.eg/vfcash", false},
		{"https://paypal.me.evil.example/", false},
		{"http://github.com/", false},
		{"https://github.com.evil.example/", false},
		{"https://user@github.com/", false},
		{"https://github.com:8443/", false},
		{"https://example.com/", false},
		{"mailto:someone@example.com", false},
		{"javascript:alert(1)", false},
	}
	for _, c := range cases {
		if _, err := ValidateExternalURL(c.in); (err == nil) != c.ok {
			t.Errorf("ValidateExternalURL(%q) err=%v, want ok=%v", c.in, err, c.ok)
		}
	}
}
