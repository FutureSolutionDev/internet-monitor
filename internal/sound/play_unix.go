//go:build !windows

package sound

import (
	"os/exec"
	"runtime"
	"sync"
)

var (
	playMu  sync.Mutex
	current *exec.Cmd
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

func Play() {
	path := RingtonePath()
	if path == "" {
		return
	}
	playMu.Lock()
	defer playMu.Unlock()

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
