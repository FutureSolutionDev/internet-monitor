package tray

import (
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/types"
)

type TrayNotifier struct {
	t *Tray
}

func NewNotifier(t *Tray) *TrayNotifier {
	return &TrayNotifier{t: t}
}

func (n *TrayNotifier) OnTick(result types.CheckResult, status types.Status) {
	n.t.updateTooltip(status, result)
}

func (n *TrayNotifier) OnEvent(event types.Event) {
	n.t.applyStatus(notifytext.StatusFromEventType(event.EventType), types.CheckResult{
		TCPPingOK:  !event.Reason.TCPPingFailed,
		HTTPOK:     !event.Reason.HTTPFailed,
		DNSOK:      !event.Reason.DNSFailed,
		LatencyMs:  event.Reason.AvgLatencyMs,
		PacketLoss: event.Reason.PacketLossPct,
	})
}
