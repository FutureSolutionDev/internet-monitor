// Package sound resolves and plays the notification ringtone. It is shared by
// the tray and GUI builds so the path-resolution and playback logic exist once.
package sound

import (
	"internet-monitor/dashboard"
	"os"
	"path/filepath"
	"sync"
)

var (
	defaultPath string
	once        sync.Once
)

// RingtonePath returns the sound file to play: a user-supplied
// "notification.mp3" in the working directory if present, otherwise the
// embedded default extracted once to a temp file. Returns "" if unavailable.
func RingtonePath() string {
	if wd, err := os.Getwd(); err == nil {
		custom := filepath.Join(wd, "notification.mp3")
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}
	once.Do(func() {
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
			defaultPath = path
		}
	})
	return defaultPath
}
