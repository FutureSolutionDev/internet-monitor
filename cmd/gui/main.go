package main

import (
	"fmt"
	"internet-monitor/config"
	"internet-monitor/dashboard"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	webview "github.com/webview/webview_go"
)

func main() {
	// Always resolve paths relative to the executable
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
	dash.Start()

	checker := monitor.NewChecker(cfg)
	go monitoringLoop(cfg, checker, lgr, dash)

	// Give the HTTP server a moment to bind
	time.Sleep(150 * time.Millisecond)

	// Open native window — blocks until the window is closed
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("مراقب الإنترنت — Internet Monitor")
	w.SetSize(1100, 750, webview.HintNone)
	w.Navigate(dash.URL())
	w.Run()
}

// ── Monitoring loop (extracted from tray package) ─────────────

func monitoringLoop(cfg *config.Config, checker *monitor.Checker, lgr *logger.Logger, dash *dashboard.Server) {
	var currentStatus *monitor.Status
	var statusSince time.Time
	consecutiveFails := 0

	doCheck := func() {
		result := checker.Check()
		newStatus := determineStatus(result, &consecutiveFails, cfg)

		dash.UpdateTick(result, newStatus)

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

// ── Cross-platform notifications ──────────────────────────────

func sendNotification(status monitor.Status, result monitor.CheckResult) {
	title, body := notifyText(status, result)
	if title == "" {
		return
	}
	switch detectOS() {
	case "windows":
		notifyWindows(title, body)
	case "darwin":
		notifyMac(title, body)
	default:
		notifyLinux(title, body)
	}
}

func notifyText(status monitor.Status, result monitor.CheckResult) (string, string) {
	switch status {
	case monitor.StatusConnected:
		return "Internet Restored", fmt.Sprintf("Latency: %dms", result.LatencyMs)
	case monitor.StatusDisconnected:
		parts := []string{}
		if !result.TCPPingOK {
			parts = append(parts, "TCP ping")
		}
		if !result.HTTPOK {
			parts = append(parts, "HTTP")
		}
		if !result.DNSOK {
			parts = append(parts, "DNS")
		}
		body := "Connection lost"
		if len(parts) > 0 {
			body = strings.Join(parts, ", ") + " failed"
		}
		return "Internet Disconnected", body
	case monitor.StatusDegraded:
		return "Connection Degraded", fmt.Sprintf("Loss: %.0f%%  Latency: %dms", result.PacketLoss, result.LatencyMs)
	}
	return "", ""
}

func notifyWindows(title, body string) {
	title = strings.ReplaceAll(title, "'", "''")
	body = strings.ReplaceAll(body, "'", "''")
	script := fmt.Sprintf(`
$app='Internet Monitor'
[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null
$tpl=[Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$nodes=$tpl.GetElementsByTagName('text')
$nodes[0].InnerText='%s'
$nodes[1].InnerText='%s'
$toast=[Windows.UI.Notifications.ToastNotification]::new($tpl)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($app).Show($toast)
`, title, body)
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func notifyMac(title, body string) {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
	exec.Command("osascript", "-e", script).Start()
}

func notifyLinux(title, body string) {
	exec.Command("notify-send", title, body).Start()
}

func detectOS() string {
	if _, err := os.Stat("/System/Library/CoreServices/SystemVersion.plist"); err == nil {
		return "darwin"
	}
	if _, err := os.Stat("C:\\Windows"); err == nil {
		return "windows"
	}
	return "linux"
}
