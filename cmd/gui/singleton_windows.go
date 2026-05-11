//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modKernel32GUI     = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexGUI = modKernel32GUI.NewProc("CreateMutexW")
)

func ensureSingleInstance() {
	name, _ := syscall.UTF16PtrFromString("Local\\InternetMonitorGUI_4f8a2b1c")
	_, _, lastErr := procCreateMutexGUI.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if lastErr == syscall.ERROR_ALREADY_EXISTS {
		os.Exit(0)
	}
}
