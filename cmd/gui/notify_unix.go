//go:build darwin || linux

package main

import (
	"fmt"
	"internet-monitor/monitor"
	"internet-monitor/notifytext"
	"os/exec"
	"runtime"
)

func playRingtone() {
	path := getRingtonePath()
	if path == "" {
		return
	}
	switch runtime.GOOS {
	case "darwin":
		exec.Command("afplay", path).Start()
	default: // linux — try common players
		if exec.Command("mpg123", "-q", path).Start() == nil {
			return
		}
		exec.Command("ffplay", "-nodisp", "-autoexit", path).Start()
	}
}

func sendNotification(status monitor.Status, result monitor.CheckResult) {
	title, body := notifyText(status, result)
	if title == "" {
		return
	}
	playRingtone()
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification "%s" with title "%s"`,
			notifytext.EscapeAppleScript(body), notifytext.EscapeAppleScript(title))
		exec.Command("osascript", "-e", script).Start()
	default:
		exec.Command("notify-send", "-a", "Internet Monitor", title, body).Start()
	}
}

// TestNotification plays the ringtone and shows a sample notification.
func TestNotification() {
	playRingtone()
	switch runtime.GOOS {
	case "darwin":
		exec.Command("osascript", "-e",
			`display notification "الصوت والإشعار يعملان" with title "اختبار الإشعار"`).Start()
	default:
		exec.Command("notify-send", "-a", "Internet Monitor",
			"Test Notification", "🔔 Sound and notification are working").Start()
	}
}
