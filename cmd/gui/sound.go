package main

import (
	"internet-monitor/dashboard"
	"os"
	"path/filepath"
	"sync"
)

var (
	defaultRingtonePath string
	ringtoneOnce        sync.Once
)

// getRingtonePath returns the path to the sound file to play.
// Priority: notification.mp3 next to the executable > embedded default.
func getRingtonePath() string {
	// Check for user-supplied custom sound next to the exe first.
	exeDir, err := os.Getwd()
	if err == nil {
		custom := filepath.Join(exeDir, "notification.mp3")
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}

	// Fall back to extracted embedded default.
	ringtoneOnce.Do(func() {
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
			defaultRingtonePath = path
		}
	})
	return defaultRingtonePath
}
