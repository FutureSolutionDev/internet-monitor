//go:build windows

package tray

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const toastAppID = "InternetMonitor"

func init() {
	registerToastApp()
}

// registerToastApp registers the app AUMID in HKCU so Windows shows
// toast notifications as banners (not just in Action Center).
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

func Notify(title, message string) {
	title = strings.ReplaceAll(title, "'", "''")
	message = strings.ReplaceAll(message, "'", "''")

	script := fmt.Sprintf(`
$app = '%s'
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$tpl = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$nodes = $tpl.GetElementsByTagName('text')
$nodes[0].InnerText = '%s'
$nodes[1].InnerText = '%s'
$toast = [Windows.UI.Notifications.ToastNotification]::new($tpl)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($app).Show($toast)
`, toastAppID, title, message)

	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-NonInteractive", "-Command", script)
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
