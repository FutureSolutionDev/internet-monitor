package main

import (
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
)

func notifyText(status monitor.Status, result monitor.CheckResult) (title, body string) {
	return notifytext.Build(notifytext.Lang(), status, result)
}
