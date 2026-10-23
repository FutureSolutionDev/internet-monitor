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

func Notify(title, message string) {
	playTraySound()
	notifyLogf("[notify] Notify(tray): title=%q", title)
	ShowNotification(title, message)
}

func startMenuLnkPath() string {
	return filepath.Join(os.Getenv("APPDATA"),
		"Microsoft", "Windows", "Start Menu", "Programs",
		"Internet Monitor.lnk")
}

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
