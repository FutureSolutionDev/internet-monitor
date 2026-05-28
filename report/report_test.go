package report

import (
	"internet-monitor/types"
	"testing"
	"time"
)

func ev(ts time.Time, typ string, dur float64, tcp, http, dns bool) types.Event {
	return types.Event{
		Timestamp:       ts,
		EventType:       typ,
		DurationSeconds: dur,
		Reason:          types.EventReason{TCPPingFailed: tcp, HTTPFailed: http, DNSFailed: dns},
	}
}

func TestSummarizeOutages(t *testing.T) {
	base := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	now := base.Add(time.Hour)
	events := []types.Event{
		ev(base, "connected", 0, false, false, false),
		ev(base.Add(5*time.Minute), "disconnected", 300, true, false, true), // prev connected lasted 300s
		ev(base.Add(8*time.Minute), "connected", 180, false, false, false),  // disconnected lasted 180s
	}

	s := Summarize(events, nil, "2026-05", now)

	if s.Disconnections != 1 {
		t.Errorf("Disconnections = %d, want 1", s.Disconnections)
	}
	if s.TotalDowntimeSecs != 180 {
		t.Errorf("TotalDowntimeSecs = %v, want 180", s.TotalDowntimeSecs)
	}
	if s.LongestOutageSecs != 180 {
		t.Errorf("LongestOutageSecs = %v, want 180", s.LongestOutageSecs)
	}
	if s.MTTRSecs != 180 {
		t.Errorf("MTTRSecs = %v, want 180", s.MTTRSecs)
	}
	if s.Causes["tcp"].Outages != 1 || s.Causes["dns"].Outages != 1 || s.Causes["http"].Outages != 0 {
		t.Errorf("cause attribution wrong: %+v", s.Causes)
	}
	if s.Causes["tcp"].DowntimeSecs != 180 {
		t.Errorf("tcp downtime = %v, want 180", s.Causes["tcp"].DowntimeSecs)
	}
	// tcp+dns failed (http ok) -> no L4 reachability -> "down"
	if s.OutageTypes["down"] != 1 {
		t.Errorf("OutageTypes = %+v, want one 'down'", s.OutageTypes)
	}
}

func TestSummarizeOngoingOutage(t *testing.T) {
	base := time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC)
	now := base.Add(2 * time.Minute) // outage still ongoing
	events := []types.Event{
		ev(base, "disconnected", 10, true, false, false),
	}
	s := Summarize(events, nil, "2026-05", now)
	if s.Disconnections != 1 || s.TotalDowntimeSecs != 120 {
		t.Errorf("ongoing outage: disc=%d downtime=%v, want 1/120", s.Disconnections, s.TotalDowntimeSecs)
	}
}

func TestSummarizeSamples(t *testing.T) {
	base := time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)
	samples := []types.MetricSample{
		{Timestamp: base, Samples: 12, UpSamples: 12, AvgLatencyMs: 40, MaxLatencyMs: 90, AvgLossPct: 0},
		{Timestamp: base.Add(time.Minute), Samples: 12, UpSamples: 6, AvgLatencyMs: 80, MaxLatencyMs: 200, AvgLossPct: 10},
	}
	s := Summarize(nil, samples, "2026-05", base.Add(time.Hour))

	if s.MonitoredSeconds != 120 {
		t.Errorf("MonitoredSeconds = %v, want 120", s.MonitoredSeconds)
	}
	if s.UptimePct != 75 { // 18 up / 24 total
		t.Errorf("UptimePct = %v, want 75", s.UptimePct)
	}
	if len(s.Trend) != 1 {
		t.Fatalf("Trend points = %d, want 1", len(s.Trend))
	}
	if s.Trend[0].MaxLatencyMs != 200 {
		t.Errorf("MaxLatencyMs = %d, want 200", s.Trend[0].MaxLatencyMs)
	}
	if s.Trend[0].AvgLatencyMs != 60 { // (40*12 + 80*12)/24
		t.Errorf("AvgLatencyMs = %d, want 60", s.Trend[0].AvgLatencyMs)
	}
}
