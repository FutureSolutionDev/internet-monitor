//go:build windows

package main

import (
	"internet-monitor/monitor"
	"internet-monitor/notifytext"
	"internet-monitor/sound"
	"internet-monitor/tray"
	"log"
	"os"
	"sync"
	"time"

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
	// 1. Register AUMID display name + icon (used by WinRT toast notifications).
	if k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Classes\AppUserModelId\`+toastAppID,
		registry.SET_VALUE,
	); err == nil {
		k.SetStringValue("DisplayName", "Internet Monitor")
		if exe, err2 := os.Executable(); err2 == nil {
			k.SetStringValue("IconUri", exe)
		}
		k.Close()
	}

	// 2. Enable banner notifications in Windows notification settings.
	//    Without Enabled=1 Windows 10 silently drops toasts for unknown AUMIDs.
	//    We only set this on first run (key doesn't exist yet) to avoid
	//    overriding a user's explicit preference.
	const notifPath = `SOFTWARE\Microsoft\Windows\CurrentVersion\Notifications\Settings\` + toastAppID
	if k, created, err := registry.CreateKey(
		registry.CURRENT_USER, notifPath, registry.SET_VALUE,
	); err == nil {
		if !created { // key already existed — check if Enabled is set
			if _, _, verr := k.GetIntegerValue("Enabled"); verr != nil {
				k.SetDWordValue("Enabled", 1)
				k.SetDWordValue("ShowInActionCenter", 1)
			}
		} else {
			k.SetDWordValue("Enabled", 1)
			k.SetDWordValue("ShowInActionCenter", 1)
		}
		k.Close()
	}
}

// ── Audio ──────────────────────────────────────────────────────

// playRingtone plays the ringtone via the shared player, which stops any
// previous sound first (non-blocking) so rapid notifications never overlap.
func playRingtone() { sound.Play() }

// ── Notifications ──────────────────────────────────────────────

var (
	notifyMu       sync.Mutex
	lastNotifyTime time.Time
)

const guiNotifyCooldown = 4 * time.Second

func showSystemNotification(title, body string) {
	// Prefer the app-named WinRT toast (falls back to the WinForms balloon when
	// the Start Menu shortcut isn't present, e.g. under `air`).
	tray.ShowNotification(title, body)
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

func TestNotification(lang string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[notify] TestNotification panic recovered: %v", r)
		}
	}()
	playRingtone()
	title, body := notifytext.TestMessage(lang)
	showSystemNotification(title, body)
}
