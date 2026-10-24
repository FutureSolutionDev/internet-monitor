package logger

import (
	"fmt"
	"internet-monitor/internal/i18n"
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
	"strings"
	"time"
)

type remoteStyle struct {
	color int
	emoji string
	slack string
	title func(lang string) string
}

var remoteStyles = map[string]remoteStyle{
	"down":     {0xEF4444, "❌", ":x:", func(l string) string { return i18n.T(l, "remote.title.down", nil) }},
	"degraded": {0xEAB308, "⚠️", ":warning:", func(l string) string { return i18n.T(l, "remote.title.degraded", nil) }},
	"restored": {0x22C55E, "✅", ":white_check_mark:", func(l string) string { return i18n.T(l, "remote.title.restored", nil) }},
	"reminder": {0xEF4444, "⏰", ":alarm_clock:", func(l string) string { return i18n.T(l, "remote.title.reminder", nil) }},
}

func styleOf(kind string) remoteStyle {
	if s, ok := remoteStyles[kind]; ok {
		return s
	}
	return remoteStyles["down"]
}

func EventTitle(lang, kind string) string { return styleOf(kind).title(lang) }

type remoteLine struct{ name, value string }

func eventLines(lang string, ev monitor.Event) (checks, metrics []remoteLine) {
	mark := func(failed bool) string {
		if failed {
			return "❌ " + i18n.T(lang, "remote.failed", nil)
		}
		return "✅ " + i18n.T(lang, "remote.ok", nil)
	}
	checks = []remoteLine{
		{"🔌 " + i18n.T(lang, "remote.tcp", nil), mark(ev.Reason.TCPPingFailed)},
		{"🌐 " + i18n.T(lang, "remote.http", nil), mark(ev.Reason.HTTPFailed)},
		{"🔍 " + i18n.T(lang, "remote.dns", nil), mark(ev.Reason.DNSFailed)},
	}
	if ev.Reason.PacketLossPct > 0 {
		metrics = append(metrics, remoteLine{"📉 " + i18n.T(lang, "remote.loss", nil), fmt.Sprintf("%.1f%%", ev.Reason.PacketLossPct)})
	}
	if ev.Reason.AvgLatencyMs > 0 {
		metrics = append(metrics, remoteLine{"⚡ " + i18n.T(lang, "remote.latency", nil), notifytext.Latency(lang, ev.Reason.AvgLatencyMs)})
	}
	if ev.DurationSeconds >= 1 {
		metrics = append(metrics, remoteLine{"⏱️ " + i18n.T(lang, "remote.duration", nil), notifytext.Duration(lang, ev.DurationSeconds)})
	}
	return checks, metrics
}

func BuildEventPayload(lang string, ev monitor.Event, kind, url string) interface{} {
	st := styleOf(kind)
	title := EventTitle(lang, kind)
	checks, metrics := eventLines(lang, ev)
	if IsDiscord(url) {
		fields := make([]map[string]interface{}, 0, len(checks)+len(metrics))
		for _, l := range append(checks, metrics...) {
			fields = append(fields, field(l.name, l.value, true))
		}
		return map[string]interface{}{
			"username": "Internet Monitor",
			"embeds": []map[string]interface{}{{
				"title":     st.emoji + " " + title,
				"color":     st.color,
				"fields":    fields,
				"timestamp": ev.Timestamp.UTC().Format(time.RFC3339),
				"footer":    map[string]string{"text": "Internet Monitor • " + i18n.T(lang, "remote.footer", nil)},
			}},
		}
	}
	lines := []string{fmt.Sprintf("%s *%s*", st.slack, title), ""}
	parts := make([]string, 0, len(checks))
	for _, l := range checks {
		parts = append(parts, fmt.Sprintf("*%s:* %s", l.name, l.value))
	}
	lines = append(lines, strings.Join(parts, "  "))
	for _, l := range metrics {
		lines = append(lines, fmt.Sprintf("*%s:* %s", l.name, l.value))
	}
	return map[string]interface{}{"text": strings.Join(lines, "\n")}
}

func TelegramEventText(lang string, ev monitor.Event, kind string) string {
	checks, metrics := eventLines(lang, ev)
	var b strings.Builder
	fmt.Fprintf(&b, "%s *%s*", styleOf(kind).emoji, EventTitle(lang, kind))
	for _, l := range append(checks, metrics...) {
		fmt.Fprintf(&b, "\n%s: %s", l.name, l.value)
	}
	return b.String()
}

func speedLines(lang string, ev SpeedTestEvent, thresholdMbps float64, below bool) []remoteLine {
	lines := []remoteLine{{"📥 " + i18n.T(lang, "remote.download", nil), fmt.Sprintf("%.1f Mbps", ev.DownloadMbps)}}
	if ev.UploadMbps != nil {
		lines = append(lines, remoteLine{"📤 " + i18n.T(lang, "remote.upload", nil), fmt.Sprintf("%.1f Mbps", *ev.UploadMbps)})
	}
	lines = append(lines, remoteLine{"📡 " + i18n.T(lang, "remote.latency", nil), notifytext.Latency(lang, ev.LatencyMs)})
	if ev.DurationSeconds > 0 {
		lines = append(lines, remoteLine{"⏱️ " + i18n.T(lang, "remote.duration", nil), fmt.Sprintf("%.1f s", ev.DurationSeconds)})
	}
	if below && thresholdMbps > 0 {
		lines = append(lines, remoteLine{"⚠️ " + i18n.T(lang, "remote.threshold", nil), fmt.Sprintf("%.1f Mbps", thresholdMbps)})
	}
	return lines
}

func speedTitle(lang string, below bool) (string, string, int) {
	if below {
		return "⚠️", i18n.T(lang, "remote.speed.drop", nil), 0xEAB308
	}
	return "🚀", i18n.T(lang, "remote.speed.done", nil), 0x22C55E
}

func BuildSpeedResultPayload(lang string, ev SpeedTestEvent, thresholdMbps float64, below bool, url string) interface{} {
	icon, title, color := speedTitle(lang, below)
	lines := speedLines(lang, ev, thresholdMbps, below)
	if IsDiscord(url) {
		fields := make([]map[string]interface{}, 0, len(lines))
		for _, l := range lines {
			fields = append(fields, field(l.name, l.value, true))
		}
		return map[string]interface{}{
			"username": "Internet Monitor",
			"embeds": []map[string]interface{}{{
				"title":     icon + " " + title,
				"color":     color,
				"fields":    fields,
				"timestamp": ev.Timestamp.UTC().Format(time.RFC3339),
			}},
		}
	}
	parts := make([]string, 0, len(lines))
	for _, l := range lines {
		parts = append(parts, fmt.Sprintf("%s: *%s*", l.name, l.value))
	}
	return map[string]interface{}{
		"blocks": []map[string]interface{}{{
			"type": "section",
			"text": map[string]interface{}{"type": "mrkdwn", "text": fmt.Sprintf("*%s %s*\n%s", icon, title, strings.Join(parts, " | "))},
		}},
	}
}

func TelegramSpeedText(lang string, ev SpeedTestEvent, thresholdMbps float64, below bool) string {
	icon, title, _ := speedTitle(lang, below)
	var b strings.Builder
	fmt.Fprintf(&b, "%s *%s*", icon, title)
	for _, l := range speedLines(lang, ev, thresholdMbps, below) {
		fmt.Fprintf(&b, "\n%s: %s", l.name, l.value)
	}
	return b.String()
}
