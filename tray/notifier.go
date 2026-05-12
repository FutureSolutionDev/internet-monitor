package tray

import (
	"fmt"
	"internet-monitor/types"
	"strings"
	"sync"
	"time"
)

const notifyCooldown = 4 * time.Second

// TrayNotifier updates the system tray icon and sends OS notifications.
type TrayNotifier struct {
	t        *Tray
	mu       sync.Mutex
	lastTime time.Time
}

func NewNotifier(t *Tray) *TrayNotifier {
	return &TrayNotifier{t: t}
}

// OnTick updates the tooltip on every check cycle (latency/loss may change).
func (n *TrayNotifier) OnTick(result types.CheckResult, status types.Status) {
	n.t.updateTooltip(status, result)
}

// OnEvent updates icon + menu label + tooltip only when status changes.
func (n *TrayNotifier) OnEvent(event types.Event) {
	s := eventTypeToStatus(event.EventType)
	n.t.applyStatus(s, types.CheckResult{
		TCPPingOK:  !event.Reason.TCPPingFailed,
		HTTPOK:     !event.Reason.HTTPFailed,
		DNSOK:      !event.Reason.DNSFailed,
		LatencyMs:  event.Reason.AvgLatencyMs,
		PacketLoss: event.Reason.PacketLossPct,
	})

	title := titleForStatus(s)
	body := bodyForEvent(event)

	// Debounce: skip if a notification was sent within the cooldown window.
	n.mu.Lock()
	if time.Since(n.lastTime) < notifyCooldown {
		n.mu.Unlock()
		return
	}
	n.lastTime = time.Now()
	n.mu.Unlock()

	go Notify(title, body)
}

func eventTypeToStatus(eventType string) types.Status {
	switch eventType {
	case "connected":
		return types.StatusConnected
	case "degraded":
		return types.StatusDegraded
	default:
		return types.StatusDisconnected
	}
}

func titleForStatus(s types.Status) string {
	switch s {
	case types.StatusConnected:
		return "Internet Restored"
	case types.StatusDegraded:
		return "Connection Degraded"
	default:
		return "Internet Disconnected"
	}
}

func bodyForEvent(event types.Event) string {
	r := event.Reason
	s := eventTypeToStatus(event.EventType)

	switch s {
	case types.StatusConnected:
		if r.AvgLatencyMs > 0 {
			return fmt.Sprintf("Latency: %dms", r.AvgLatencyMs)
		}
		return "All checks passing"
	case types.StatusDegraded:
		return fmt.Sprintf("Loss: %.0f%%  Latency: %dms", r.PacketLossPct, r.AvgLatencyMs)
	default:
		parts := []string{}
		if r.TCPPingFailed {
			parts = append(parts, "TCP")
		}
		if r.HTTPFailed {
			parts = append(parts, "HTTP")
		}
		if r.DNSFailed {
			parts = append(parts, "DNS")
		}
		if len(parts) == 0 {
			return "Connection lost"
		}
		return strings.Join(parts, " + ") + " failed"
	}
}
