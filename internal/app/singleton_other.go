//go:build !windows

package app

import (
	"os"
	"path/filepath"
	"syscall"
)

var lockFile *os.File

func acquireInstance() bool {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	_ = os.MkdirAll(filepath.Join(dir, "internet-monitor"), 0o755)
	f, err := os.OpenFile(filepath.Join(dir, "internet-monitor", "instance.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return true
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return false
	}
	lockFile = f
	return true
}

func ReleaseSingleInstance() {
	if lockFile != nil {
		_ = lockFile.Close()
		lockFile = nil
	}
}
