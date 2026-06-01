package logger

import (
	"internet-monitor/monitor"
	"strings"
	"testing"
	"time"
)

func TestTelegramEventText(t *testing.T) {
	ev := monitor.Event{
		Timestamp:       time.Now(),
		EventType:       "disconnected",
		DurationSeconds: 42,
		Reason:          monitor.EventReason{TCPPingFailed: true, DNSFailed: true, AvgLatencyMs: 0},
	}
	got := TelegramEventText(ev)
	if !strings.Contains(got, "Internet Disconnected") {
		t.Errorf("missing title: %q", got)
	}
	if !strings.Contains(got, "Duration: 42s") {
		t.Errorf("missing duration: %q", got)
	}
}

func TestTelegramSpeedText(t *testing.T) {
	up := 12.3
	ev := SpeedTestEvent{DownloadMbps: 95.5, UploadMbps: &up, LatencyMs: 18}
	got := TelegramSpeedText(ev, 50, false)
	if !strings.Contains(got, "95.5 Mbps") || !strings.Contains(got, "12.3 Mbps") {
		t.Errorf("download/upload missing: %q", got)
	}

	got2 := TelegramSpeedText(SpeedTestEvent{DownloadMbps: 4}, 50, true)
	if !strings.Contains(got2, "Below threshold") {
		t.Errorf("threshold warning missing: %q", got2)
	}
}
