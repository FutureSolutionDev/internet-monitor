package core

import (
	"internet-monitor/config"
	"internet-monitor/types"
)

// DetermineStatus applies threshold rules to a check result.
// Pure function with no side effects.
func DetermineStatus(result types.CheckResult, consecFails *int, cfg *config.Config) types.Status {
	if !result.TCPPingOK || !result.HTTPOK || !result.DNSOK {
		*consecFails++
	} else {
		*consecFails = 0
	}
	if *consecFails >= cfg.FailThreshold {
		return types.StatusDisconnected
	}
	if result.PacketLoss > cfg.PacketLossThreshold ||
		(result.LatencyMs > int64(cfg.LatencyThreshold) && result.LatencyMs > 0) {
		return types.StatusDegraded
	}
	return types.StatusConnected
}
