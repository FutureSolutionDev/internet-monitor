//go:build darwin || linux

package main

import (
	"fmt"
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/sound"
	"os/exec"
	"runtime"
)

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
