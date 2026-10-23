package core

import (
	"fmt"
	"internet-monitor/internal/types"
)

type DummyNotifier struct{}

func (d *DummyNotifier) OnTick(_ types.CheckResult, status types.Status) {
	fmt.Printf("[DummyNotifier] tick: %s\n", status)
}

func (d *DummyNotifier) OnEvent(event types.Event) {
	fmt.Printf("[DummyNotifier] event: %s\n", event.EventType)
}
