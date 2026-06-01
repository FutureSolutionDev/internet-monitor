package main

import (
	"internet-monitor/monitor"
	"internet-monitor/notifytext"
)

// notifyText returns the notification title/body for a status + result.
// Logic lives in the shared notifytext package (also used by the tray build).
func notifyText(status monitor.Status, result monitor.CheckResult) (title, body string) {
	return notifytext.Build(notifytext.Lang(), status, result)
}
