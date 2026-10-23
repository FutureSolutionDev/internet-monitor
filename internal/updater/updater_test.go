package updater

import (
	"strings"
	"testing"
)

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
		{"v1.2.3", "dev", 0},
		{"v1.2.3", "", 0},
		{"v1.10.0", "v1.9.0", 1},
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestTrustedURL(t *testing.T) {
	for _, u := range []string{
		"https://github.com/FutureSolutionDev/internet-monitor/releases/download/v1/x.exe",
		"https://objects.githubusercontent.com/abc",
	} {
		if err := trustedURL(u); err != nil {
			t.Errorf("%s rejected: %v", u, err)
		}
	}
	for _, u := range []string{
		"http://github.com/x",
		"https://evil.example/x.exe",
		"https://github.com.evil.example/x",
	} {
		if trustedURL(u) == nil {
			t.Errorf("%s accepted", u)
		}
	}
}

func TestPickAssetSkipsGUIForTray(t *testing.T) {
	names := []string{
		"internet-monitor-gui-windows.exe",
		"internet-monitor-gui-macos-arm64",
		"internet-monitor-gui-macos-intel",
		"internet-monitor-gui-linux",
		"internet-monitor-windows.exe",
		"internet-monitor-macos-arm64",
		"internet-monitor-macos-intel",
		"internet-monitor-linux",
	}
	assets := make([]ghAsset, 0, len(names))
	for _, n := range names {
		assets = append(assets, ghAsset{Name: n, BrowserDownloadURL: "https://github.com/x/y/releases/download/v1/" + n})
	}
	got := pickAsset(assets)
	if got == "" {
		t.Fatal("no asset picked")
	}
	if strings.Contains(got, "gui") {
		t.Errorf("tray build picked GUI asset %q", got)
	}
}
