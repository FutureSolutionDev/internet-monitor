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
