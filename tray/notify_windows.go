//go:build windows

package tray

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const notifyAUMID = "InternetMonitor"

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
func Notify(title, message string) {
	playTraySound()
	notifyLogf("[notify] Notify(tray): title=%q", title)
	ShowNotification(title, message)
}

// startMenuLnkPath is the Start Menu shortcut whose AppUserModelID associates
// our notifications with the app's name (required for WinRT toasts on Win10/11).
func startMenuLnkPath() string {
	return filepath.Join(os.Getenv("APPDATA"),
		"Microsoft", "Windows", "Start Menu", "Programs",
		"Internet Monitor.lnk")
}

// ShowNotification shows a desktop notification. It prefers a WinRT toast —
// which renders under the app's name/icon — when the Start Menu shortcut (with
// our AppUserModelID) exists. Otherwise, or if the toast can't be shown, it
// falls back to the PowerShell WinForms balloon (shows under "PowerShell" but
// always renders, e.g. under `air` where the exe is a throwaway in tmp).
func ShowNotification(title, message string) {
	if _, err := os.Stat(startMenuLnkPath()); err == nil {
		if showWinRTToast(title, message) {
			return
		}
		notifyLogf("[notify] toast failed, falling back to balloon")
	} else {
		notifyLogf("[notify] no Start Menu shortcut yet, using balloon")
	}
	ShowBalloon(title, message)
}

// showWinRTToast fires a Windows 10/11 toast under our AUMID via PowerShell.
// Returns true only if PowerShell reported the toast was shown.
func showWinRTToast(title, body string) bool {
	t := strings.ReplaceAll(title, "'", "''")
	b := strings.ReplaceAll(body, "'", "''")
	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null
$tpl=[Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$n=$tpl.GetElementsByTagName('text')
$n.Item(0).AppendChild($tpl.CreateTextNode('%s'))|Out-Null
$n.Item(1).AppendChild($tpl.CreateTextNode('%s'))|Out-Null
$toast=[Windows.UI.Notifications.ToastNotification]::new($tpl)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('%s').Show($toast)
Write-Host 'TOAST_SHOWN'`, t, b, notifyAUMID)

	// NOTE: do NOT set DETACHED_PROCESS here — it detaches stdout so
	// CombinedOutput() returns "" even though the toast was shown, making us
	// wrongly think it failed and fall back to the (PowerShell-named) balloon.
	// HideWindow alone keeps the console hidden.
	cmd := exec.Command("powershell",
		"-NoProfile", "-ExecutionPolicy", "Bypass",
		"-WindowStyle", "Hidden", "-NonInteractive",
		"-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "TOAST_SHOWN") {
		notifyLogf("[notify] showWinRTToast: err=%v out=%q", err, strings.TrimSpace(string(out)))
		return false
	}
	return true
}

func OpenURL(url string) {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func OpenFolder(path string) {
	exec.Command("explorer", path).Start()
}
