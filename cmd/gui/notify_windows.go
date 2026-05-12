//go:build windows

package main

import (
	"internet-monitor/monitor"
	"internet-monitor/tray"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const toastAppID = "InternetMonitor"

func init() {
	registerToastApp()
	// Ensure Start Menu shortcut exists so Windows 10/11 shows banner popups.
	go func() {
		log.Println("[notify] ensuring Start Menu shortcut…")
		tray.EnsureStartMenuShortcut(toastAppID)
	}()
}

func registerToastApp() {
	k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Classes\AppUserModelId\`+toastAppID,
		registry.SET_VALUE,
	)
	if err != nil {
		return
	}
	defer k.Close()
	k.SetStringValue("DisplayName", "Internet Monitor")
}

// ── Audio (MCI) ────────────────────────────────────────────────

var (
	modWinmm       = syscall.NewLazyDLL("winmm.dll")
	procMciSendStr = modWinmm.NewProc("mciSendStringW")

	ringMu      sync.Mutex
	ringPlaying bool
)

func mciCall(cmd string) {
	p, _ := syscall.UTF16PtrFromString(cmd)
	procMciSendStr.Call(uintptr(unsafe.Pointer(p)), 0, 0, 0)
}

func playRingtone() {
	path := getRingtonePath()
	if path == "" {
		return
	}
	go func() {
		ringMu.Lock()
		if ringPlaying {
			mciCall("stop im_ring")
			mciCall("close im_ring")
		}
		ringPlaying = true
		ringMu.Unlock()

		mciCall(`open "` + path + `" type mpegvideo alias im_ring`)
		mciCall("play im_ring")
		time.Sleep(15 * time.Second)
		mciCall("stop im_ring")
		mciCall("close im_ring")

		ringMu.Lock()
		ringPlaying = false
		ringMu.Unlock()
	}()
}

// ── Notifications ──────────────────────────────────────────────

var (
	notifyMu       sync.Mutex
	lastNotifyTime time.Time
)

const guiNotifyCooldown = 4 * time.Second

func showSystemNotification(title, body string) {
	// Prefer WinRT toast (banner popup) if Start Menu shortcut already exists.
	// Fall back to Shell_NotifyIcon balloon otherwise.
	lnk := startMenuLnkPath()
	if _, err := os.Stat(lnk); err == nil {
		tray.ShowWinRTToast(title, body)
	} else {
		tray.ShowBalloon(title, body)
	}
}

func startMenuLnkPath() string {
	return filepath.Join(os.Getenv("APPDATA"),
		"Microsoft", "Windows", "Start Menu", "Programs",
		"Internet Monitor.lnk")
}

func sendNotification(status monitor.Status, result monitor.CheckResult) {
	title, body := notifyText(status, result)
	if title == "" {
		return
	}
	notifyMu.Lock()
	if time.Since(lastNotifyTime) < guiNotifyCooldown {
		notifyMu.Unlock()
		return
	}
	lastNotifyTime = time.Now()
	notifyMu.Unlock()

	playRingtone()
	showSystemNotification(title, body)
}

func TestNotification() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[notify] TestNotification panic recovered: %v", r)
		}
	}()
	playRingtone()
	showSystemNotification("🔔 اختبار الإشعار / Test", "الصوت والإشعار يعملان ✅")
}
