// Package sound resolves and plays the notification chime. It is shared by the
// tray and GUI builds so the path-resolution and playback logic exist once.
//
// The native Windows player uses winmm PlaySound, which is WAV-only, so the
// embedded chime is a WAV. A user can override it with notification.wav next to
// the executable. (The legacy notification.mp3 custom sound is no longer used
// by the native player — PlaySound cannot decode MP3.)
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
// "notification.wav" in the working directory if present, otherwise the
// embedded default WAV extracted once to a temp file. Returns "" if unavailable.
func RingtonePath() string {
	if wd, err := os.Getwd(); err == nil {
		custom := filepath.Join(wd, "notification.wav")
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}
	once.Do(func() {
		data := dashboard.NotificationWav()
		if len(data) == 0 {
			return
		}
		dir, err := os.MkdirTemp("", "internet-monitor-")
		if err != nil {
			return
		}
		path := filepath.Join(dir, "notification.wav")
		if os.WriteFile(path, data, 0644) == nil {
			defaultPath = path
		}
	})
	return defaultPath
}
