package main

import "internet-monitor/sound"

// getRingtonePath resolves the ringtone via the shared sound package.
func getRingtonePath() string { return sound.RingtonePath() }
