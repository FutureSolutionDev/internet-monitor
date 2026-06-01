package dashboard

import (
	"context"
	"internet-monitor/types"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServeMetrics(t *testing.T) {
	s := NewServer(0, "config.json", "logs", "v1.2.3", nil)
	s.UpdateTick(types.CheckResult{TCPPingOK: true, HTTPOK: true, DNSOK: true, LatencyMs: 42}, types.StatusConnected)

	rec := httptest.NewRecorder()
	s.serveMetrics(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := rec.Body.String()
	for _, want := range []string{
		`internet_monitor_build_info{version="v1.2.3"} 1`,
		"internet_monitor_up 1",
		"internet_monitor_latency_ms 42",
		"internet_monitor_checks_total 1",
		"internet_monitor_up_checks_total 1",
		"# TYPE internet_monitor_up gauge",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q\n---\n%s", want, body)
		}
	}
}

func TestJitterOf(t *testing.T) {
	if got := jitterOf([]int64{10}); got != 0 {
		t.Errorf("single sample jitter = %d, want 0", got)
	}
	// diffs: 10,10,10 -> mean 10
	if got := jitterOf([]int64{10, 20, 10, 20}); got != 10 {
		t.Errorf("jitter = %d, want 10", got)
	}
	if got := jitterOf([]int64{50, 50, 50}); got != 0 {
		t.Errorf("steady jitter = %d, want 0", got)
	}
}

func TestGracefulShutdown(t *testing.T) {
	// port 0 lets the OS pick a free port; we only care that Shutdown returns.
	s := NewServer(0, "config.json", "logs", "test", nil)
	s.Start()
	time.Sleep(50 * time.Millisecond) // let ListenAndServe bind

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- s.Shutdown(ctx) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Shutdown returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Shutdown did not return within 3s")
	}

	// A second call must be safe (sync.Once guards the channel close).
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatalf("second Shutdown error: %v", err)
	}
}
