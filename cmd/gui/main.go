//go:generate go-winres simply --arch amd64 --icon ../../dashboard/assets/favicon.png

package main

import (
	"internet-monitor/config"
	"internet-monitor/dashboard"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"log"
	"os"
	"path/filepath"
	"time"

	webview "github.com/webview/webview_go"
)

func main() {
	if exePath, err := os.Executable(); err == nil {
		os.Chdir(filepath.Dir(exePath))
	}

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("config error: %v — using defaults", err)
		cfg = &config.Default
	}

	lgr, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	dash := dashboard.NewServer(cfg.DashboardPort, "config.json", cfg.LogDir)
	dash.OnTestNotification = TestNotification
	dash.Start()

	checker := monitor.NewChecker(cfg)

	// Create webview window
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("مراقب الإنترنت — Internet Monitor")
	w.SetSize(1100, 750, webview.HintNone)

	// Get native window handle (HWND on Windows)
	hwnd := uintptr(w.Window())

	// Expose minimize-to-tray to JavaScript
	w.Bind("nativeMinimizeToTray", func() {
		hideWindow(hwnd)
	})

	// Start system tray (Windows: real tray; other OS: no-op)
	stopTray := initTray(w, hwnd)
	defer stopTray()

	// Start monitoring loop
	go monitoringLoop(cfg, checker, lgr, dash)

	time.Sleep(150 * time.Millisecond)
	w.Navigate(dash.URL())
	w.Run()
}

func monitoringLoop(cfg *config.Config, checker *monitor.Checker, lgr *logger.Logger, dash *dashboard.Server) {
	var currentStatus *monitor.Status
	var statusSince time.Time
	consecutiveFails := 0

	doCheck := func() {
		result := checker.Check()
		newStatus := determineStatus(result, &consecutiveFails, cfg)

		dash.UpdateTick(result, newStatus)

		// Always keep tray icon in sync with current status
		updateTrayStatus(newStatus)

		if currentStatus == nil || *currentStatus != newStatus {
			duration := 0.0
			if !statusSince.IsZero() {
				duration = time.Since(statusSince).Seconds()
			}
			statusSince = time.Now()

			event := monitor.Event{
				Timestamp:       result.Timestamp,
				EventType:       newStatus.String(),
				DurationSeconds: duration,
				Reason: monitor.EventReason{
					TCPPingFailed: !result.TCPPingOK,
					HTTPFailed:    !result.HTTPOK,
					DNSFailed:     !result.DNSOK,
					PacketLossPct: result.PacketLoss,
					AvgLatencyMs:  result.LatencyMs,
				},
			}
			lgr.Log(event)
			dash.AddEvent(event)

			if currentStatus != nil {
				go sendNotification(newStatus, result)
			}
			s := newStatus
			currentStatus = &s
		}
	}

	doCheck()
	ticker := time.NewTicker(cfg.CheckInterval())
	defer ticker.Stop()
	for range ticker.C {
		doCheck()
	}
}

func determineStatus(result monitor.CheckResult, consecutiveFails *int, cfg *config.Config) monitor.Status {
	if !result.TCPPingOK || !result.HTTPOK || !result.DNSOK {
		*consecutiveFails++
	} else {
		*consecutiveFails = 0
	}
	if *consecutiveFails >= cfg.FailThreshold {
		return monitor.StatusDisconnected
	}
	if result.PacketLoss > cfg.PacketLossThreshold ||
		(result.LatencyMs > int64(cfg.LatencyThreshold) && result.LatencyMs > 0) {
		return monitor.StatusDegraded
	}
	return monitor.StatusConnected
}
