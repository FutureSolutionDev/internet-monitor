package tray

import "internet-monitor/types"

// TrayNotifier updates the system tray icon and sends OS notifications.
type TrayNotifier struct {
	t *Tray
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
	go Notify(titleForStatus(s), bodyForEvent(event))
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
	parts := []string{}
	if r.TCPPingFailed {
		parts = append(parts, "TCP ping failed")
	}
	if r.HTTPFailed {
		parts = append(parts, "HTTP failed")
	}
	if r.DNSFailed {
		parts = append(parts, "DNS failed")
	}
	if len(parts) == 0 {
		return ""
	}
	msg := ""
	for i, p := range parts {
		if i > 0 {
			msg += ", "
		}
		msg += p
	}
	return msg
}
