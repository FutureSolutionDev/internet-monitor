//go:build !windows

package main

import (
	"internet-monitor/monitor"

	webview "github.com/webview/webview_go"
)

// restoreWindow / setWindowIconFromExe have no non-Windows equivalent and are
// only called from tray_windows.go, so they are intentionally omitted here.
func hideWindow(_ uintptr)    {}
func setWindowIcon(_ uintptr) {}

func initTray(_ webview.WebView, _ uintptr) func() { return func() {} }
func updateTrayStatus(_ monitor.Status)            {}
