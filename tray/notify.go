// Platform-specific implementations:
//
//	notify_windows.go — Windows (PowerShell toast + rundll32)
//	notify_darwin.go  — macOS   (osascript + open)
//	notify_linux.go   — Linux   (notify-send + xdg-open)
package tray

// Logf, if set, receives notification-path diagnostics (wired to logger.AppLog
// by main so they land in logs/app.log; the standard log package is invisible
// in a -H=windowsgui build). Declared here (not in notify_windows.go) so main
// can wire it on every platform.
var Logf func(format string, args ...interface{})
