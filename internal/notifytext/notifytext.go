package notifytext

import (
	"fmt"
	"internet-monitor/internal/types"
	"strings"
	"sync/atomic"
)

var lang atomic.Value

func SetLang(l string) {
	if l != "ar" {
		l = "en"
	}
	lang.Store(l)
}

func Lang() string {
	if v, ok := lang.Load().(string); ok && v != "" {
		return v
	}
	return "en"
}

func TestMessage(lang string) (title, body string) {
	if lang == "ar" {
		return "اختبار الإشعار", "🔔 الإشعار والصوت يعملان بشكل صحيح"
	}
	return "Test Notification", "🔔 Notifications and sound are working"
}

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

func EscapeAppleScript(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(s)
}
