//go:build windows

package main

import (
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/tray"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	webview "github.com/webview/webview_go"

	"fyne.io/systray"
)

var (
	modUser32               = syscall.NewLazyDLL("user32.dll")
	procShowWindow          = modUser32.NewProc("ShowWindow")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
	procSendMessageW        = modUser32.NewProc("SendMessageW")
	procLoadImageW          = modUser32.NewProc("LoadImageW")
)

const (
	wmSetIcon      = 0x0080
	iconSmall      = 0
	iconBig        = 1
	imageIcon      = 1
	lrLoadFromFile = 0x0010
	lrDefaultSize  = 0x0040
)

func hideWindow(hwnd uintptr) { procShowWindow.Call(hwnd, 0) }

func restoreWindow(hwnd uintptr) {
	procShowWindow.Call(hwnd, 9)
	procShowWindow.Call(hwnd, 5)
	procSetForegroundWindow.Call(hwnd)
}

func setWindowIcon(hwnd uintptr) {
	icoData := tray.GrayIcon()
	if len(icoData) == 0 {
		return
	}
	tmp, err := os.CreateTemp("", "im-icon*.ico")
	if err != nil {
		return
	}
	tmp.Write(icoData)
	tmp.Close()
	defer os.Remove(tmp.Name())

	path, err := syscall.UTF16PtrFromString(tmp.Name())
	if err != nil {
		return
	}
	hIcon, _, _ := procLoadImageW.Call(
		0,
		uintptr(unsafe.Pointer(path)),
		imageIcon,
		0, 0,
		lrLoadFromFile|lrDefaultSize,
	)
	if hIcon == 0 {
		return
	}
	procSendMessageW.Call(hwnd, wmSetIcon, iconBig, hIcon)
	procSendMessageW.Call(hwnd, wmSetIcon, iconSmall, hIcon)
}

func setWindowIconFromExe(hwnd uintptr) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exe, _ = filepath.EvalSymlinks(exe)
	path, err := syscall.UTF16PtrFromString(exe)
	if err != nil {
		return
	}
	hIcon, _, _ := procLoadImageW.Call(
		0,
		uintptr(unsafe.Pointer(path)),
		imageIcon,
		0, 0,
		lrLoadFromFile|lrDefaultSize,
	)
	if hIcon == 0 {
		return
	}
	procSendMessageW.Call(hwnd, wmSetIcon, iconBig, hIcon)
	procSendMessageW.Call(hwnd, wmSetIcon, iconSmall, hIcon)
}

func initTray(w webview.WebView, hwnd uintptr) func() {
	quit := make(chan struct{})
	done := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		systray.Run(func() {
			systray.SetIcon(tray.GrayIcon())
			systray.SetTooltip("Internet Monitor — جاري الفحص...")

			mOpen := systray.AddMenuItem("فتح النافذة / Open", "إظهار النافذة الرئيسية")
			systray.AddSeparator()
			mExit := systray.AddMenuItem("إغلاق / Exit", "إنهاية البرنامج")

			go func() {
				for {
					select {
					case <-mOpen.ClickedCh:
						restoreWindow(hwnd)
					case <-mExit.ClickedCh:
						systray.Quit()

						w.Dispatch(func() { w.Terminate() })
					case <-quit:
						systray.Quit()
						return
					}
				}
			}()
		}, nil)
		close(done)
	}()

	return func() {
		select {
		case <-quit:
		default:
			close(quit)
		}
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
