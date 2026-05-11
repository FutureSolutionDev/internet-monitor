package logger

import (
	"fmt"
	"internet-monitor/monitor"
	"strings"
	"time"
)

// ── URL Validation ────────────────────────────────────────────

func IsDiscord(url string) bool {
	return strings.Contains(url, "discord.com/api/webhooks") ||
		strings.Contains(url, "discordapp.com/api/webhooks")
}

func IsSlack(url string) bool {
	return strings.Contains(url, "hooks.slack.com") ||
		strings.Contains(url, "slack.com/services/")
}

// IsSupportedWebhook returns true only for Discord and Slack URLs.
func IsSupportedWebhook(url string) bool {
	return IsDiscord(url) || IsSlack(url)
}

// ── Connectivity Event Payloads ───────────────────────────────

func BuildEventPayload(event monitor.Event, url string) interface{} {
	if IsDiscord(url) {
		return discordEventPayload(event)
	}
	return slackEventPayload(event)
}

func discordEventPayload(event monitor.Event) map[string]interface{} {
	colorMap := map[string]int{
		"connected":    0x22C55E,
		"degraded":     0xEAB308,
		"disconnected": 0xEF4444,
	}
	emojiMap := map[string]string{
		"connected":    "✅",
		"degraded":     "⚠️",
		"disconnected": "❌",
	}
	color := colorMap[event.EventType]
	emoji := emojiMap[event.EventType]

	fields := []map[string]interface{}{}
	if event.Reason.TCPPingFailed {
		fields = append(fields, field("TCP Ping", "❌ Failed", true))
	}
	if event.Reason.HTTPFailed {
		fields = append(fields, field("HTTP", "❌ Failed", true))
	}
	if event.Reason.DNSFailed {
		fields = append(fields, field("DNS", "❌ Failed", true))
	}
	if event.Reason.PacketLossPct > 0 {
		fields = append(fields, field("Packet Loss", fmt.Sprintf("%.0f%%", event.Reason.PacketLossPct), true))
	}
	if event.Reason.AvgLatencyMs > 0 {
		fields = append(fields, field("Latency", fmt.Sprintf("%dms", event.Reason.AvgLatencyMs), true))
	}
	if event.DurationSeconds > 0 {
		fields = append(fields, field("Duration", fmt.Sprintf("%.0fs", event.DurationSeconds), true))
	}

	return map[string]interface{}{
		"username": "Internet Monitor",
		"embeds": []map[string]interface{}{{
			"title":     fmt.Sprintf("%s Internet %s", emoji, event.EventType),
			"color":     color,
			"fields":    fields,
			"timestamp": event.Timestamp.UTC().Format(time.RFC3339),
			"footer":    map[string]string{"text": "Internet Monitor"},
		}},
	}
}

func slackEventPayload(event monitor.Event) map[string]interface{} {
	emojiMap := map[string]string{
		"connected":    ":white_check_mark:",
		"degraded":     ":warning:",
		"disconnected": ":x:",
	}
	emoji := emojiMap[event.EventType]

	reasons := []string{}
	if event.Reason.TCPPingFailed {
		reasons = append(reasons, "TCP Ping failed")
	}
	if event.Reason.HTTPFailed {
		reasons = append(reasons, "HTTP failed")
	}
	if event.Reason.DNSFailed {
		reasons = append(reasons, "DNS failed")
	}
	if event.Reason.PacketLossPct > 0 {
		reasons = append(reasons, fmt.Sprintf("Loss %.0f%%", event.Reason.PacketLossPct))
	}

	text := fmt.Sprintf("%s *Internet %s*", emoji, event.EventType)
	if len(reasons) > 0 {
		text += "\n" + strings.Join(reasons, " | ")
	}
	if event.DurationSeconds > 0 {
		text += fmt.Sprintf("\nDuration: %.0fs", event.DurationSeconds)
	}

	return map[string]interface{}{
		"text": text,
	}
}

// ── Test Results Payloads ─────────────────────────────────────

// TestResult holds a single target test result (shared between caller and formatter).
type TestResult struct {
	Target    string
	OK        bool
	LatencyMs int64
	Error     string
}

type TestResults struct {
	PingTargets []TestResult
	HTTPTarget  *TestResult
	DNSTarget   *TestResult
}

func (r TestResults) AllOK() bool {
	for _, p := range r.PingTargets {
		if !p.OK {
			return false
		}
	}
	if r.HTTPTarget != nil && !r.HTTPTarget.OK {
		return false
	}
	if r.DNSTarget != nil && !r.DNSTarget.OK {
		return false
	}
	return true
}

func BuildTestPayload(results TestResults, url string) interface{} {
	if IsDiscord(url) {
		return discordTestPayload(results)
	}
	return slackTestPayload(results)
}

func discordTestPayload(results TestResults) map[string]interface{} {
	allOK := results.AllOK()
	color := 0x22C55E
	if !allOK {
		color = 0xEF4444
	}
	title := "🔍 Manual Test — ✅ All Targets OK"
	if !allOK {
		title = "🔍 Manual Test — ⚠️ Some Targets Failed"
	}

	fields := []map[string]interface{}{}

	for _, r := range results.PingTargets {
		val := fmt.Sprintf("✅ %dms", r.LatencyMs)
		if !r.OK {
			val = "❌ " + r.Error
		}
		fields = append(fields, field("TCP: "+r.Target, val, true))
	}
	if results.HTTPTarget != nil {
		val := fmt.Sprintf("✅ %dms", results.HTTPTarget.LatencyMs)
		if !results.HTTPTarget.OK {
			val = "❌ " + results.HTTPTarget.Error
		}
		fields = append(fields, field("HTTP", val, true))
	}
	if results.DNSTarget != nil {
		val := fmt.Sprintf("✅ %dms", results.DNSTarget.LatencyMs)
		if !results.DNSTarget.OK {
			val = "❌ " + results.DNSTarget.Error
		}
		fields = append(fields, field("DNS", val, true))
	}

	return map[string]interface{}{
		"username": "Internet Monitor",
		"embeds": []map[string]interface{}{{
			"title":     title,
			"color":     color,
			"fields":    fields,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"footer":    map[string]string{"text": "Internet Monitor"},
		}},
	}
}

func slackTestPayload(results TestResults) map[string]interface{} {
	allOK := results.AllOK()
	header := ":mag: Manual Test — :white_check_mark: All OK"
	if !allOK {
		header = ":mag: Manual Test — :warning: Some Failed"
	}

	lines := []string{header}
	for _, r := range results.PingTargets {
		if r.OK {
			lines = append(lines, fmt.Sprintf("• TCP %s: ✅ %dms", r.Target, r.LatencyMs))
		} else {
			lines = append(lines, fmt.Sprintf("• TCP %s: ❌ %s", r.Target, r.Error))
		}
	}
	if results.HTTPTarget != nil {
		if results.HTTPTarget.OK {
			lines = append(lines, fmt.Sprintf("• HTTP: ✅ %dms", results.HTTPTarget.LatencyMs))
		} else {
			lines = append(lines, "• HTTP: ❌ "+results.HTTPTarget.Error)
		}
	}
	if results.DNSTarget != nil {
		if results.DNSTarget.OK {
			lines = append(lines, fmt.Sprintf("• DNS: ✅ %dms", results.DNSTarget.LatencyMs))
		} else {
			lines = append(lines, "• DNS: ❌ "+results.DNSTarget.Error)
		}
	}

	return map[string]interface{}{
		"text": strings.Join(lines, "\n"),
	}
}

// field is a Discord embed field helper.
func field(name, value string, inline bool) map[string]interface{} {
	return map[string]interface{}{"name": name, "value": value, "inline": inline}
}
