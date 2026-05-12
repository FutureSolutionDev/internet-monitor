//go:build windows

package tray

import (
	"internet-monitor/dashboard"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	modWinmmTray   = syscall.NewLazyDLL("winmm.dll")
	procMciTray    = modWinmmTray.NewProc("mciSendStringW")
	trayRingMu     sync.Mutex
	trayRingPlaying bool

	defaultTrayRingPath string
	trayRingOnce        sync.Once
)

func mciCallTray(cmd string) {
	p, _ := syscall.UTF16PtrFromString(cmd)
	procMciTray.Call(uintptr(unsafe.Pointer(p)), 0, 0, 0)
}

func getTrayRingtonePath() string {
	exeDir, err := os.Getwd()
	if err == nil {
		custom := filepath.Join(exeDir, "notification.mp3")
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}
	trayRingOnce.Do(func() {
		data := dashboard.RingtoneMp3()
		if len(data) == 0 {
			return
		}
		dir, err := os.MkdirTemp("", "internet-monitor-")
		if err != nil {
			return
		}
		path := filepath.Join(dir, "Ringtone.mp3")
		if os.WriteFile(path, data, 0644) == nil {
			defaultTrayRingPath = path
		}
	})
	return defaultTrayRingPath
}

func playTraySound() {
	path := getTrayRingtonePath()
	if path == "" {
		return
	}
	trayRingMu.Lock()
	if trayRingPlaying {
		mciCallTray("stop tray_ring")
		mciCallTray("close tray_ring")
	}
	trayRingPlaying = true
	trayRingMu.Unlock()

	mciCallTray(`open "` + path + `" type mpegvideo alias tray_ring`)
	mciCallTray("play tray_ring")
	time.Sleep(15 * time.Second)
	mciCallTray("stop tray_ring")
	mciCallTray("close tray_ring")

	trayRingMu.Lock()
	trayRingPlaying = false
	trayRingMu.Unlock()
}
