//go:build darwin || linux

package main

import (
	"fmt"
	"internet-monitor/monitor"
	"os/exec"
	"runtime"
)

func sendNotification(status monitor.Status, result monitor.CheckResult) {
	title, body := notifyText(status, result)
	if title == "" {
		return
	}
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
		exec.Command("osascript", "-e", script).Start()
	default: // linux
		exec.Command("notify-send", "-a", "Internet Monitor", title, body).Start()
	}
}
