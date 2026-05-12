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
	// 1. Register the AUMID DisplayName in the registry (Action Center label).
	if k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Classes\AppUserModelId\`+notifyAUMID,
		registry.SET_VALUE,
	); err == nil {
		k.SetStringValue("DisplayName", "Internet Monitor")
		k.Close()
	}

	// 2. Create the Start Menu shortcut with AUMID property (runs once).
	//    This is what makes Windows 10/11 show banner popups instead of
	//    silently routing notifications to the Action Center only.
	go func() {
		log.Println("[notify] ensuring Start Menu shortcut…")
		EnsureStartMenuShortcut(notifyAUMID)
	}()
}

// Notify shows a system notification. Uses WinRT PowerShell toast (richer UI)
// if the Start Menu shortcut exists; falls back to Shell_NotifyIcon balloon.
func Notify(title, message string) {
	appdataDir := os.Getenv("APPDATA")
	lnk := filepath.Join(appdataDir,
		"Microsoft", "Windows", "Start Menu", "Programs",
		"Internet Monitor.lnk")

	if _, err := os.Stat(lnk); err == nil {
		// Shortcut registered → WinRT toast shows as proper banner popup.
		showWinRTToast(title, message)
	} else {
		// Shortcut not yet created → Shell_NotifyIcon balloon (Action Center).
		ShowBalloon(title, message)
	}
}

// ShowWinRTToast is exported so cmd/gui can call it directly.
// showWinRTToast shows a Windows 10/11 toast notification via PowerShell + WinRT.
// Requires the Start Menu shortcut to be already created (ensureStartMenuShortcut).
func ShowWinRTToast(title, body string) {
	showWinRTToast(title, body)
}

func showWinRTToast(title, body string) {
	t := strings.ReplaceAll(title, "'", "''")
	b := strings.ReplaceAll(body, "'", "''")

	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null
$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml('<toast><visual><binding template="ToastText02"><text id="1">%s</text><text id="2">%s</text></binding></visual></toast>')
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('%s').Show($toast)
`, t, b, notifyAUMID)

	cmd := exec.Command("powershell",
		"-WindowStyle", "Hidden",
		"-NonInteractive",
		"-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func OpenURL(url string) {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func OpenFolder(path string) {
	exec.Command("explorer", path).Start()
}
