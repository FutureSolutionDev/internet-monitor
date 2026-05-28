package updater

import "testing"

func TestParseChecksums(t *testing.T) {
	data := []byte(
		"aaa111  internet-monitor-windows.exe\n" +
			"bbb222 *internet-monitor-linux\n" +
			"ccc333  internet-monitor-macos-arm64\n")

	if got := parseChecksums(data, "internet-monitor-windows.exe"); got != "aaa111" {
		t.Errorf("windows hash = %q, want aaa111", got)
	}
	if got := parseChecksums(data, "internet-monitor-linux"); got != "bbb222" {
		t.Errorf("linux hash (binary marker) = %q, want bbb222", got)
	}
	if got := parseChecksums(data, "not-listed"); got != "" {
		t.Errorf("missing asset = %q, want empty", got)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.2.0", "v1.1.9", 1},
		{"1.2.0", "v1.2.0", 0},
		{"v1.0.0", "v1.0.1", -1},
		{"v2.0.0", "v1.9.9", 1},
		{"v1.2.3", "dev", 0}, // dev build is always "current"
		{"v1.2.3", "", 0},
		{"v1.10.0", "v1.9.0", 1}, // numeric, not lexicographic
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
