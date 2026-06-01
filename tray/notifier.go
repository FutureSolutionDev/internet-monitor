package tray

import (
	"internet-monitor/notifytext"
	"internet-monitor/types"
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
