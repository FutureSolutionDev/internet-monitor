// Package notifytext builds the bilingual notification title/body shown by both
// the tray and GUI builds. It is pure Go (no CGO) so the logic can be shared
// across platform binaries and unit-tested.
package notifytext

import (
	"fmt"
	"internet-monitor/types"
	"strings"
	"sync/atomic"
)

// lang holds the current UI language ("ar"/"en") for OS notifications. It is
// set from config at startup and on every config reload, and read by the live
// notifiers (tray/GUI) which don't otherwise have access to the config.
var lang atomic.Value // string

// SetLang updates the language used by Lang()/BuildCurrent. Safe for concurrent use.
func SetLang(l string) {
	if l != "ar" {
		l = "en"
	}
	lang.Store(l)
}

// Lang returns the current notification language ("ar"/"en"), defaulting to "en".
func Lang() string {
	if v, ok := lang.Load().(string); ok && v != "" {
		return v
	}
	return "en"
}

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

// Build returns the notification (title, body) for a status and check result,
// in a single language ("ar" for Arabic, anything else = English) so OS
// notifications never mix the two.
func Build(lang string, status types.Status, r types.CheckResult) (title, body string) {
	ar := lang == "ar"
	switch status {
	case types.StatusConnected:
		if ar {
			title = "✅ عاد الإنترنت"
			if r.LatencyMs > 0 {
				return title, fmt.Sprintf("زمن الاستجابة: %dms", r.LatencyMs)
			}
			return title, "جميع الفحوصات ناجحة"
		}
		title = "✅ Internet Restored"
		if r.LatencyMs > 0 {
			return title, fmt.Sprintf("Latency: %dms", r.LatencyMs)
		}
		return title, "All checks passing"

	case types.StatusDegraded:
		if ar {
			return "⚠️ الإنترنت ضعيف", fmt.Sprintf("فقدان: %.0f%% — زمن: %dms", r.PacketLoss, r.LatencyMs)
		}
		return "⚠️ Internet Degraded", fmt.Sprintf("Loss: %.0f%% — latency: %dms", r.PacketLoss, r.LatencyMs)

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
		if ar {
			title = "🔴 انقطع الإنترنت"
			if len(parts) > 0 {
				return title, "فشل: " + strings.Join(parts, " + ")
			}
			return title, "فقدان الاتصال"
		}
		title = "🔴 Internet Disconnected"
		if len(parts) > 0 {
			return title, "Failed: " + strings.Join(parts, " + ")
		}
		return title, "Connection lost"
	}
}

// EscapeAppleScript escapes a string for safe embedding inside an AppleScript
// double-quoted literal passed to osascript on macOS.
func EscapeAppleScript(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(s)
}
