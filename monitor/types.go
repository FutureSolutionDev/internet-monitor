package monitor

import "internet-monitor/types"

// Type aliases — existing code importing from monitor continues to work unchanged.
type Status = types.Status
type CheckResult = types.CheckResult
type Event = types.Event
type EventReason = types.EventReason

const (
	StatusConnected    = types.StatusConnected
	StatusDegraded     = types.StatusDegraded
	StatusDisconnected = types.StatusDisconnected
)
