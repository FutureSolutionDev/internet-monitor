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

func IsSupportedWebhook(url string) bool {
	return IsDiscord(url) || IsSlack(url)
}

// ── Connectivity Event Payloads (developer-detail level) ──────

func BuildEventPayload(event monitor.Event, url string) interface{} {
	if IsDiscord(url) {
		return discordEventPayload(event)
	}
	return slackEventPayload(event)
}

func discordEventPayload(event monitor.Event) map[string]interface{} {
	colors := map[string]int{
		"connected":    0x22C55E,
		"degraded":     0xEAB308,
		"disconnected": 0xEF4444,
	}
	emojis := map[string]string{
		"connected":    "✅",
		"degraded":     "⚠️",
		"disconnected": "❌",
	}

	color := colors[event.EventType]
	emoji := emojis[event.EventType]

	// ── Check results section ────────────────────────────────
	checks := []map[string]interface{}{}

	tcpVal := "✅ OK"
	if event.Reason.TCPPingFailed {
		tcpVal = "❌ Failed"
	}
	checks = append(checks, field("🔌 TCP Ping", tcpVal, true))

	httpVal := "✅ OK"
	if event.Reason.HTTPFailed {
		httpVal = "❌ Failed"
	}
	checks = append(checks, field("🌐 HTTP", httpVal, true))

	dnsVal := "✅ OK"
	if event.Reason.DNSFailed {
		dnsVal = "❌ Failed"
	}
	checks = append(checks, field("🔍 DNS", dnsVal, true))

	// ── Metrics section ──────────────────────────────────────
	metrics := []map[string]interface{}{}

	lossStr := fmt.Sprintf("%.1f%%", event.Reason.PacketLossPct)
	if event.Reason.PacketLossPct == 0 {
		lossStr = "0%"
	}
	metrics = append(metrics, field("📉 Packet Loss", lossStr, true))

	latStr := fmt.Sprintf("%dms", event.Reason.AvgLatencyMs)
	if event.Reason.AvgLatencyMs == 0 {
		latStr = "—"
	}
	metrics = append(metrics, field("⚡ Latency", latStr, true))

	durStr := "—"
	if event.DurationSeconds > 0 {
		d := time.Duration(event.DurationSeconds) * time.Second
		if d < time.Minute {
			durStr = fmt.Sprintf("%.0fs", event.DurationSeconds)
		} else {
			durStr = fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
		}
	}
	metrics = append(metrics, field("⏱️ Duration", durStr, true))

	// ── Combine all fields ───────────────────────────────────
	allFields := append(checks, metrics...)

	return map[string]interface{}{
		"username": "Internet Monitor",
		"embeds": []map[string]interface{}{{
			"title":       fmt.Sprintf("%s Internet %s", emoji, strings.Title(event.EventType)),
			"color":       color,
			"fields":      allFields,
			"timestamp":   event.Timestamp.UTC().Format(time.RFC3339),
			"footer":      map[string]string{"text": "Internet Monitor • Event Log"},
		}},
	}
}

func slackEventPayload(event monitor.Event) map[string]interface{} {
	emojis := map[string]string{
		"connected":    ":white_check_mark:",
		"degraded":     ":warning:",
		"disconnected": ":x:",
	}
	emoji := emojis[event.EventType]

	lines := []string{
		fmt.Sprintf("%s *Internet %s*", emoji, strings.Title(event.EventType)),
		"",
	}

	lines = append(lines, fmt.Sprintf("*TCP Ping:* %s  *HTTP:* %s  *DNS:* %s",
		boolCheck(!event.Reason.TCPPingFailed),
		boolCheck(!event.Reason.HTTPFailed),
		boolCheck(!event.Reason.DNSFailed),
	))

	if event.Reason.PacketLossPct > 0 {
		lines = append(lines, fmt.Sprintf("*Packet Loss:* %.1f%%", event.Reason.PacketLossPct))
	}
	if event.Reason.AvgLatencyMs > 0 {
		lines = append(lines, fmt.Sprintf("*Latency:* %dms", event.Reason.AvgLatencyMs))
	}
	if event.DurationSeconds > 0 {
		lines = append(lines, fmt.Sprintf("*Duration:* %.0fs", event.DurationSeconds))
	}

	return map[string]interface{}{"text": strings.Join(lines, "\n")}
}

// ── Test Results Payloads ─────────────────────────────────────

type TestResult struct {
	Target    string
	OK        bool
	LatencyMs int64
	Error     string
}

type TestResults struct {
	PingTargets []TestResult
	HTTPTargets []TestResult
	DNSTargets  []TestResult
}

func (r TestResults) AllOK() bool {
	for _, p := range r.PingTargets {
		if !p.OK {
			return false
		}
	}
	for _, h := range r.HTTPTargets {
		if !h.OK {
			return false
		}
	}
	for _, d := range r.DNSTargets {
		if !d.OK {
			return false
		}
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
			val = fmt.Sprintf("❌ %s", r.Error)
		}
		fields = append(fields, field("🔌 TCP: "+r.Target, val, true))
	}

	for _, r := range results.HTTPTargets {
		val := fmt.Sprintf("✅ %dms", r.LatencyMs)
		if !r.OK {
			val = fmt.Sprintf("❌ %s", r.Error)
		}
		short := r.Target
		if len(short) > 35 {
			short = short[:32] + "…"
		}
		fields = append(fields, field("🌐 HTTP: "+short, val, true))
	}

	for _, r := range results.DNSTargets {
		val := fmt.Sprintf("✅ %dms", r.LatencyMs)
		if !r.OK {
			val = fmt.Sprintf("❌ %s", r.Error)
		}
		fields = append(fields, field("🔍 DNS: "+r.Target, val, true))
	}

	return map[string]interface{}{
		"username": "Internet Monitor",
		"embeds": []map[string]interface{}{{
			"title":     title,
			"color":     color,
			"fields":    fields,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"footer":    map[string]string{"text": "Internet Monitor • Manual Test"},
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
		lines = append(lines, formatSlackResult("TCP: "+r.Target, r.OK, r.LatencyMs, r.Error))
	}
	for _, r := range results.HTTPTargets {
		lines = append(lines, formatSlackResult("HTTP: "+r.Target, r.OK, r.LatencyMs, r.Error))
	}
	for _, r := range results.DNSTargets {
		lines = append(lines, formatSlackResult("DNS: "+r.Target, r.OK, r.LatencyMs, r.Error))
	}

	return map[string]interface{}{"text": strings.Join(lines, "\n")}
}

// ── Helpers ───────────────────────────────────────────────────

func field(name, value string, inline bool) map[string]interface{} {
	return map[string]interface{}{"name": name, "value": value, "inline": inline}
}

func boolCheck(ok bool) string {
	if ok {
		return ":white_check_mark:"
	}
	return ":x:"
}

func formatSlackResult(label string, ok bool, latMs int64, errStr string) string {
	if ok {
		return fmt.Sprintf("• %s: ✅ %dms", label, latMs)
	}
	return fmt.Sprintf("• %s: ❌ %s", label, errStr)
}
