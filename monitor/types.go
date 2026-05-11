package monitor

import "time"

type Status int

const (
	StatusConnected    Status = iota
	StatusDegraded
	StatusDisconnected
)

func (s Status) String() string {
	switch s {
	case StatusConnected:
		return "connected"
	case StatusDegraded:
		return "degraded"
	default:
		return "disconnected"
	}
}

type CheckResult struct {
	TCPPingOK  bool
	HTTPOK     bool
	DNSOK      bool
	LatencyMs  int64
	PacketLoss float64
	Timestamp  time.Time
}

type Event struct {
	Timestamp       time.Time   `json:"timestamp"`
	EventType       string      `json:"event"`
	DurationSeconds float64     `json:"duration_seconds,omitempty"`
	Reason          EventReason `json:"reason"`
}

type EventReason struct {
	TCPPingFailed bool    `json:"tcp_ping_failed"`
	HTTPFailed    bool    `json:"http_failed"`
	DNSFailed     bool    `json:"dns_failed"`
	PacketLossPct float64 `json:"packet_loss_pct"`
	AvgLatencyMs  int64   `json:"avg_latency_ms"`
}
