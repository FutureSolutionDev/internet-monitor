package main

import (
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/types"
)

type guiNotifier struct{}

func (g *guiNotifier) OnTick(_ types.CheckResult, status types.Status) {
	updateTrayStatus(status)
}

func (g *guiNotifier) OnEvent(event types.Event) {
	s := notifytext.StatusFromEventType(event.EventType)
	go sendNotification(s, monitor.CheckResult{
		TCPPingOK:  !event.Reason.TCPPingFailed,
		HTTPOK:     !event.Reason.HTTPFailed,
		DNSOK:      !event.Reason.DNSFailed,
		LatencyMs:  event.Reason.AvgLatencyMs,
		PacketLoss: event.Reason.PacketLossPct,
	})
}
