//go:build !windows

package sound

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	playMu    sync.Mutex
	current   *exec.Cmd
	scaledDir string
	scaledFor string
)

func playerCommand(path string) *exec.Cmd {
	if runtime.GOOS == "darwin" {
		return exec.Command("afplay", path)
	}
	for _, p := range [][]string{{"paplay"}, {"pw-play"}, {"aplay", "-q"}, {"ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet"}} {
		if _, err := exec.LookPath(p[0]); err == nil {
			return exec.Command(p[0], append(p[1:], path)...)
		}
	}
	return nil
}

func scaledFile(path string, volume int) string {
	data, key, ok := scaled(path, volume)
	if !ok {
		return path
	}
	if scaledDir == "" {
		dir, err := os.MkdirTemp("", "internet-monitor-volume-")
		if err != nil {
			return path
		}
		scaledDir = dir
	}
	out := filepath.Join(scaledDir, "notification.wav")
	if scaledFor != key {
		if err := os.WriteFile(out, data, 0o644); err != nil {
			logf("[sound] write scaled sound: %v", err)
			return path
		}
		scaledFor = key
	}
	return out
}

func PlayAt(volume int) {
	if volume <= 0 {
		return
	}
	path := RingtonePath()
	if path == "" {
		return
	}
	playMu.Lock()
	defer playMu.Unlock()

	if volume < 100 {
		path = scaledFile(path, volume)
	}

	if current != nil && current.Process != nil {
		_ = current.Process.Kill()
		current = nil
	}

	cmd := playerCommand(path)
	if cmd == nil {
		logf("[sound] no audio player found (install pulseaudio-utils, pipewire or alsa-utils)")
		return
	}
	if err := cmd.Start(); err != nil {
		logf("[sound] player failed: %v", err)
		return
	}
	current = cmd
	go func(c *exec.Cmd) { _ = c.Wait() }(cmd)
}
