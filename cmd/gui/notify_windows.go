//go:build windows

package main

import (
	"fmt"
	"internet-monitor/monitor"
	"os/exec"
	"strings"
	"sync"
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

	ringMu      sync.Mutex
	ringPlaying bool
)

func mciCall(cmd string) {
	p, _ := syscall.UTF16PtrFromString(cmd)
	procMciSendStr.Call(uintptr(unsafe.Pointer(p)), 0, 0, 0)
}

func playRingtone() {
	path := getRingtonePath()
	if path == "" {
		return
	}
	go func() {
		ringMu.Lock()
		if ringPlaying {
			// Stop whatever is playing before starting a new one.
			mciCall("stop im_ring")
			mciCall("close im_ring")
		}
		ringPlaying = true
		ringMu.Unlock()

		mciCall(`open "` + path + `" type mpegvideo alias im_ring`)
		mciCall("play im_ring")
		time.Sleep(15 * time.Second)
		mciCall("stop im_ring")
		mciCall("close im_ring")

		ringMu.Lock()
		ringPlaying = false
		ringMu.Unlock()
	}()
}

var (
	notifyMu       sync.Mutex
	lastNotifyTime time.Time
)

const guiNotifyCooldown = 4 * time.Second

func sendNotification(status monitor.Status, result monitor.CheckResult) {
	title, body := notifyText(status, result)
	if title == "" {
		return
	}
	notifyMu.Lock()
	if time.Since(lastNotifyTime) < guiNotifyCooldown {
		notifyMu.Unlock()
		return
	}
	lastNotifyTime = time.Now()
	notifyMu.Unlock()

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
