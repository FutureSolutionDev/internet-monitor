//go:build windows

package app

import (
	"golang.org/x/sys/windows"
)

const mutexName = `Local\InternetMonitor_4f8a2b1c`

var instanceMutex windows.Handle

func acquireInstance() bool {
	name, err := windows.UTF16PtrFromString(mutexName)
	if err != nil {
		return true
	}
	h, err := windows.CreateMutex(nil, false, name)
	if err == windows.ERROR_ALREADY_EXISTS {
		if h != 0 {
			_ = windows.CloseHandle(h)
		}
		return false
	}
	instanceMutex = h
	return true
}

func ReleaseSingleInstance() {
	if instanceMutex != 0 {
		_ = windows.CloseHandle(instanceMutex)
		instanceMutex = 0
	}
}
