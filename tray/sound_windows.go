//go:build windows

package tray

import "internet-monitor/sound"

// playTraySound plays the notification ringtone (shared with the GUI build).
func playTraySound() { sound.Play() }
