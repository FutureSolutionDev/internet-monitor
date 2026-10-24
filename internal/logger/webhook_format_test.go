package logger

import (
	"encoding/json"
	"internet-monitor/internal/monitor"
	"strings"
	"testing"
	"time"
)

func TestTelegramEventText(t *testing.T) {
	ev := monitor.Event{
		Timestamp:       time.Now(),
		EventType:       "disconnected",
		DurationSeconds: 42,
		Reason:          monitor.EventReason{TCPPingFailed: true, DNSFailed: true},
	}
	got := TelegramEventText("en", ev, "down")
	for _, want := range []string{"Internet disconnected", "Duration: 42 s", "DNS: ❌ Failed", "HTTP: ✅ OK"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestEventPayloadIsLocalized(t *testing.T) {
	ev := monitor.Event{Timestamp: time.Now(), EventType: "connected", DurationSeconds: 125}
	for _, url := range []string{"https://discord.com/api/webhooks/1/x", "https://hooks.slack.com/services/T/B/X"} {
		data, _ := json.Marshal(BuildEventPayload("ar", ev, "restored", url))
		s := string(data)
		if !strings.Contains(s, "عاد") || strings.Contains(s, "Internet restored") {
			t.Errorf("%s payload not Arabic: %s", url, s)
		}
	}
}

func TestReminderTitle(t *testing.T) {
	if got := EventTitle("en", "reminder"); !strings.Contains(got, "still") {
		t.Errorf("reminder title = %q", got)
	}
	if got := EventTitle("en", "bogus"); got != EventTitle("en", "down") {
		t.Errorf("unknown kind title = %q", got)
	}
}

func TestTelegramSpeedText(t *testing.T) {
	up := 12.3
	ev := SpeedTestEvent{DownloadMbps: 95.5, UploadMbps: &up, LatencyMs: 18}
	got := TelegramSpeedText("en", ev, 50, false)
	if !strings.Contains(got, "95.5 Mbps") || !strings.Contains(got, "12.3 Mbps") || strings.Contains(got, "Threshold") {
		t.Errorf("unexpected speed text: %q", got)
	}
	got2 := TelegramSpeedText("en", SpeedTestEvent{DownloadMbps: 4}, 50, true)
	if !strings.Contains(got2, "Threshold: 50.0 Mbps") || !strings.Contains(got2, "Speed drop") {
		t.Errorf("threshold warning missing: %q", got2)
	}
}
