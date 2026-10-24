package tray

import (
	"internet-monitor/internal/alerts"
	"internet-monitor/internal/sound"
)

type Desktop struct {
	ActionURL func() string
}

func (d Desktop) Show(t alerts.Toast) {
	sound.PlayAt(t.Volume)
	action := ""
	if d.ActionURL != nil {
		action = d.ActionURL()
	}
	showDesktop(t, action)
}

func toastTag(kind alerts.Kind) (tag, group string) {
	switch kind {
	case alerts.KindSpeed:
		return "speed", "speed"
	case alerts.KindTest:
		return "test", "test"
	default:
		return "status", "connection"
	}
}
