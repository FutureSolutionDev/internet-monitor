package main

import (
	"fmt"
	"internet-monitor/monitor"
	"strings"
)

func notifyText(status monitor.Status, result monitor.CheckResult) (title, body string) {
	switch status {
	case monitor.StatusConnected:
		return "✅ الإنترنت عاد / Restored",
			fmt.Sprintf("زمن الاستجابة: %dms", result.LatencyMs)

	case monitor.StatusDisconnected:
		var parts []string
		if !result.TCPPingOK {
			parts = append(parts, "TCP")
		}
		if !result.HTTPOK {
			parts = append(parts, "HTTP")
		}
		if !result.DNSOK {
			parts = append(parts, "DNS")
		}
		if len(parts) > 0 {
			body = strings.Join(parts, " + ") + " فشل"
		} else {
			body = "فقدان الاتصال"
		}
		return "🔴 الإنترنت انقطع / Disconnected", body

	case monitor.StatusDegraded:
		return "⚠️ الإنترنت ضعيف / Degraded",
			fmt.Sprintf("فقدان: %.0f%% — زمن: %dms", result.PacketLoss, result.LatencyMs)
	}
	return "", ""
}
