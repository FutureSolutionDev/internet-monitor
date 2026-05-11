//go:build windows

package main

import (
	"fmt"
	"internet-monitor/monitor"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const toastAppID = "InternetMonitor"

func init() {
	registerToastApp()
}

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

var (
	modWinmm       = syscall.NewLazyDLL("winmm.dll")
	procMciSendStr = modWinmm.NewProc("mciSendStringW")
)

func playRingtone() {
	path := getRingtonePath()
	if path == "" {
		return
	}
	go func() {
		openCmd, _ := syscall.UTF16PtrFromString(`open "` + path + `" type mpegvideo alias im_ring`)
		playCmd, _ := syscall.UTF16PtrFromString("play im_ring")
		stopCmd, _ := syscall.UTF16PtrFromString("close im_ring")

		procMciSendStr.Call(uintptr(unsafe.Pointer(openCmd)), 0, 0, 0)
		procMciSendStr.Call(uintptr(unsafe.Pointer(playCmd)), 0, 0, 0)
		time.Sleep(15 * time.Second)
		procMciSendStr.Call(uintptr(unsafe.Pointer(stopCmd)), 0, 0, 0)
	}()
}

func sendNotification(status monitor.Status, result monitor.CheckResult) {
	title, body := notifyText(status, result)
	if title == "" {
		return
	}
	playRingtone()
	showToast(title, body)
}

func TestNotification() {
	playRingtone()
	showToast("اختبار الإشعار / Test Notification", "🔔 الصوت والإشعار يعملان بشكل صحيح")
}

func showToast(title, body string) {
	t := strings.ReplaceAll(title, "'", "''")
	b := strings.ReplaceAll(body, "'", "''")
	script := fmt.Sprintf(`
$app='%s'
[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null
$tpl=[Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$nodes=$tpl.GetElementsByTagName('text')
$nodes[0].InnerText='%s'
$nodes[1].InnerText='%s'
$toast=[Windows.UI.Notifications.ToastNotification]::new($tpl)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($app).Show($toast)
`, toastAppID, t, b)
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}
