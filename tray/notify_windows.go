//go:build windows

package tray

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const notifyAUMID = "InternetMonitor"

// Logf, if set, receives notification-path diagnostics (wired to logger.AppLog
// by main so they land in logs/app.log; the standard log package is invisible
// in a -H=windowsgui build).
var Logf func(format string, args ...interface{})

func notifyLogf(format string, args ...interface{}) {
	if Logf != nil {
		Logf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func init() {
	// 1. Register AUMID display name + icon (used by WinRT toast notifications).
	if k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Classes\AppUserModelId\`+notifyAUMID,
		registry.SET_VALUE,
	); err == nil {
		k.SetStringValue("DisplayName", "Internet Monitor")
		if exe, err2 := os.Executable(); err2 == nil {
			k.SetStringValue("IconUri", exe)
		}
		k.Close()
	}

	// 2. Enable banner notifications. Only set Enabled=1 on first run.
	const notifPath = `SOFTWARE\Microsoft\Windows\CurrentVersion\Notifications\Settings\` + notifyAUMID
	if k, created, err := registry.CreateKey(
		registry.CURRENT_USER, notifPath, registry.SET_VALUE,
	); err == nil {
		if !created {
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

	go func() {
		log.Println("[notify] ensuring Start Menu shortcut…")
		EnsureStartMenuShortcut(notifyAUMID)
	}()
}

// Notify shows a system notification with sound (tray binary).
//
// Routes straight to ShowBalloon, which renders via a PowerShell WinForms
// balloon — reliable from an unpackaged exe. (The old WinRT-toast path failed
// silently with no box, so it was removed.) Sound plays inline; sound.Play
// stops any prior sound first.
func Notify(title, message string) {
	playTraySound()
	notifyLogf("[notify] Notify(tray): title=%q", title)
	ShowBalloon(title, message)
}

func OpenURL(url string) {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func OpenFolder(path string) {
	exec.Command("explorer", path).Start()
}
