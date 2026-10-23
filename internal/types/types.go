package types

import "time"

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

type TargetResult struct {
	Layer     string `json:"layer"`
	Target    string `json:"target"`
	OK        bool   `json:"ok"`
	LatencyMs int64  `json:"latency_ms"`
}

type CheckResult struct {
	TCPPingOK  bool
	HTTPOK     bool
	DNSOK      bool
	LatencyMs  int64
	PacketLoss float64
	Timestamp  time.Time
	Targets    []TargetResult `json:"targets,omitempty"`
	Gateway    string         `json:"gateway,omitempty"`
	GatewayOK  bool           `json:"gateway_ok,omitempty"`
}

func Diagnose(r CheckResult) string {
	switch {
	case r.TCPPingOK && r.HTTPOK && r.DNSOK:
		return "ok"
	case !r.TCPPingOK:
		return "down"
	case !r.DNSOK:
		return "dns"
	case !r.HTTPOK:
		return "http"
	default:
		return "partial"
	}
}

type EventReason struct {
	TCPPingFailed bool    `json:"tcp_ping_failed"`
	HTTPFailed    bool    `json:"http_failed"`
	DNSFailed     bool    `json:"dns_failed"`
	PacketLossPct float64 `json:"packet_loss_pct"`
	AvgLatencyMs  int64   `json:"avg_latency_ms"`
}

type Event struct {
	Timestamp       time.Time   `json:"timestamp"`
	EventType       string      `json:"event"`
	DurationSeconds float64     `json:"duration_seconds,omitempty"`
	Reason          EventReason `json:"reason"`
}

type MetricSample struct {
	Timestamp    time.Time `json:"timestamp"`
	Samples      int       `json:"samples"`
	UpSamples    int       `json:"up_samples"`
	AvgLatencyMs int64     `json:"avg_latency_ms"`
	MaxLatencyMs int64     `json:"max_latency_ms"`
	JitterMs     int64     `json:"jitter_ms"`
	AvgLossPct   float64   `json:"avg_loss_pct"`
	TCPFails     int       `json:"tcp_fails"`
	HTTPFails    int       `json:"http_fails"`
	DNSFails     int       `json:"dns_fails"`
}
