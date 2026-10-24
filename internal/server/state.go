package server

import (
	"internet-monitor/internal/types"
	"internet-monitor/internal/updater"
	"net/http"
	"sync"
	"time"
)

const (
	maxHistory = 60
	maxEvents  = 100
	maxTicks   = 20
)

type TickEntry struct {
	Time       string  `json:"time"`
	Timestamp  int64   `json:"ts"`
	Status     string  `json:"status"`
	LatencyMs  int64   `json:"latency_ms"`
	PacketLoss float64 `json:"packet_loss_pct"`
	TCPOk      bool    `json:"tcp_ok"`
	HTTPOk     bool    `json:"http_ok"`
	DNSOk      bool    `json:"dns_ok"`
}

type EventEntry struct {
	Time       string  `json:"time"`
	Timestamp  int64   `json:"ts"`
	EventType  string  `json:"event_type"`
	Duration   float64 `json:"duration_seconds"`
	TCPFailed  bool    `json:"tcp_failed"`
	HTTPFailed bool    `json:"http_failed"`
	DNSFailed  bool    `json:"dns_failed"`
	PacketLoss float64 `json:"packet_loss_pct"`
	LatencyMs  int64   `json:"latency_ms"`
}

type Snapshot struct {
	Type             string               `json:"type"`
	BootID           string               `json:"boot_id"`
	Version          string               `json:"version"`
	ServerTime       int64                `json:"server_time"`
	ConfigRevision   uint64               `json:"config_revision"`
	ConfigError      string               `json:"config_error,omitempty"`
	Language         string               `json:"language"`
	CheckIntervalSec int                  `json:"check_interval_sec"`
	Status           string               `json:"status"`
	StatusSince      int64                `json:"status_since"`
	LastCheck        int64                `json:"last_check"`
	LatencyMs        int64                `json:"latency_ms"`
	AvgLatencyMs     int64                `json:"avg_latency_ms"`
	PacketLoss       float64              `json:"packet_loss"`
	TCPPingOK        bool                 `json:"tcp_ping_ok"`
	HTTPOK           bool                 `json:"http_ok"`
	DNSOK            bool                 `json:"dns_ok"`
	Diagnosis        string               `json:"diagnosis"`
	Gateway          string               `json:"gateway,omitempty"`
	TotalChecks      int                  `json:"total_checks"`
	UpChecks         int                  `json:"up_checks"`
	Disconnections   int                  `json:"disconnections"`
	UptimeSeconds    float64              `json:"uptime_seconds"`
	UptimePct        float64              `json:"uptime_pct"`
	JitterMs         int64                `json:"jitter_ms"`
	LatencyHistory   []int64              `json:"latency_history"`
	Targets          []types.TargetResult `json:"targets"`
	Events           []EventEntry         `json:"events"`
	Ticks            []TickEntry          `json:"ticks"`
	Availability     *Availability        `json:"availability,omitempty"`
	UpdateInfo       *updater.Info        `json:"update_info,omitempty"`
	UpdateStatus     string               `json:"update_status"`
	UpdateError      string               `json:"update_error,omitempty"`
	SystemNotifs     bool                 `json:"system_notifs"`
	SpeedTestRunning bool                 `json:"speed_test_running"`
}

type liveState struct {
	mu             sync.RWMutex
	started        time.Time
	status         string
	statusSince    time.Time
	lastCheck      time.Time
	result         types.CheckResult
	totalChecks    int
	upTicks        int
	disconnections int
	latSum         int64
	latCount       int64
	latencyHistory []int64
	events         []EventEntry
	ticks          []TickEntry
}

func newLiveState() *liveState {
	now := time.Now()
	return &liveState{started: now, statusSince: now, status: "checking"}
}

func (s *Server) UpdateTick(result types.CheckResult, status types.Status) {
	l := s.live
	l.mu.Lock()
	if l.status != status.String() {
		l.status = status.String()
		l.statusSince = result.Timestamp
	}
	l.lastCheck = result.Timestamp
	l.result = result
	l.totalChecks++
	if status != types.StatusDisconnected {
		l.upTicks++
	}
	if result.LatencyMs > 0 {
		l.latSum += result.LatencyMs
		l.latCount++
	}
	l.latencyHistory = append(l.latencyHistory, result.LatencyMs)
	if len(l.latencyHistory) > maxHistory {
		l.latencyHistory = l.latencyHistory[len(l.latencyHistory)-maxHistory:]
	}
	tick := TickEntry{
		Time:       result.Timestamp.Format("15:04:05"),
		Timestamp:  result.Timestamp.UnixMilli(),
		Status:     status.String(),
		LatencyMs:  result.LatencyMs,
		PacketLoss: result.PacketLoss,
		TCPOk:      result.TCPPingOK,
		HTTPOk:     result.HTTPOK,
		DNSOk:      result.DNSOK,
	}
	l.ticks = append([]TickEntry{tick}, l.ticks...)
	if len(l.ticks) > maxTicks {
		l.ticks = l.ticks[:maxTicks]
	}
	l.mu.Unlock()
	s.broadcast("tick")
}

func (s *Server) AddEvent(event types.Event) {
	entry := EventEntry{
		Time:       event.Timestamp.Format("15:04:05"),
		Timestamp:  event.Timestamp.UnixMilli(),
		EventType:  event.EventType,
		Duration:   event.DurationSeconds,
		TCPFailed:  event.Reason.TCPPingFailed,
		HTTPFailed: event.Reason.HTTPFailed,
		DNSFailed:  event.Reason.DNSFailed,
		PacketLoss: event.Reason.PacketLossPct,
		LatencyMs:  event.Reason.AvgLatencyMs,
	}
	l := s.live
	l.mu.Lock()
	l.events = append([]EventEntry{entry}, l.events...)
	if len(l.events) > maxEvents {
		l.events = l.events[:maxEvents]
	}
	if event.EventType == types.StatusDisconnected.String() {
		l.disconnections++
	}
	l.mu.Unlock()
	s.broadcast("event")
}

func (s *Server) snapshot(msgType string) Snapshot {
	cfg := s.config()
	var rev uint64
	var cfgErr string
	if s.store != nil {
		rev = s.store.Revision()
		if err := s.store.LoadError(); err != nil {
			cfgErr = err.Error()
		}
	}
	info, upStatus, upErr := s.update.get()

	l := s.live
	l.mu.RLock()
	defer l.mu.RUnlock()

	r := l.result
	uptimePct := 0.0
	if l.totalChecks > 0 {
		uptimePct = float64(l.upTicks) / float64(l.totalChecks) * 100
	}
	var avg int64
	if l.latCount > 0 {
		avg = l.latSum / l.latCount
	}
	diag := types.Diagnose(r)
	if l.totalChecks == 0 {
		diag = "checking"
	} else if diag == "down" && r.Gateway != "" {
		diag = "lan"
		if r.GatewayOK {
			diag = "isp"
		}
	}
	var lastCheck int64
	if !l.lastCheck.IsZero() {
		lastCheck = l.lastCheck.UnixMilli()
	}

	return Snapshot{
		Type:             msgType,
		BootID:           s.bootID,
		Version:          s.version,
		ServerTime:       time.Now().UnixMilli(),
		ConfigRevision:   rev,
		ConfigError:      cfgErr,
		Language:         cfg.Language,
		CheckIntervalSec: cfg.CheckIntervalSec,
		Status:           l.status,
		StatusSince:      l.statusSince.UnixMilli(),
		LastCheck:        lastCheck,
		LatencyMs:        r.LatencyMs,
		AvgLatencyMs:     avg,
		PacketLoss:       r.PacketLoss,
		TCPPingOK:        r.TCPPingOK,
		HTTPOK:           r.HTTPOK,
		DNSOK:            r.DNSOK,
		Diagnosis:        diag,
		Gateway:          r.Gateway,
		TotalChecks:      l.totalChecks,
		UpChecks:         l.upTicks,
		Disconnections:   l.disconnections,
		UptimeSeconds:    time.Since(l.started).Seconds(),
		UptimePct:        uptimePct,
		JitterMs:         jitterOf(l.latencyHistory),
		LatencyHistory:   append([]int64{}, l.latencyHistory...),
		Targets:          append([]types.TargetResult{}, r.Targets...),
		Events:           append([]EventEntry{}, l.events...),
		Ticks:            append([]TickEntry{}, l.ticks...),
		Availability:     s.avail.get(),
		UpdateInfo:       info,
		UpdateStatus:     upStatus,
		UpdateError:      upErr,
		SystemNotifs:     s.native,
		SpeedTestRunning: s.speed.running.Load(),
	}
}

func jitterOf(hist []int64) int64 {
	if len(hist) < 2 {
		return 0
	}
	var sum, n int64
	for i := 1; i < len(hist); i++ {
		d := hist[i] - hist[i-1]
		if d < 0 {
			d = -d
		}
		sum += d
		n++
	}
	return sum / n
}

func (s *Server) serveState(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.snapshot("init"))
}

func (s *Server) serveVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": s.version, "boot_id": s.bootID})
}
