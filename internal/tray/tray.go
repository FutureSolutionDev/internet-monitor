package tray

import (
	"internet-monitor/internal/i18n"
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
	"sync"
	"time"

	"fyne.io/systray"
)

type Tray struct {
	dashURL string
	logDir  string

	mStatus    *systray.MenuItem
	mLastEvent *systray.MenuItem
	mOpenDash  *systray.MenuItem
	mOpenLogs  *systray.MenuItem
	mExit      *systray.MenuItem

	mu         sync.Mutex
	ready      bool
	known      bool
	status     monitor.Status
	result     monitor.CheckResult
	lastChange time.Time
}

func New(logDir, dashURL string) *Tray {
	return &Tray{logDir: logDir, dashURL: dashURL}
}

func (t *Tray) OnReady() {
	lang := notifytext.Lang()
	systray.SetIcon(GrayIcon())
	systray.SetTooltip(i18n.T(lang, "app_name", nil) + " — " + i18n.T(lang, "tray.starting", nil))

	t.mStatus = systray.AddMenuItem(i18n.T(lang, "tray.status.checking", nil), "")
	t.mStatus.Disable()
	t.mLastEvent = systray.AddMenuItem(i18n.T(lang, "tray.last_change_none", nil), "")
	t.mLastEvent.Disable()
	systray.AddSeparator()
	t.mOpenDash = systray.AddMenuItem(i18n.T(lang, "tray.open_dashboard", nil), i18n.T(lang, "tray.open_dashboard_hint", nil))
	t.mOpenLogs = systray.AddMenuItem(i18n.T(lang, "tray.open_logs", nil), i18n.T(lang, "tray.open_logs_hint", nil))
	systray.AddSeparator()
	t.mExit = systray.AddMenuItem(i18n.T(lang, "tray.exit", nil), i18n.T(lang, "tray.exit_hint", nil))

	t.mu.Lock()
	t.ready = true
	t.mu.Unlock()
	notifytext.OnLangChange(func(string) { t.relabel() })

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
	case monitor.StatusDegraded:
		systray.SetIcon(YellowIcon())
	case monitor.StatusDisconnected:
		systray.SetIcon(RedIcon())
	}
	t.mu.Lock()
	t.known, t.status, t.result, t.lastChange = true, status, result, time.Now()
	t.mu.Unlock()
	t.relabel()
}

func (t *Tray) updateTooltip(status monitor.Status, result monitor.CheckResult) {
	t.mu.Lock()
	t.known, t.status, t.result = true, status, result
	ready := t.ready
	t.mu.Unlock()
	if ready {
		lang := notifytext.Lang()
		systray.SetTooltip(i18n.T(lang, "app_name", nil) + " — " + notifytext.StatusDetail(lang, status, result))
	}
}

func (t *Tray) relabel() {
	t.mu.Lock()
	ready, known, status, result, last := t.ready, t.known, t.status, t.result, t.lastChange
	t.mu.Unlock()
	if !ready {
		return
	}
	lang := notifytext.Lang()
	t.mOpenDash.SetTitle(i18n.T(lang, "tray.open_dashboard", nil))
	t.mOpenDash.SetTooltip(i18n.T(lang, "tray.open_dashboard_hint", nil))
	t.mOpenLogs.SetTitle(i18n.T(lang, "tray.open_logs", nil))
	t.mOpenLogs.SetTooltip(i18n.T(lang, "tray.open_logs_hint", nil))
	t.mExit.SetTitle(i18n.T(lang, "tray.exit", nil))
	t.mExit.SetTooltip(i18n.T(lang, "tray.exit_hint", nil))
	if !known {
		t.mStatus.SetTitle(i18n.T(lang, "tray.status.checking", nil))
		return
	}
	t.mStatus.SetTitle(notifytext.StatusDetail(lang, status, result))
	if last.IsZero() {
		t.mLastEvent.SetTitle(i18n.T(lang, "tray.last_change_none", nil))
	} else {
		t.mLastEvent.SetTitle(i18n.T(lang, "tray.last_change", i18n.Vars{"time": last.Format("15:04:05")}))
	}
	systray.SetTooltip(i18n.T(lang, "app_name", nil) + " — " + notifytext.StatusDetail(lang, status, result))
}
