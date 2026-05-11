package logger

import "internet-monitor/types"

// LogNotifier is a no-op Notifier — the engine calls lgr.Log() directly.
// This stub satisfies the core.Notifier interface for future webhook-only use.
type LogNotifier struct {
	l *Logger
}

func NewNotifier(l *Logger) *LogNotifier {
	return &LogNotifier{l: l}
}

func (n *LogNotifier) OnTick(_ types.CheckResult, _ types.Status) {}

func (n *LogNotifier) OnEvent(_ types.Event) {}
