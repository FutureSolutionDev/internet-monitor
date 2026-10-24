//go:build windows

package tray

import (
	"internet-monitor/internal/alerts"
	"internet-monitor/internal/i18n"
	"internet-monitor/internal/types"
	"log"
	"os"

	"golang.org/x/sys/windows"
)

func notifyLogf(format string, args ...interface{}) {
	if Logf != nil {
		Logf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func statusIcon(s types.Status) string {
	switch s {
	case types.StatusConnected:
		return statusIconPath("ok", 34, 197, 94)
	case types.StatusDegraded:
		return statusIconPath("warn", 234, 179, 8)
	default:
		return statusIconPath("down", 239, 68, 68)
	}
}

func showDesktop(t alerts.Toast, action string) {
	tag, group := toastTag(t.Kind)
	c := ToastContent{
		Title:       t.Title,
		Body:        t.Body,
		Attribution: i18n.T(t.Lang, "app_name", nil) + " · " + t.Time.Local().Format("15:04"),
		IconPath:    statusIcon(t.Status),
		ActionLabel: i18n.T(t.Lang, "notify.action.open", nil),
		ActionURL:   action,
	}
	if err := showToast(c, tag, group); err != nil {
		notifyLogf("[notify] native toast failed: %v", err)
		return
	}
	notifyLogf("[notify] native toast shown: kind=%s", t.Kind)
}

func shellOpen(target string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, file, nil, nil, windows.SW_SHOWNORMAL)
}

func OpenURL(url string) {
	valid, err := ValidateWebURL(url)
	if err != nil {
		notifyLogf("[notify] OpenURL rejected: %v", err)
		return
	}
	if err := shellOpen(valid); err != nil {
		notifyLogf("[notify] OpenURL failed: %v", err)
	}
}

func OpenFolder(path string) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		notifyLogf("[notify] OpenFolder rejected %q: not an existing directory", path)
		return
	}
	if err := shellOpen(path); err != nil {
		notifyLogf("[notify] OpenFolder failed: %v", err)
	}
}
