package types

import "time"

// Status represents the current connectivity state.
type Status int

const (
	StatusConnected Status = iota
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

// CheckResult holds the outcome of a single check cycle.
type CheckResult struct {
	TCPPingOK  bool
	HTTPOK     bool
	DNSOK      bool
	LatencyMs  int64
	PacketLoss float64
	Timestamp  time.Time
}

// Diagnose classifies a check result into a coarse failure type, helping
// distinguish a DNS problem from HTTP filtering from a full outage.
// Returns one of: "ok", "down", "dns", "http", "partial".
func Diagnose(r CheckResult) string {
	switch {
	case r.TCPPingOK && r.HTTPOK && r.DNSOK:
		return "ok"
	case !r.TCPPingOK:
		return "down" // no L4 reachability — likely router/ISP
	case !r.DNSOK:
		return "dns" // reachable but name resolution failing
	case !r.HTTPOK:
		return "http" // reachable, DNS ok, but HTTP blocked/captive portal
	default:
		return "partial"
	}
}

// EventReason contains per-layer failure details for a connectivity event.
type EventReason struct {
	TCPPingFailed bool    `json:"tcp_ping_failed"`
	HTTPFailed    bool    `json:"http_failed"`
	DNSFailed     bool    `json:"dns_failed"`
	PacketLossPct float64 `json:"packet_loss_pct"`
	AvgLatencyMs  int64   `json:"avg_latency_ms"`
}

// Event represents a connectivity state change with cause and duration.
type Event struct {
	Timestamp       time.Time   `json:"timestamp"`
	EventType       string      `json:"event"`
	DurationSeconds float64     `json:"duration_seconds,omitempty"`
	Reason          EventReason `json:"reason"`
}

// MetricSample is a one-minute aggregate of check results, persisted to
// metrics_YYYY-MM-DD.jsonl and used to build rich monthly reports.
type MetricSample struct {
	Timestamp    time.Time `json:"timestamp"`  // start of the minute bucket
	Samples      int       `json:"samples"`    // checks in this minute
	UpSamples    int       `json:"up_samples"` // checks that were not disconnected
	AvgLatencyMs int64     `json:"avg_latency_ms"`
	MaxLatencyMs int64     `json:"max_latency_ms"`
	JitterMs     int64     `json:"jitter_ms"`
	AvgLossPct   float64   `json:"avg_loss_pct"`
	TCPFails     int       `json:"tcp_fails"`
	HTTPFails    int       `json:"http_fails"`
	DNSFails     int       `json:"dns_fails"`
}
