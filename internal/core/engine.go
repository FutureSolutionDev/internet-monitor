package core

import (
	"context"
	"internet-monitor/internal/config"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/types"
	"log"
	"sync"
	"time"
)

type Notifier interface {
	OnTick(result types.CheckResult, status types.Status)

	OnEvent(event types.Event)
}

type MultiNotifier []Notifier

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

type Engine struct {
	cfg     config.Config
	checker *monitor.Checker
	lgr     *logger.Logger
	version string

	Notifier Notifier

	ctx    context.Context
	cancel context.CancelFunc

	currentStatus *types.Status
	statusSince   time.Time
	consecFails   int

	agg minuteAgg

	reload chan config.Config
	stop   chan struct{}
	done   chan struct{}
	once   sync.Once
}

const logRetention = 90 * 24 * time.Hour

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

func New(cfg config.Config, checker *monitor.Checker, lgr *logger.Logger, version string) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		cfg:     cfg.Clone(),
		checker: checker,
		lgr:     lgr,
		version: version,
		ctx:     ctx,
		cancel:  cancel,
		reload:  make(chan config.Config, 1),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (e *Engine) ApplyConfig(cfg config.Config) {
	select {
	case <-e.reload:
	default:
	}
	select {
	case e.reload <- cfg:
	default:
	}
}

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
				notifytext.SetLang(cfg.Language)
				ticker.Reset(cfg.CheckInterval())
				e.lgr.AppLog("CONFIG reloaded: interval=%ds targets(ping=%d http=%d dns=%d)",
					cfg.CheckIntervalSec, len(cfg.PingTargets), len(cfg.HTTPTargets), len(cfg.DNSTargets))
			case <-e.stop:
				e.flushAgg()
				return
			}
		}
	}()

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

}

func (e *Engine) Stop() {
	e.once.Do(func() {
		close(e.stop)
		e.cancel()
	})
	select {
	case <-e.done:
	case <-time.After(5 * time.Second):
	}
}

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

func (e *Engine) safeNotify(fn func()) {
	defer func() {
		if r := recover(); r != nil && e.lgr != nil {
			e.lgr.AppLog("PANIC recovered in notifier: %v", r)
		}
	}()
	fn()
}

func (e *Engine) runCheck() {
	result := e.checker.Check(e.ctx)
	if e.ctx.Err() != nil {
		return
	}
	newStatus := DetermineStatus(result, &e.consecFails, &e.cfg)

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
