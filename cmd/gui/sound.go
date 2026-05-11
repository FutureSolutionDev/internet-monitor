package main

import (
	"internet-monitor/dashboard"
	"os"
	"path/filepath"
	"sync"
)

var (
	ringtonePath string
	ringtoneOnce sync.Once
)

// getRingtonePath extracts Ringtone.mp3 from embedded assets to a temp file
// on first call, then returns the same path on subsequent calls.
func getRingtonePath() string {
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
			ringtonePath = path
		}
	})
	return ringtonePath
}
