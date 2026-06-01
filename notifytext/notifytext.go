// Package notifytext builds the bilingual notification title/body shown by both
// the tray and GUI builds. It is pure Go (no CGO) so the logic can be shared
// across platform binaries and unit-tested.
package notifytext

import (
	"fmt"
	"internet-monitor/types"
	"strings"
)

// TestMessage returns the (title, body) for a manual test notification in the
// given UI language ("ar" for Arabic, anything else = English).
func TestMessage(lang string) (title, body string) {
	if lang == "ar" {
		return "اختبار الإشعار", "🔔 الإشعار والصوت يعملان بشكل صحيح"
	}
	return "Test Notification", "🔔 Notifications and sound are working"
}

// StatusFromEventType maps an Event.EventType string to a Status.
func StatusFromEventType(eventType string) types.Status {
	switch eventType {
	case "connected":
		return types.StatusConnected
	case "degraded":
		return types.StatusDegraded
	default:
		return types.StatusDisconnected
	}
}

// Build returns the notification (title, body) for a status and check result.
func Build(status types.Status, r types.CheckResult) (title, body string) {
	switch status {
	case types.StatusConnected:
		if r.LatencyMs > 0 {
			return "✅ الإنترنت عاد / Restored", fmt.Sprintf("زمن الاستجابة: %dms", r.LatencyMs)
		}
		return "✅ الإنترنت عاد / Restored", "جميع الفحوصات ناجحة"

	case types.StatusDegraded:
		return "⚠️ الإنترنت ضعيف / Degraded",
			fmt.Sprintf("فقدان: %.0f%% — زمن: %dms", r.PacketLoss, r.LatencyMs)

	default:
		var parts []string
		if !r.TCPPingOK {
			parts = append(parts, "TCP")
		}
		if !r.HTTPOK {
			parts = append(parts, "HTTP")
		}
		if !r.DNSOK {
			parts = append(parts, "DNS")
		}
		if len(parts) > 0 {
			return "🔴 الإنترنت انقطع / Disconnected", strings.Join(parts, " + ") + " فشل"
		}
		return "🔴 الإنترنت انقطع / Disconnected", "فقدان الاتصال"
	}
}

// EscapeAppleScript escapes a string for safe embedding inside an AppleScript
// double-quoted literal passed to osascript on macOS.
func EscapeAppleScript(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(s)
}
