//go:build darwin || linux

package main

import (
	"fmt"
	"internet-monitor/monitor"
	"internet-monitor/notifytext"
	"internet-monitor/sound"
	"os/exec"
	"runtime"
)

// playRingtone plays the ringtone via the shared player, which stops any
// previous sound first so rapid notifications never overlap.
func playRingtone() { sound.Play() }

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

// TestNotification plays the ringtone and shows a sample notification in the
// given UI language.
func TestNotification(lang string) {
	playRingtone()
	title, body := notifytext.TestMessage(lang)
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification "%s" with title "%s"`,
			notifytext.EscapeAppleScript(body), notifytext.EscapeAppleScript(title))
		exec.Command("osascript", "-e", script).Start()
	default:
		exec.Command("notify-send", "-a", "Internet Monitor", title, body).Start()
	}
}
