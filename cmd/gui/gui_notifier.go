package main

import (
	"internet-monitor/internal/types"
)

type guiNotifier struct{}

func (g *guiNotifier) OnTick(_ types.CheckResult, status types.Status) {
	updateTrayStatus(status)
}

func (g *guiNotifier) OnEvent(types.Event) {}
