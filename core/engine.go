package core

import (
	"internet-monitor/config"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"internet-monitor/notifytext"
	"internet-monitor/types"
	"internet-monitor/updater"
	"log"
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

// recoverFan recovers from a panic in a single child notifier so the remaining
// siblings still receive the tick/event.
func recoverFan() {
	if r := recover(); r != nil {
		log.Printf("[notifier] panic recovered: %v", r)
	}
}

func (m MultiNotifier) OnTick(r types.CheckResult, s types.Status) {
	for _, n := range m {
		if n == nil {
			continue
		}
		func(n Notifier) {
			defer recoverFan()
			n.OnTick(r, s)
		}(n)
	}
}

func (m MultiNotifier) OnEvent(e types.Event) {
	for _, n := range m {
		if n == nil {
			continue
		}
		func(n Notifier) {
			defer recoverFan()
			n.OnEvent(e)
		}(n)
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

	agg minuteAgg

	reload chan *config.Config
	stop   chan struct{}
	done   chan struct{}
	once   sync.Once
}

const logRetention = 90 * 24 * time.Hour

// minuteAgg accumulates check results within a one-minute bucket. Accessed only
// from the run goroutine, so it needs no locking.
type minuteAgg struct {
	bucket    time.Time
	samples   int
	up        int
	latSum    int64
	latMax    int64
	lossSum   float64
	tcpFails  int
	httpFails int
	dnsFails  int

	prevLat   int64
	havePrev  bool
	jitterSum int64
	jitterN   int
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
				notifytext.SetLang(cfg.Language) // live notifications follow language changes
				ticker.Reset(cfg.CheckInterval())
				e.lgr.AppLog("CONFIG reloaded: interval=%ds targets(ping=%d http=%d dns=%d)",
					cfg.CheckIntervalSec, len(cfg.PingTargets), len(cfg.HTTPTargets), len(cfg.DNSTargets))
			case <-e.stop:
				e.flushAgg() // persist the partial minute before exiting
				return
			}
		}
	}()

	// Daily retention cleanup of old log files.
	go func() {
		for {
			e.lgr.CleanupOldLogs(logRetention)
			select {
			case <-time.After(24 * time.Hour):
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

// accumulate folds a check result into the current minute bucket, flushing a
// completed minute to the metrics log. Runs only in the engine goroutine.
func (e *Engine) accumulate(r types.CheckResult, status types.Status) {
	min := r.Timestamp.Truncate(time.Minute)
	if e.agg.bucket.IsZero() {
		e.agg.bucket = min
	}
	if !min.Equal(e.agg.bucket) {
		e.flushAgg()
		e.agg = minuteAgg{bucket: min}
	}
	e.agg.samples++
	if status != types.StatusDisconnected {
		e.agg.up++
	}
	e.agg.latSum += r.LatencyMs
	if r.LatencyMs > e.agg.latMax {
		e.agg.latMax = r.LatencyMs
	}
	if e.agg.havePrev {
		d := r.LatencyMs - e.agg.prevLat
		if d < 0 {
			d = -d
		}
		e.agg.jitterSum += d
		e.agg.jitterN++
	}
	e.agg.prevLat = r.LatencyMs
	e.agg.havePrev = true
	e.agg.lossSum += r.PacketLoss
	if !r.TCPPingOK {
		e.agg.tcpFails++
	}
	if !r.HTTPOK {
		e.agg.httpFails++
	}
	if !r.DNSOK {
		e.agg.dnsFails++
	}
}

// flushAgg writes the current minute bucket (if non-empty) to the metrics log.
func (e *Engine) flushAgg() {
	if e.agg.samples == 0 || e.lgr == nil {
		return
	}
	jitter := int64(0)
	if e.agg.jitterN > 0 {
		jitter = e.agg.jitterSum / int64(e.agg.jitterN)
	}
	e.lgr.LogSample(types.MetricSample{
		Timestamp:    e.agg.bucket,
		Samples:      e.agg.samples,
		UpSamples:    e.agg.up,
		AvgLatencyMs: e.agg.latSum / int64(e.agg.samples),
		MaxLatencyMs: e.agg.latMax,
		JitterMs:     jitter,
		AvgLossPct:   e.agg.lossSum / float64(e.agg.samples),
		TCPFails:     e.agg.tcpFails,
		HTTPFails:    e.agg.httpFails,
		DNSFails:     e.agg.dnsFails,
	})
}

// safeNotify runs a notifier callback, recovering from a panic so one bad
// notifier can't take down the monitoring loop.
func (e *Engine) safeNotify(fn func()) {
	defer func() {
		if r := recover(); r != nil && e.lgr != nil {
			e.lgr.AppLog("PANIC recovered in notifier: %v", r)
		}
	}()
	fn()
}

func (e *Engine) runCheck() {
	result := e.checker.Check()
	newStatus := DetermineStatus(result, &e.consecFails, e.cfg)

	e.accumulate(result, newStatus)

	if e.Notifier != nil {
		e.safeNotify(func() { e.Notifier.OnTick(result, newStatus) })
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

		// First observation is a baseline, not a transition: only persist it if
		// the link starts unhealthy (so a "started while down" is recorded), and
		// never fire a notification for it.
		isFirst := e.currentStatus == nil
		if !isFirst || newStatus != types.StatusConnected {
			e.lgr.Log(event)
		}
		if !isFirst && e.Notifier != nil {
			e.safeNotify(func() { e.Notifier.OnEvent(event) })
		}

		s := newStatus
		e.currentStatus = &s
	}
}
