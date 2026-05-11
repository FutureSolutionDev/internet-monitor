//go:build !windows

package startup

func Supported() bool        { return false }
func IsEnabled() bool        { return false }
func SetEnabled(_ bool) error { return nil }
