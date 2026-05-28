package core

import (
	"internet-monitor/config"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"internet-monitor/types"
	"internet-monitor/updater"
	"sync"
	"time"
)

// Notifier receives monitoring events. Implementations must be goroutine-safe.
type Notifier interface {
	// OnTick is called every check cycle regardless of status change.
	OnTick(result types.CheckResult, status types.Status)
	// OnEvent is called only when connectivity status changes.
	OnEvent(event types.Event)
}

// MultiNotifier fans out to multiple Notifier implementations in order.
type MultiNotifier []Notifier

func (m MultiNotifier) OnTick(r types.CheckResult, s types.Status) {
	for _, n := range m {
		if n != nil {
			n.OnTick(r, s)
		}
	}
}

func (m MultiNotifier) OnEvent(e types.Event) {
	for _, n := range m {
		if n != nil {
			n.OnEvent(e)
		}
	}
}

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

// Engine runs the monitoring loop and dispatches events to registered Notifiers.
type Engine struct {
	cfg     *config.Config
	checker *monitor.Checker
	lgr     *logger.Logger
	version string

	Notifier          Notifier
	OnUpdateAvailable func(*updater.Info)

	currentStatus *types.Status
	statusSince   time.Time
	consecFails   int

	reload chan *config.Config
	stop   chan struct{}
	done   chan struct{}
	once   sync.Once
}

// New creates a monitoring engine. Call Start() to begin monitoring.
func New(cfg *config.Config, checker *monitor.Checker, lgr *logger.Logger, version string) *Engine {
	return &Engine{
		cfg:     cfg,
		checker: checker,
		lgr:     lgr,
		version: version,
		reload:  make(chan *config.Config, 1),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// ApplyConfig hot-applies a new configuration to the running engine. The change
// is handed to the run goroutine, which owns all config reads, so it takes
// effect (new targets/thresholds/interval/webhook) without a restart and without
// data races. Non-blocking: a stale pending reload is replaced by the latest.
func (e *Engine) ApplyConfig(cfg *config.Config) {
	if cfg == nil {
		return
	}
	select {
	case <-e.reload:
	default:
	}
	select {
	case e.reload <- cfg:
	default:
	}
}

// Start launches the check ticker and update-checker goroutines. Non-blocking.
func (e *Engine) Start() {
	go func() {
		defer close(e.done)
		e.runCheck()
		ticker := time.NewTicker(e.cfg.CheckInterval())
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				e.runCheck()
			case cfg := <-e.reload:
				e.cfg = cfg
				e.checker.SetConfig(cfg)
				e.lgr.SetConfig(cfg)
				ticker.Reset(cfg.CheckInterval())
				e.lgr.AppLog("CONFIG reloaded: interval=%ds targets(ping=%d http=%d dns=%d)",
					cfg.CheckIntervalSec, len(cfg.PingTargets), len(cfg.HTTPTargets), len(cfg.DNSTargets))
			case <-e.stop:
				return
			}
		}
	}()

	go func() {
		select {
		case <-time.After(30 * time.Second):
		case <-e.stop:
			return
		}
		for {
			if info, err := updater.Check(e.version); err == nil && info.HasUpdate {
				e.lgr.AppLog("UPDATE available: %s (current: %s)", info.LatestVersion, info.CurrentVersion)
				if e.OnUpdateAvailable != nil {
					e.OnUpdateAvailable(info)
				}
			}
			select {
			case <-time.After(6 * time.Hour):
			case <-e.stop:
				return
			}
		}
	}()
}

// Stop signals all engine goroutines to exit. Blocks up to 2 seconds.
func (e *Engine) Stop() {
	e.once.Do(func() { close(e.stop) })
	select {
	case <-e.done:
	case <-time.After(2 * time.Second):
	}
}

func (e *Engine) runCheck() {
	result := e.checker.Check()
	newStatus := DetermineStatus(result, &e.consecFails, e.cfg)

	if e.Notifier != nil {
		e.Notifier.OnTick(result, newStatus)
	}

	if e.currentStatus == nil || *e.currentStatus != newStatus {
		duration := 0.0
		if !e.statusSince.IsZero() {
			duration = time.Since(e.statusSince).Seconds()
		}
		e.statusSince = time.Now()

		event := types.Event{
			Timestamp:       result.Timestamp,
			EventType:       newStatus.String(),
			DurationSeconds: duration,
			Reason: types.EventReason{
				TCPPingFailed: !result.TCPPingOK,
				HTTPFailed:    !result.HTTPOK,
				DNSFailed:     !result.DNSOK,
				PacketLossPct: result.PacketLoss,
				AvgLatencyMs:  result.LatencyMs,
			},
		}

		e.lgr.Log(event)

		if e.currentStatus != nil && e.Notifier != nil {
			e.Notifier.OnEvent(event)
		}

		s := newStatus
		e.currentStatus = &s
	}
}
