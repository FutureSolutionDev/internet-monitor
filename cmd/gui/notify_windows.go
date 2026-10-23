//go:build windows

package main

import (
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/sound"
	"internet-monitor/internal/tray"
	"log"
	"sync"
	"time"
)

func playRingtone() { sound.Play() }

var (
	notifyMu       sync.Mutex
	lastNotifyTime time.Time
)

const guiNotifyCooldown = 4 * time.Second

func showSystemNotification(title, body string) {

	tray.ShowNotification(title, body)
}

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
	showSystemNotification(title, body)
}

func TestNotification(lang string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[notify] TestNotification panic recovered: %v", r)
		}
	}()
	playRingtone()
	title, body := notifytext.TestMessage(lang)
	showSystemNotification(title, body)
}
