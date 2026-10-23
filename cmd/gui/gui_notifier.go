package main

import (
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/tray"
	"internet-monitor/internal/types"
	"sync"
	"time"
)

const notifyCooldown = 4 * time.Second

type guiNotifier struct {
	mu   sync.Mutex
	last time.Time
}

func (g *guiNotifier) OnTick(_ types.CheckResult, status types.Status) {
	updateTrayStatus(status)
}

func (g *guiNotifier) OnEvent(event types.Event) {
	status := notifytext.StatusFromEventType(event.EventType)
	title, body := notifytext.Build(notifytext.Lang(), status, types.CheckResult{
		TCPPingOK:  !event.Reason.TCPPingFailed,
		HTTPOK:     !event.Reason.HTTPFailed,
		DNSOK:      !event.Reason.DNSFailed,
		LatencyMs:  event.Reason.AvgLatencyMs,
		PacketLoss: event.Reason.PacketLossPct,
	})
	if title == "" {
		return
	}
	g.mu.Lock()
	if time.Since(g.last) < notifyCooldown {
		g.mu.Unlock()
		return
	}
	g.last = time.Now()
	g.mu.Unlock()
	go tray.Notify(title, body)
}

func TestNotification(lang string) {
	title, body := notifytext.TestMessage(lang)
	tray.Notify(title, body)
}
