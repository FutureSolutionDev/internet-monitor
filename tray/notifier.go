package tray

import "internet-monitor/types"

// TrayNotifier updates the system tray icon and sends OS notifications.
type TrayNotifier struct {
	t *Tray
}

func NewNotifier(t *Tray) *TrayNotifier {
	return &TrayNotifier{t: t}
}

func (n *TrayNotifier) OnTick(result types.CheckResult, status types.Status) {
	n.t.applyStatus(status, result)
}

func (n *TrayNotifier) OnEvent(event types.Event) {
	s := types.Status(0)
	switch event.EventType {
	case "connected":
		s = types.StatusConnected
	case "degraded":
		s = types.StatusDegraded
	default:
		s = types.StatusDisconnected
	}
	go Notify(titleForStatus(s), bodyForEvent(event))
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
