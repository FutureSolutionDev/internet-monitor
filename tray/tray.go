package tray

import (
	"fmt"
	"internet-monitor/monitor"
	"time"

	"github.com/getlantern/systray"
)

type Tray struct {
	dashURL string
	logDir  string

	mStatus    *systray.MenuItem
	mLastEvent *systray.MenuItem
	mOpenDash  *systray.MenuItem
	mOpenLogs  *systray.MenuItem
	mExit      *systray.MenuItem
}

func New(logDir, dashURL string) *Tray {
	return &Tray{logDir: logDir, dashURL: dashURL}
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

	go t.handleMenu()
}

func (t *Tray) OnExit() {}

func (t *Tray) handleMenu() {
	for {
		select {
		case <-t.mOpenDash.ClickedCh:
			OpenURL(t.dashURL)
		case <-t.mOpenLogs.ClickedCh:
			OpenFolder(t.logDir)
		case <-t.mExit.ClickedCh:
			systray.Quit()
		}
	}
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
