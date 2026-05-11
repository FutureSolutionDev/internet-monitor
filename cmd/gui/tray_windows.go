//go:build windows

package main

import (
	"internet-monitor/monitor"
	"internet-monitor/tray"
	"runtime"
	"syscall"

	webview "github.com/webview/webview_go"

	"github.com/getlantern/systray"
)

var (
	modUser32      = syscall.NewLazyDLL("user32.dll")
	procShowWindow = modUser32.NewProc("ShowWindow")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
)

// hideWindow hides the webview window (minimizes to tray).
func hideWindow(hwnd uintptr) {
	procShowWindow.Call(hwnd, 0) // SW_HIDE
}

// restoreWindow brings the webview window back.
func restoreWindow(hwnd uintptr) {
	procShowWindow.Call(hwnd, 9) // SW_RESTORE
	procShowWindow.Call(hwnd, 5) // SW_SHOW
	procSetForegroundWindow.Call(hwnd)
}

// initTray starts the system tray in a background OS thread.
// Returns a stop function to quit the tray when the app exits.
func initTray(w webview.WebView, hwnd uintptr) func() {
	quit := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		systray.Run(func() {
			systray.SetIcon(tray.GrayIcon())
			systray.SetTooltip("Internet Monitor — جاري الفحص...")

			mOpen := systray.AddMenuItem("فتح النافذة / Open", "إظهار النافذة الرئيسية")
			systray.AddSeparator()
			mExit := systray.AddMenuItem("إغلاق / Exit", "إنهاء البرنامج")

			go func() {
				for {
					select {
					case <-mOpen.ClickedCh:
						restoreWindow(hwnd)
					case <-mExit.ClickedCh:
						systray.Quit()
						w.Dispatch(func() { w.Destroy() })
					case <-quit:
						systray.Quit()
						return
					}
				}
			}()
		}, nil)
	}()

	return func() {
		select {
		case <-quit:
		default:
			close(quit)
		}
	}
}

// updateTrayStatus changes the tray icon and tooltip to reflect connection state.
func updateTrayStatus(status monitor.Status) {
	switch status {
	case monitor.StatusConnected:
		systray.SetIcon(tray.GreenIcon())
		systray.SetTooltip("Internet Monitor — متصل ✅")
	case monitor.StatusDegraded:
		systray.SetIcon(tray.YellowIcon())
		systray.SetTooltip("Internet Monitor — ضعيف ⚠️")
	case monitor.StatusDisconnected:
		systray.SetIcon(tray.RedIcon())
		systray.SetTooltip("Internet Monitor — منقطع ❌")
	}
}
