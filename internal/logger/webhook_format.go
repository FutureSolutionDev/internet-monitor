package logger

import (
	"fmt"
	"internet-monitor/internal/config"
	"strings"
	"time"
)

func IsDiscord(url string) bool { return config.IsDiscordWebhook(url) }

func IsSlack(url string) bool { return config.IsSlackWebhook(url) }

func IsSupportedWebhook(url string) bool { return config.IsSupportedWebhook(url) }

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

func field(name, value string, inline bool) map[string]interface{} {
	return map[string]interface{}{"name": name, "value": value, "inline": inline}
}

func formatSlackResult(label string, ok bool, latMs int64, errStr string) string {
	if ok {
		return fmt.Sprintf("• %s: ✅ %dms", label, latMs)
	}
	return fmt.Sprintf("• %s: ❌ %s", label, errStr)
}
