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
	_, body := Build(types.StatusDisconnected, types.CheckResult{TCPPingOK: false, HTTPOK: true, DNSOK: false})
	if !strings.Contains(body, "TCP") || !strings.Contains(body, "DNS") || strings.Contains(body, "HTTP") {
		t.Errorf("unexpected disconnected body: %q", body)
	}
}

func TestBuildConnectedFallback(t *testing.T) {
	_, body := Build(types.StatusConnected, types.CheckResult{LatencyMs: 0})
	if body == "" {
		t.Error("connected body should not be empty when latency is 0")
	}
}

func TestEscapeAppleScript(t *testing.T) {
	got := EscapeAppleScript(`a"b\c`)
	if got != `a\"b\\c` {
		t.Errorf("EscapeAppleScript = %q", got)
	}
}
