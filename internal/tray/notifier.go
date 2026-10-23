package tray

import (
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/types"
	"sync"
	"time"
)

const notifyCooldown = 4 * time.Second

type TrayNotifier struct {
	t        *Tray
	mu       sync.Mutex
	lastTime time.Time
}

func NewNotifier(t *Tray) *TrayNotifier {
	return &TrayNotifier{t: t}
}

func (n *TrayNotifier) OnTick(result types.CheckResult, status types.Status) {
	n.t.updateTooltip(status, result)
}

func (n *TrayNotifier) OnEvent(event types.Event) {
	s := notifytext.StatusFromEventType(event.EventType)
	r := types.CheckResult{
		TCPPingOK:  !event.Reason.TCPPingFailed,
		HTTPOK:     !event.Reason.HTTPFailed,
		DNSOK:      !event.Reason.DNSFailed,
		LatencyMs:  event.Reason.AvgLatencyMs,
		PacketLoss: event.Reason.PacketLossPct,
	}
	n.t.applyStatus(s, r)

	title, body := notifytext.Build(notifytext.Lang(), s, r)

	n.mu.Lock()
	if time.Since(n.lastTime) < notifyCooldown {
		n.mu.Unlock()
		return
	}
	n.lastTime = time.Now()
	n.mu.Unlock()

	go Notify(title, body)
}
