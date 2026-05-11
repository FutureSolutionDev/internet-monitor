package main

import (
	"internet-monitor/monitor"
	"internet-monitor/types"
)

// guiNotifier updates the background tray icon and sends OS notifications.
type guiNotifier struct{}

func (g *guiNotifier) OnTick(_ types.CheckResult, status types.Status) {
	updateTrayStatus(status)
}

func (g *guiNotifier) OnEvent(event types.Event) {
	var s monitor.Status
	switch event.EventType {
	case "connected":
		s = monitor.StatusConnected
	case "degraded":
		s = monitor.StatusDegraded
	default:
		s = monitor.StatusDisconnected
	}
	go sendNotification(s, monitor.CheckResult{
		TCPPingOK:  !event.Reason.TCPPingFailed,
		HTTPOK:     !event.Reason.HTTPFailed,
		DNSOK:      !event.Reason.DNSFailed,
		LatencyMs:  event.Reason.AvgLatencyMs,
		PacketLoss: event.Reason.PacketLossPct,
	})
}
