package tray

import (
	"fmt"
	"os/exec"
	"strings"
)

func Notify(title, message string) {
	// Sanitize for PowerShell single-quoted string
	title = strings.ReplaceAll(title, "'", "''")
	message = strings.ReplaceAll(message, "'", "''")

	script := fmt.Sprintf(`
$app = 'Internet Monitor'
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$tpl = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent(
    [Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$nodes = $tpl.GetElementsByTagName('text')
$nodes[0].InnerText = '%s'
$nodes[1].InnerText = '%s'
$toast = [Windows.UI.Notifications.ToastNotification]::new($tpl)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($app).Show($toast)
`, title, message)

	exec.Command("powershell", "-WindowStyle", "Hidden", "-NonInteractive", "-Command", script).Start()
}
