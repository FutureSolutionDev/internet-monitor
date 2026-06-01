package notifytext

import (
	"internet-monitor/types"
	"strings"
	"testing"
)

func TestStatusFromEventType(t *testing.T) {
	cases := map[string]types.Status{
		"connected":    types.StatusConnected,
		"degraded":     types.StatusDegraded,
		"disconnected": types.StatusDisconnected,
		"weird":        types.StatusDisconnected,
	}
	for in, want := range cases {
		if got := StatusFromEventType(in); got != want {
			t.Errorf("StatusFromEventType(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestBuildDisconnectedListsFailedLayers(t *testing.T) {
	_, body := Build("en", types.StatusDisconnected, types.CheckResult{TCPPingOK: false, HTTPOK: true, DNSOK: false})
	if !strings.Contains(body, "TCP") || !strings.Contains(body, "DNS") || strings.Contains(body, "HTTP") {
		t.Errorf("unexpected disconnected body: %q", body)
	}
}

func TestBuildConnectedFallback(t *testing.T) {
	_, body := Build("en", types.StatusConnected, types.CheckResult{LatencyMs: 0})
	if body == "" {
		t.Error("connected body should not be empty when latency is 0")
	}
}

// Build must never mix languages: each language's text contains only its own
// script (the bug the user reported was Arabic+English in one notification).
func TestBuildSingleLanguage(t *testing.T) {
	hasArabic := func(s string) bool {
		for _, r := range s {
			if r >= 0x0600 && r <= 0x06FF {
				return true
			}
		}
		return false
	}
	statuses := []types.Status{types.StatusConnected, types.StatusDegraded, types.StatusDisconnected}
	for _, st := range statuses {
		// English must contain no Arabic.
		if tt, bb := Build("en", st, types.CheckResult{}); hasArabic(tt) || hasArabic(bb) {
			t.Errorf("en %v leaked Arabic: %q / %q", st, tt, bb)
		}
		// Arabic title must contain Arabic (not be English-only).
		if tt, _ := Build("ar", st, types.CheckResult{}); !hasArabic(tt) {
			t.Errorf("ar %v has no Arabic title: %q", st, tt)
		}
	}
}

func TestEscapeAppleScript(t *testing.T) {
	got := EscapeAppleScript(`a"b\c`)
	if got != `a\"b\\c` {
		t.Errorf("EscapeAppleScript = %q", got)
	}
}
