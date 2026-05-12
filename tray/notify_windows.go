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

func init() {
	if k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Classes\AppUserModelId\`+notifyAUMID,
		registry.SET_VALUE,
	); err == nil {
		k.SetStringValue("DisplayName", "Internet Monitor")
		k.Close()
	}
	go func() {
		log.Println("[notify] ensuring Start Menu shortcut…")
		EnsureStartMenuShortcut(notifyAUMID)
	}()
}

// Notify shows a system notification with sound (tray binary).
func Notify(title, message string) {
	go playTraySound()
	lnk := filepath.Join(os.Getenv("APPDATA"),
		"Microsoft", "Windows", "Start Menu", "Programs",
		"Internet Monitor.lnk")
	if _, err := os.Stat(lnk); err == nil {
		showWinRTToast(title, message)
	} else {
		ShowBalloon(title, message)
	}
}

// ShowWinRTToast is exported so cmd/gui can call it directly.
func ShowWinRTToast(title, body string) {
	showWinRTToast(title, body)
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
$tpl   = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$nodes = $tpl.GetElementsByTagName('text')
$nodes.Item(0).AppendChild($tpl.CreateTextNode('%s')) | Out-Null
$nodes.Item(1).AppendChild($tpl.CreateTextNode('%s')) | Out-Null
$toast = [Windows.UI.Notifications.ToastNotification]::new($tpl)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('%s').Show($toast)
`, t, b, notifyAUMID)

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
