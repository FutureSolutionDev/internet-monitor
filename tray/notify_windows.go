//go:build windows

package tray

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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
// Routes straight to ShowBalloon: the balloon uses Shell_NotifyIcon on the
// existing tray icon and falls back to a PowerShell WinForms balloon, which
// reliably shows a box from an unpackaged exe. The previous WinRT-toast path
// failed silently (no box) on unpackaged Go binaries, so it's no longer the
// default. Sound is played inline (sound.Play stops any prior sound first).
func Notify(title, message string) {
	playTraySound()
	notifyLogf("[notify] Notify(tray): title=%q", title)
	ShowBalloon(title, message)
}

// ShowWinRTToast is kept for compatibility but now routes to the balloon,
// because the WinRT toast fails silently from an unpackaged exe.
func ShowWinRTToast(title, body string) {
	notifyLogf("[notify] ShowWinRTToast -> balloon (toast unreliable unpackaged): title=%q", title)
	ShowBalloon(title, body)
}

// showWinRTToast fires a Windows 10/11 WinRT toast via PowerShell.
// Uses GetTemplateContent to avoid having to load Windows.Data.Xml.Dom
// separately. AppendChild+CreateTextNode is more reliable than InnerText.
// DETACHED_PROCESS prevents the notification from stealing console focus.
func showWinRTToast(title, body string) {
	t := strings.ReplaceAll(title, "'", "''")
	b := strings.ReplaceAll(body, "'", "''")

	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('%s')
Write-Host "notifier.Setting=$($notifier.Setting)"
if ($notifier.Setting -ne 0) { Write-Host "BLOCKED setting=$($notifier.Setting)"; exit 0 }
$tpl   = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$nodes = $tpl.GetElementsByTagName('text')
$nodes.Item(0).AppendChild($tpl.CreateTextNode('%s')) | Out-Null
$nodes.Item(1).AppendChild($tpl.CreateTextNode('%s')) | Out-Null
$toast = [Windows.UI.Notifications.ToastNotification]::new($tpl)
$notifier.Show($toast)
Write-Host "SHOWN"
`, notifyAUMID, t, b)

	cmd := exec.Command("powershell",
		"-WindowStyle", "Hidden",
		"-NonInteractive",
		"-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x00000008, // DETACHED_PROCESS — no console focus steal
	}
	go func() {
		out, err := cmd.CombinedOutput()
		if err != nil || len(out) > 0 {
			log.Printf("[notify] showWinRTToast: err=%v output=%s", err, out)
		} else {
			log.Println("[notify] showWinRTToast: OK")
		}
	}()
}

func OpenURL(url string) {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func OpenFolder(path string) {
	exec.Command("explorer", path).Start()
}
