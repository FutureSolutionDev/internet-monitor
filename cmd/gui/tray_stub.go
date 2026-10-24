//go:build !windows

package main

import (
	"internet-monitor/internal/monitor"

	webview "github.com/webview/webview_go"
)

func hideWindow(_ uintptr)    {}
func setWindowIcon(_ uintptr) {}

func initTray(_ webview.WebView, _ uintptr) func() { return func() {} }
func updateTrayStatus(_ monitor.Status)            {}
