//go:build windows

package main

import (
	"fmt"
	"internet-monitor/monitor"
	"os/exec"
	"strings"
	"syscall"
)

func sendNotification(status monitor.Status, result monitor.CheckResult) {
	title, body := notifyText(status, result)
	if title == "" {
		return
	}
	title = strings.ReplaceAll(title, "'", "''")
	body = strings.ReplaceAll(body, "'", "''")

	script := fmt.Sprintf(`
$app='Internet Monitor'
[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null
$tpl=[Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$nodes=$tpl.GetElementsByTagName('text')
$nodes[0].InnerText='%s'
$nodes[1].InnerText='%s'
$toast=[Windows.UI.Notifications.ToastNotification]::new($tpl)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($app).Show($toast)
`, title, body)

	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}
