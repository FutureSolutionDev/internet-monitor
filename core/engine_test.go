package core

import (
	"internet-monitor/config"
	"internet-monitor/types"
	"testing"
)

func TestDetermineStatus(t *testing.T) {
	cfg := &config.Config{
		FailThreshold:       3,
		PacketLossThreshold: 20,
		LatencyThreshold:    500,
	}

	t.Run("all ok -> connected", func(t *testing.T) {
		fails := 0
		r := types.CheckResult{TCPPingOK: true, HTTPOK: true, DNSOK: true, LatencyMs: 50}
		if got := DetermineStatus(r, &fails, cfg); got != types.StatusConnected {
			t.Errorf("got %v, want connected", got)
		}
	})

	t.Run("high latency -> degraded", func(t *testing.T) {
		fails := 0
		r := types.CheckResult{TCPPingOK: true, HTTPOK: true, DNSOK: true, LatencyMs: 900}
		if got := DetermineStatus(r, &fails, cfg); got != types.StatusDegraded {
			t.Errorf("got %v, want degraded", got)
		}
	})

	t.Run("consecutive failures reach threshold -> disconnected", func(t *testing.T) {
		fails := 0
		r := types.CheckResult{TCPPingOK: false, HTTPOK: false, DNSOK: false}
		DetermineStatus(r, &fails, cfg) // 1
		DetermineStatus(r, &fails, cfg) // 2
		if got := DetermineStatus(r, &fails, cfg); got != types.StatusDisconnected {
			t.Errorf("got %v, want disconnected after %d fails", got, fails)
		}
	})

	t.Run("single failure below threshold is not disconnected", func(t *testing.T) {
		fails := 0
		r := types.CheckResult{TCPPingOK: false, HTTPOK: true, DNSOK: true}
		if got := DetermineStatus(r, &fails, cfg); got == types.StatusDisconnected {
			t.Errorf("got disconnected on first failure; threshold=%d", cfg.FailThreshold)
		}
	})

	t.Run("recovery resets consecutive failure counter", func(t *testing.T) {
		fails := 2
		r := types.CheckResult{TCPPingOK: true, HTTPOK: true, DNSOK: true, LatencyMs: 10}
		DetermineStatus(r, &fails, cfg)
		if fails != 0 {
			t.Errorf("consecFails = %d, want 0 after success", fails)
		}
	})
}

func TestApplyConfigNonBlocking(t *testing.T) {
	e := New(&config.Default, nil, nil, "dev")
	// Buffer is 1; calling repeatedly must never block even with no consumer.
	for i := 0; i < 5; i++ {
		e.ApplyConfig(&config.Default)
	}
	e.ApplyConfig(nil) // must be a safe no-op
}
