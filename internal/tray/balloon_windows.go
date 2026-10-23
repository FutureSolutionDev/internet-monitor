//go:build windows

package tray

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

const balloonCooldown = 4 * time.Second

var (
	balloonMu   sync.Mutex
	lastBalloon time.Time
)

func showBalloonWinForms(title, message string) {
	t := strings.ReplaceAll(title, "'", "''")
	m := strings.ReplaceAll(message, "'", "''")

	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$n = [System.Windows.Forms.NotifyIcon]::new()
$n.Icon = [System.Drawing.SystemIcons]::Application
$n.BalloonTipIcon = [System.Windows.Forms.ToolTipIcon]::Info
$n.BalloonTipTitle = '%s'
$n.BalloonTipText  = '%s'
$n.Visible = $true
$n.ShowBalloonTip(5000)
Start-Sleep -Seconds 6
$n.Visible = $false
$n.Dispose()`, t, m)

	cmd := exec.Command("powershell",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-WindowStyle", "Hidden",
		"-NonInteractive",
		"-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	go func() {
		if out, err := cmd.CombinedOutput(); err != nil {
			notifyLogf("[notify] WinForms balloon error: %v output=%q", err, strings.TrimSpace(string(out)))
		}
	}()
}

func ShowBalloon(title, message string) {
	balloonMu.Lock()
	if time.Since(lastBalloon) < balloonCooldown {
		balloonMu.Unlock()
		notifyLogf("[notify] ShowBalloon: debounced")
		return
	}
	lastBalloon = time.Now()
	balloonMu.Unlock()
	showBalloonWinForms(title, message)
}
