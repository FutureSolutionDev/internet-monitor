//go:build windows

package tray

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32dll     = windows.NewLazySystemDLL("user32.dll")
	shell32dll    = windows.NewLazySystemDLL("shell32.dll")
	shellNotify   = shell32dll.NewProc("Shell_NotifyIconW")
	enumWindowsW  = user32dll.NewProc("EnumWindows")
	getWinThread  = user32dll.NewProc("GetWindowThreadProcessId")
	getClassNameW = user32dll.NewProc("GetClassNameW")
)

// balloonNID mirrors NOTIFYICONDATA (Vista+ full layout, 976 bytes on 64-bit).
// Go's automatic field alignment produces the same padding as the C struct.
type balloonNID struct {
	Size            uint32
	Wnd             windows.Handle // 8-byte aligned → 4 bytes auto-padding before
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            windows.Handle // 8-byte aligned → 4 bytes auto-padding before
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Timeout         uint32 // union with Version; deprecated Vista+
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        windows.GUID
	BalloonIcon     windows.Handle
}

const (
	nimModify   = 0x00000001
	nifInfo     = 0x00000020
	niifNoSound = 0x00000010 // suppress default system sound (we play our own)
)

var ourPID = uint32(os.Getpid())

// Package-level result + mutex so the EnumWindows callback never writes to
// a Go stack pointer (which could become stale if the goroutine stack grows).
var (
	enumResult   windows.Handle
	enumResultMu sync.Mutex
)

// enumCB is cached once — syscall.NewCallback allocates permanent memory.
// The closure only reads package-level vars and writes to enumResult (heap).
var enumCB = syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
	pid := new(uint32)
	getWinThread.Call(hwnd, uintptr(unsafe.Pointer(pid)))
	if *pid != ourPID {
		return 1 // continue — different process
	}
	className := new([256]uint16)
	getClassNameW.Call(hwnd, uintptr(unsafe.Pointer(className)), 256)
	if windows.UTF16ToString(className[:]) == "SystrayClass" {
		enumResult = windows.Handle(hwnd)
		return 0 // stop enumeration
	}
	return 1 // continue
})

// findSystrayHWND returns the HWND of the SystrayClass window that belongs
// to THIS process (registered by getlantern/systray with icon ID=100).
func findSystrayHWND() windows.Handle {
	enumResultMu.Lock()
	defer enumResultMu.Unlock()
	enumResult = 0
	enumWindowsW.Call(enumCB, 0)
	return enumResult
}

// tryNativeBalloon calls Shell_NotifyIcon(NIM_MODIFY, NIF_INFO) on the
// existing systray icon. Returns true when Shell_NotifyIcon succeeds.
func tryNativeBalloon(title, message string) bool {
	hwnd := findSystrayHWND()
	if hwnd == 0 {
		notifyLogf("[notify] tryNativeBalloon: SystrayClass window not found — systray not yet ready?")
		return false
	}
	notifyLogf("[notify] tryNativeBalloon: found SystrayClass HWND=0x%X", hwnd)

	nid := &balloonNID{
		Wnd:       hwnd,
		ID:        100, // getlantern/systray always registers with ID=100
		Flags:     nifInfo,
		InfoFlags: niifNoSound,
	}
	nid.Size = uint32(unsafe.Sizeof(*nid))

	if title != "" {
		t := windows.StringToUTF16(title)
		n := copy(nid.InfoTitle[:len(nid.InfoTitle)-1], t)
		nid.InfoTitle[n] = 0
	}
	if message != "" {
		m := windows.StringToUTF16(message)
		n := copy(nid.Info[:len(nid.Info)-1], m)
		nid.Info[n] = 0
	}

	ret, _, lastErr := shellNotify.Call(nimModify, uintptr(unsafe.Pointer(nid)))
	if ret == 0 {
		notifyLogf("[notify] tryNativeBalloon: Shell_NotifyIcon failed — err=%v", lastErr)
		return false
	}
	notifyLogf("[notify] tryNativeBalloon: Shell_NotifyIcon succeeded")
	return true
}

// showBalloonWinForms uses PowerShell + System.Windows.Forms as a reliable
// fallback that works without AUMID registration or a Start Menu shortcut.
func showBalloonWinForms(title, message string) {
	t := strings.ReplaceAll(title, "'", "''")
	m := strings.ReplaceAll(message, "'", "''")

	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$n = [System.Windows.Forms.NotifyIcon]::new()
$n.Icon = [System.Drawing.SystemIcons]::Application
$n.BalloonTipIcon = [System.Windows.Forms.ToolTipIcon]::Info
$n.BalloonTipTitle = '%s'
$n.BalloonTipText  = '%s'
$n.Visible = $true
$n.ShowBalloonTip(5000)
Start-Sleep -Seconds 6
$n.Visible = $false
$n.Dispose()`, t, m)

	cmd := exec.Command("powershell",
		"-WindowStyle", "Hidden",
		"-NonInteractive",
		"-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

// ShowBalloon shows a notification balloon.
// Tries native Shell_NotifyIcon first; falls back to PowerShell WinForms.
func ShowBalloon(title, message string) {
	notifyLogf("[notify] ShowBalloon: title=%q", title)
	if !tryNativeBalloon(title, message) {
		notifyLogf("[notify] ShowBalloon: native failed, falling back to PowerShell WinForms")
		showBalloonWinForms(title, message)
	}
}
