//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modKernel32     = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex = modKernel32.NewProc("CreateMutexW")
)

func ensureSingleInstance() {
	name, _ := syscall.UTF16PtrFromString("Local\\InternetMonitor_4f8a2b1c")
	_, _, lastErr := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if lastErr == syscall.ERROR_ALREADY_EXISTS {
		os.Exit(0) // Already running — exit silently
	}
	// Don't close the handle; it lives until the process exits
}
