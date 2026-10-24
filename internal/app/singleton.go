package app

import (
	"internet-monitor/internal/updater"
	"os"
	"time"
)

func AcquireSingleInstance() bool {
	if os.Getenv(updater.RelaunchEnv) == "" {
		return acquireInstance()
	}
	_ = os.Unsetenv(updater.RelaunchEnv)
	deadline := time.Now().Add(15 * time.Second)
	for {
		if acquireInstance() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(200 * time.Millisecond)
	}
}
