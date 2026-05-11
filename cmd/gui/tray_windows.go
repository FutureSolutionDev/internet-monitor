//go:build windows

package main

import (
	"internet-monitor/monitor"
	"internet-monitor/tray"
	"runtime"
	"syscall"
	"time"

	webview "github.com/webview/webview_go"

	"github.com/getlantern/systray"
)

var (
	modUser32               = syscall.NewLazyDLL("user32.dll")
	procShowWindow          = modUser32.NewProc("ShowWindow")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
)

func hideWindow(hwnd uintptr) { procShowWindow.Call(hwnd, 0) }

func restoreWindow(hwnd uintptr) {
	procShowWindow.Call(hwnd, 9)
	procShowWindow.Call(hwnd, 5)
	procSetForegroundWindow.Call(hwnd)
}

// initTray starts systray in a locked OS thread.
// The returned stop func blocks until systray has fully exited — prevents zombie processes.
func initTray(w webview.WebView, hwnd uintptr) func() {
	quit := make(chan struct{})
	done := make(chan struct{}) // closed when systray.Run() returns

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
						w.Terminate()
					case <-quit:
						systray.Quit()
						return
					}
				}
			}()
		}, nil)
		close(done) // signal that the message loop has fully exited
	}()

	return func() {
		select {
		case <-quit:
		default:
			close(quit)
		}
		// Wait up to 2s for systray to exit cleanly.
		// On Windows, systray.Run() occasionally doesn't return after Quit()
		// which would block os.Exit forever — the timeout breaks that deadlock.
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	}
}

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
