package server

import (
	"context"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/speedtest"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type speedRunner struct {
	s       *Server
	running atomic.Bool
	mu      sync.Mutex
	stop    context.CancelFunc
	last    *logger.SpeedTestEvent
}

type speedProgress struct {
	Type           string                 `json:"type"`
	Phase          string                 `json:"phase,omitempty"`
	Done           bool                   `json:"done"`
	Cancelled      bool                   `json:"cancelled,omitempty"`
	Error          string                 `json:"error,omitempty"`
	CurrentMbps    float64                `json:"current_mbps,omitempty"`
	ElapsedSeconds float64                `json:"elapsed_seconds,omitempty"`
	TotalSeconds   float64                `json:"total_seconds,omitempty"`
	LatencyMs      int64                  `json:"latency_ms,omitempty"`
	DownloadMbps   float64                `json:"download_mbps,omitempty"`
	Result         *logger.SpeedTestEvent `json:"result,omitempty"`
}

func (sr *speedRunner) emit(p speedProgress) {
	p.Type = "speed_test_progress"
	sr.s.broadcastJSON(p)
}

func (sr *speedRunner) lastResult() *logger.SpeedTestEvent {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.last
}

func (sr *speedRunner) cancel() {
	sr.mu.Lock()
	stop := sr.stop
	sr.mu.Unlock()
	if stop != nil {
		stop()
	}
}

func (sr *speedRunner) start(triggeredBy string) bool {
	if !sr.running.CompareAndSwap(false, true) {
		return false
	}
	cfg := sr.s.config()
	ctx, cancel := context.WithCancel(context.Background())
	sr.mu.Lock()
	sr.stop = cancel
	sr.mu.Unlock()

	stCfg := speedtest.Config{
		Endpoints:    cfg.SpeedTest.DownloadTargets,
		Parallel:     cfg.SpeedTest.ParallelConnections,
		Timeout:      time.Duration(cfg.SpeedTest.TimeoutSeconds) * time.Second,
		UploadTarget: cfg.SpeedTest.UploadTarget,
	}
	total := stCfg.Timeout.Seconds()
	sr.s.broadcast("speed_state")

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				sr.s.logf("PANIC recovered in speed test: %v", rec)
				sr.emit(speedProgress{Phase: "error", Done: true, Error: "internal_error"})
			}
			cancel()
			sr.mu.Lock()
			sr.stop = nil
			sr.mu.Unlock()
			sr.running.Store(false)
			sr.s.broadcast("speed_state")
		}()

		sr.emit(speedProgress{Phase: "ping"})
		result, err := speedtest.Run(ctx, stCfg, func(mbps float64, elapsed time.Duration) {
			sr.emit(speedProgress{Phase: "download", CurrentMbps: mbps, ElapsedSeconds: elapsed.Seconds(), TotalSeconds: total})
		})
		if ctx.Err() != nil && (result == nil || result.DownloadMbps == 0) {
			sr.emit(speedProgress{Phase: "cancelled", Done: true, Cancelled: true})
			return
		}
		if err != nil || result == nil {
			msg := "no_result"
			if err != nil {
				msg = err.Error()
			}
			sr.s.logf("SPEEDTEST failed (%s): %s", triggeredBy, msg)
			sr.emit(speedProgress{Phase: "error", Done: true, Error: msg})
			return
		}
		sr.emit(speedProgress{Phase: "ping", LatencyMs: result.LatencyMs})

		var uploadMbps *float64
		if stCfg.UploadTarget != "" {
			mbps, uerr := speedtest.MeasureUpload(ctx, stCfg, func(m float64, elapsed time.Duration) {
				sr.emit(speedProgress{Phase: "upload", CurrentMbps: m, ElapsedSeconds: elapsed.Seconds(), TotalSeconds: total})
			})
			if uerr == nil {
				uploadMbps = &mbps
			} else {
				sr.s.logf("SPEEDTEST upload failed: %v", uerr)
			}
		}

		if ctx.Err() != nil {
			sr.emit(speedProgress{Phase: "cancelled", Done: true, Cancelled: true})
			return
		}

		event := logger.SpeedTestEvent{
			Timestamp:       time.Now().UTC(),
			Event:           "speed_test",
			DownloadMbps:    result.DownloadMbps,
			UploadMbps:      uploadMbps,
			LatencyMs:       result.LatencyMs,
			DurationSeconds: result.DurationSeconds,
			Endpoints:       result.Endpoints,
			ParallelConns:   result.ParallelConns,
			TriggeredBy:     triggeredBy,
		}
		if lgr := sr.s.lgr; lgr != nil {
			lgr.LogSpeedTest(event)
			if cfg.WebhookURL != "" || (cfg.TelegramBotToken != "" && cfg.TelegramChatID != "") {
				below := cfg.SpeedTest.AlertThresholdMbps > 0 && result.DownloadMbps < cfg.SpeedTest.AlertThresholdMbps
				lgr.SendSpeedTestResult(cfg.WebhookURL, event, cfg.SpeedTest.AlertThresholdMbps, below)
			}
		}
		sr.mu.Lock()
		sr.last = &event
		sr.mu.Unlock()
		sr.emit(speedProgress{Phase: "done", Done: true, DownloadMbps: result.DownloadMbps, LatencyMs: result.LatencyMs, Result: &event})
	}()
	return true
}

func (sr *speedRunner) schedule(stop <-chan struct{}) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	var lastRun time.Time
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			every := sr.s.config().SpeedTest.ScheduleMinutes
			if every <= 0 || time.Since(lastRun) < time.Duration(every)*time.Minute {
				continue
			}
			if sr.start("schedule") {
				lastRun = time.Now()
			}
		}
	}
}

func (s *Server) serveSpeedTestStart(w http.ResponseWriter, _ *http.Request) {
	if !s.speed.start("user") {
		writeError(w, http.StatusConflict, "test_already_running", "a speed test is already running")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}

func (s *Server) serveSpeedTestCancel(w http.ResponseWriter, _ *http.Request) {
	s.speed.cancel()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
