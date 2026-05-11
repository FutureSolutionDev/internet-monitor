//go:build windows

package startup

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const runKey  = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
const appName = "InternetMonitor"

func Supported() bool { return true }

func IsEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(appName)
	return err == nil
}

func SetEnabled(enabled bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if enabled {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		return k.SetStringValue(appName, `"`+exe+`"`)
	}

	k.DeleteValue(appName) // ignore error if value doesn't exist
	return nil
}
