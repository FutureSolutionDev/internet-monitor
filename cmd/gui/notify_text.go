package main

import (
	"fmt"
	"internet-monitor/monitor"
	"strings"
)

func notifyText(status monitor.Status, result monitor.CheckResult) (title, body string) {
	switch status {
	case monitor.StatusConnected:
		return "Internet Restored", fmt.Sprintf("Latency: %dms", result.LatencyMs)
	case monitor.StatusDisconnected:
		parts := []string{}
		if !result.TCPPingOK {
			parts = append(parts, "TCP ping")
		}
		if !result.HTTPOK {
			parts = append(parts, "HTTP")
		}
		if !result.DNSOK {
			parts = append(parts, "DNS")
		}
		body = "Connection lost"
		if len(parts) > 0 {
			body = strings.Join(parts, ", ") + " failed"
		}
		return "Internet Disconnected", body
	case monitor.StatusDegraded:
		return "Connection Degraded", fmt.Sprintf("Loss: %.0f%%  Latency: %dms", result.PacketLoss, result.LatencyMs)
	}
	return "", ""
}
