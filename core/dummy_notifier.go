package core

import (
	"fmt"
	"internet-monitor/types"
)

// DummyNotifier prints events to stdout — demonstrates FR-2 extensibility.
// Adding a new notification channel requires only a new Notifier implementation;
// Engine, MultiNotifier, and all binaries need no modification.
type DummyNotifier struct{}

func (d *DummyNotifier) OnTick(_ types.CheckResult, status types.Status) {
	fmt.Printf("[DummyNotifier] tick: %s\n", status)
}

func (d *DummyNotifier) OnEvent(event types.Event) {
	fmt.Printf("[DummyNotifier] event: %s\n", event.EventType)
}
