// Platform-specific implementations:
//   notify_windows.go — Windows (PowerShell toast + rundll32)
//   notify_darwin.go  — macOS   (osascript + open)
//   notify_linux.go   — Linux   (notify-send + xdg-open)
package tray
