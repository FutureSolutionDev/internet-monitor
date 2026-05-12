//go:build windows

package main

import (
	"internet-monitor/monitor"
	"internet-monitor/tray"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	webview "github.com/webview/webview_go"

	"github.com/getlantern/systray"
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

// setWindowIcon loads the favicon ICO from a temp file and applies it to
// the WebView2 window via WM_SETICON (title bar, Alt+Tab, taskbar button).
// This is needed because go-winres is not in the dev environment, so the
// exe has no embedded icon resource.
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

// setWindowIconFromExe loads the icon from the running exe's embedded resources.
// Used after go-winres is run and a release build is produced.
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

// initTray starts systray in a locked OS thread.
// The returned stop func blocks until systray has fully exited — prevents zombie processes.
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
						// Dispatch Terminate to the main thread — PostQuitMessage(0)
						// must run on the same thread as w.Run() or WM_QUIT goes
						// to the wrong queue and the process becomes a zombie.
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
