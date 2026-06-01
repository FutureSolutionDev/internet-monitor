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

// Each balloon spawns a powershell.exe that lives ~6s; without a guard, rapid
// notifications (e.g. mashing the test button) spawn a storm of processes.
// balloonCooldown debounces calls so at most one balloon fires per window.
const balloonCooldown = 4 * time.Second

var (
	balloonMu   sync.Mutex
	lastBalloon time.Time
)

// showBalloonWinForms shows a notification balloon via PowerShell +
// System.Windows.Forms.NotifyIcon. This borrows powershell.exe's notification
// identity, so it renders reliably from an unpackaged Go exe — unlike a WinRT
// toast or a Shell_NotifyIcon balloon on our own (identity-less) window, both
// of which report success but render nothing on Win10/11.
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

// ShowBalloon shows a notification balloon (PowerShell WinForms), debounced so
// rapid calls can't spawn a storm of powershell processes.
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
