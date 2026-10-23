//go:build windows

package tray

import "internet-monitor/internal/sound"

func playTraySound() { sound.Play() }
