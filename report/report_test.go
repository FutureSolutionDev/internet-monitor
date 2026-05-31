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
	if len(s.Events) != 1 || s.Events[0].Type != "disconnected" || s.Events[0].Cause != "tcp+dns" {
		t.Errorf("Events = %+v, want one disconnected with cause tcp+dns", s.Events)
	}
	if s.Causes["tcp"].Pct != 100 {
		t.Errorf("tcp Pct = %v, want 100", s.Causes["tcp"].Pct)
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

func TestSummarizeFiltersByMonth(t *testing.T) {
	may := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	jun := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	events := []types.Event{
		ev(may, "disconnected", 0, true, false, false),
		ev(may.Add(time.Minute), "connected", 60, false, false, false),
		ev(jun, "disconnected", 0, true, false, false),
		ev(jun.Add(2*time.Minute), "connected", 120, false, false, false),
	}
	s := Summarize(events, nil, "2026-05", jun.Add(time.Hour))
	if s.Disconnections != 1 {
		t.Errorf("Disconnections = %d, want 1 (June event must be filtered out)", s.Disconnections)
	}
}

func TestSummarizeSplitsOutageAcrossDays(t *testing.T) {
	// Outage starts at 23:30 on May 10, lasts 2h (recovers 01:30 on May 11).
	start := time.Date(2026, 5, 10, 23, 30, 0, 0, time.UTC)
	events := []types.Event{
		ev(start, "disconnected", 0, true, false, false),
		ev(start.Add(2*time.Hour), "connected", 7200, false, false, false),
	}
	s := Summarize(events, nil, "2026-05", start.Add(3*time.Hour))

	if len(s.Days) != 2 {
		t.Fatalf("Days = %d, want 2", len(s.Days))
	}
	// Day 1: 30 minutes = 1800s. Day 2: 90 minutes = 5400s.
	var d10, d11 *DayStat
	for i := range s.Days {
		switch s.Days[i].Date {
		case "2026-05-10":
			d10 = &s.Days[i]
		case "2026-05-11":
			d11 = &s.Days[i]
		}
	}
	if d10 == nil || d11 == nil {
		t.Fatalf("missing day rows: %+v", s.Days)
	}
	if d10.DowntimeSecs != 1800 {
		t.Errorf("day10 downtime = %v, want 1800", d10.DowntimeSecs)
	}
	if d11.DowntimeSecs != 5400 {
		t.Errorf("day11 downtime = %v, want 5400", d11.DowntimeSecs)
	}
	if d10.Outages != 1 || d11.Outages != 0 {
		t.Errorf("outage count belongs to start day only: d10=%d d11=%d", d10.Outages, d11.Outages)
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
