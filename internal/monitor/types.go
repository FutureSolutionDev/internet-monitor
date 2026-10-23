package monitor

import "internet-monitor/internal/types"

type Status = types.Status
type CheckResult = types.CheckResult
type Event = types.Event
type EventReason = types.EventReason

const (
	StatusConnected    = types.StatusConnected
	StatusDegraded     = types.StatusDegraded
	StatusDisconnected = types.StatusDisconnected
)
