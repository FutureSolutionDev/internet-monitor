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

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("afplay", path)
	default:
		if _, err := exec.LookPath("mpg123"); err == nil {
			cmd = exec.Command("mpg123", "-q", path)
		} else {
			cmd = exec.Command("ffplay", "-nodisp", "-autoexit", path)
		}
	}
	if err := cmd.Start(); err != nil {
		return
	}
	current = cmd

	go func(c *exec.Cmd) { _ = c.Wait() }(cmd)
}
