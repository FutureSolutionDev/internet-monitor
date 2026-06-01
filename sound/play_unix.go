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

// Play stops any currently-playing ringtone process and starts the latest one,
// so rapid notifications never overlap. macOS uses afplay; Linux prefers mpg123
// and falls back to ffplay. Non-blocking: the player runs as a child process.
func Play() {
	path := RingtonePath()
	if path == "" {
		return
	}
	playMu.Lock()
	defer playMu.Unlock()

	// Kill the previous player (if still running) so audio doesn't overlap.
	if current != nil && current.Process != nil {
		_ = current.Process.Kill()
		current = nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("afplay", path)
	default: // linux
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
	// Reap the process when it finishes so it doesn't linger as a zombie.
	go func(c *exec.Cmd) { _ = c.Wait() }(cmd)
}
