package tray

import (
	"fmt"
	"internet-monitor/config"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"os/exec"
	"time"

	"github.com/getlantern/systray"
)

type Tray struct {
	cfg     *config.Config
	checker *monitor.Checker
	lgr     *logger.Logger
	dashURL string

	// Called on every check cycle (for dashboard updates)
	OnTick  func(result monitor.CheckResult, status monitor.Status)
	// Called when status changes (for dashboard events)
	OnEvent func(event monitor.Event)

	mStatus    *systray.MenuItem
	mLastEvent *systray.MenuItem
	mOpenDash  *systray.MenuItem
	mOpenLogs  *systray.MenuItem
	mExit      *systray.MenuItem
}

func New(cfg *config.Config, checker *monitor.Checker, lgr *logger.Logger, dashURL string) *Tray {
	return &Tray{cfg: cfg, checker: checker, lgr: lgr, dashURL: dashURL}
}

func (t *Tray) OnReady() {
	systray.SetIcon(GrayIcon())
	systray.SetTooltip("Internet Monitor — Initializing...")

	t.mStatus = systray.AddMenuItem("Checking...", "Current connection status")
	t.mStatus.Disable()
	t.mLastEvent = systray.AddMenuItem("Last change: —", "Time of last status change")
	t.mLastEvent.Disable()
	systray.AddSeparator()
	t.mOpenDash = systray.AddMenuItem("Open Dashboard", "Open monitoring dashboard in browser")
	t.mOpenLogs = systray.AddMenuItem("Open Logs Folder", "Open the folder containing log files")
	systray.AddSeparator()
	t.mExit = systray.AddMenuItem("Exit", "Stop monitoring and exit")

	go t.monitorLoop()
	go t.handleMenu()
}

func (t *Tray) OnExit() {}

func (t *Tray) handleMenu() {
	for {
		select {
		case <-t.mOpenDash.ClickedCh:
			OpenURL(t.dashURL)
		case <-t.mOpenLogs.ClickedCh:
			exec.Command("explorer", t.cfg.LogDir).Start()
		case <-t.mExit.ClickedCh:
			systray.Quit()
		}
	}
}

func (t *Tray) monitorLoop() {
	var currentStatus *monitor.Status
	var statusSince time.Time
	consecutiveFails := 0

	t.runCheck(&currentStatus, &statusSince, &consecutiveFails)

	ticker := time.NewTicker(t.cfg.CheckInterval())
	defer ticker.Stop()
	for range ticker.C {
		t.runCheck(&currentStatus, &statusSince, &consecutiveFails)
	}
}

func (t *Tray) runCheck(currentStatus **monitor.Status, statusSince *time.Time, consecutiveFails *int) {
	result := t.checker.Check()
	newStatus := t.determineStatus(result, consecutiveFails)

	// Notify dashboard of every tick
	if t.OnTick != nil {
		t.OnTick(result, newStatus)
	}

	if *currentStatus == nil || **currentStatus != newStatus {
		duration := 0.0
		if !statusSince.IsZero() {
			duration = time.Since(*statusSince).Seconds()
		}
		*statusSince = time.Now()

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
		t.lgr.Log(event)

		if t.OnEvent != nil {
			t.OnEvent(event)
		}

		t.applyStatus(newStatus, result)

		if *currentStatus != nil {
			go Notify(t.notifyTitle(newStatus), t.notifyBody(result))
		}

		s := newStatus
		*currentStatus = &s
	} else {
		t.updateTooltip(newStatus, result)
	}
}

func (t *Tray) determineStatus(result monitor.CheckResult, consecutiveFails *int) monitor.Status {
	if !result.TCPPingOK || !result.HTTPOK || !result.DNSOK {
		*consecutiveFails++
	} else {
		*consecutiveFails = 0
	}
	if *consecutiveFails >= t.cfg.FailThreshold {
		return monitor.StatusDisconnected
	}
	if result.PacketLoss > t.cfg.PacketLossThreshold ||
		(result.LatencyMs > int64(t.cfg.LatencyThreshold) && result.LatencyMs > 0) {
		return monitor.StatusDegraded
	}
	return monitor.StatusConnected
}

func (t *Tray) applyStatus(status monitor.Status, result monitor.CheckResult) {
	switch status {
	case monitor.StatusConnected:
		systray.SetIcon(GreenIcon())
		t.mStatus.SetTitle("Connected")
	case monitor.StatusDegraded:
		systray.SetIcon(YellowIcon())
		t.mStatus.SetTitle(fmt.Sprintf("Degraded  |  Loss: %.0f%%  |  %dms", result.PacketLoss, result.LatencyMs))
	case monitor.StatusDisconnected:
		systray.SetIcon(RedIcon())
		t.mStatus.SetTitle("Disconnected")
	}
	t.mLastEvent.SetTitle(fmt.Sprintf("Last change: %s", time.Now().Format("15:04:05")))
	t.updateTooltip(status, result)
}

func (t *Tray) updateTooltip(status monitor.Status, result monitor.CheckResult) {
	systray.SetTooltip(fmt.Sprintf("Internet Monitor — %s | Loss: %.0f%% | %dms",
		status, result.PacketLoss, result.LatencyMs))
}

func (t *Tray) notifyTitle(status monitor.Status) string {
	switch status {
	case monitor.StatusConnected:
		return "Internet Restored"
	case monitor.StatusDegraded:
		return "Connection Degraded"
	default:
		return "Internet Disconnected"
	}
}

func (t *Tray) notifyBody(result monitor.CheckResult) string {
	parts := []string{}
	if !result.TCPPingOK {
		parts = append(parts, "TCP ping failed")
	}
	if !result.HTTPOK {
		parts = append(parts, "HTTP failed")
	}
	if !result.DNSOK {
		parts = append(parts, "DNS failed")
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Latency: %dms  |  Loss: %.0f%%", result.LatencyMs, result.PacketLoss)
	}
	msg := ""
	for i, p := range parts {
		if i > 0 {
			msg += ", "
		}
		msg += p
	}
	return msg
}
