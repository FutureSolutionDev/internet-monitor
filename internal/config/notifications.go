package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	QuietModeSound = "sound"
	QuietModeAll   = "all"

	MaxMinOutageSeconds = 3600
	MaxReminderMinutes  = 1440
)

type SoundSettings struct {
	Enabled bool `json:"enabled"`
	Volume  int  `json:"volume"`
}

type QuietHours struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Days    []int  `json:"days"`
	Mode    string `json:"mode"`
}

type ChannelSwitches struct {
	Desktop  bool `json:"desktop"`
	Webhook  bool `json:"webhook"`
	Telegram bool `json:"telegram"`
}

type EventSwitches struct {
	Down     ChannelSwitches `json:"down"`
	Degraded ChannelSwitches `json:"degraded"`
	Restored ChannelSwitches `json:"restored"`
	Speed    ChannelSwitches `json:"speed"`
}

type Notifications struct {
	Sound            SoundSettings `json:"sound"`
	QuietHours       QuietHours    `json:"quiet_hours"`
	Events           EventSwitches `json:"events"`
	MinOutageSeconds int           `json:"min_outage_seconds"`
	ReminderMinutes  int           `json:"reminder_minutes"`
}

func allDays() []int { return []int{0, 1, 2, 3, 4, 5, 6} }

func DefaultNotifications() Notifications {
	on := ChannelSwitches{Desktop: true, Webhook: true, Telegram: true}
	return Notifications{
		Sound: SoundSettings{Enabled: true, Volume: 100},
		QuietHours: QuietHours{
			Start: "22:00",
			End:   "07:00",
			Days:  allDays(),
			Mode:  QuietModeSound,
		},
		Events: EventSwitches{
			Down:     on,
			Degraded: on,
			Restored: on,
			Speed:    ChannelSwitches{Webhook: true, Telegram: true},
		},
	}
}

func (n Notifications) Clone() Notifications {
	out := n
	if n.QuietHours.Days != nil {
		out.QuietHours.Days = append([]int(nil), n.QuietHours.Days...)
	}
	return out
}

func (n *Notifications) Sanitize() {
	def := DefaultNotifications()
	n.Sound.Volume = clamp(n.Sound.Volume, 0, 100)
	if _, ok := ParseClock(n.QuietHours.Start); !ok {
		n.QuietHours.Start = def.QuietHours.Start
	}
	if _, ok := ParseClock(n.QuietHours.End); !ok {
		n.QuietHours.End = def.QuietHours.End
	}
	n.QuietHours.Start = normalizeClock(n.QuietHours.Start)
	n.QuietHours.End = normalizeClock(n.QuietHours.End)
	if n.QuietHours.Days == nil {
		n.QuietHours.Days = allDays()
	} else {
		n.QuietHours.Days = cleanDays(n.QuietHours.Days)
	}
	if n.QuietHours.Mode != QuietModeSound && n.QuietHours.Mode != QuietModeAll {
		n.QuietHours.Mode = QuietModeSound
	}
	n.MinOutageSeconds = clamp(n.MinOutageSeconds, 0, MaxMinOutageSeconds)
	n.ReminderMinutes = clamp(n.ReminderMinutes, 0, MaxReminderMinutes)
}

func ParseClock(s string) (int, bool) {
	h, m, ok := strings.Cut(strings.TrimSpace(s), ":")
	if !ok || len(m) != 2 || len(h) < 1 || len(h) > 2 {
		return 0, false
	}
	hh, err1 := strconv.Atoi(h)
	mm, err2 := strconv.Atoi(m)
	if err1 != nil || err2 != nil || hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return 0, false
	}
	return hh*60 + mm, true
}

func normalizeClock(s string) string {
	mins, _ := ParseClock(s)
	return fmt.Sprintf("%02d:%02d", mins/60, mins%60)
}

func cleanDays(in []int) []int {
	seen := [7]bool{}
	out := make([]int, 0, 7)
	for _, d := range in {
		if d < 0 || d > 6 || seen[d] {
			continue
		}
		seen[d] = true
		out = append(out, d)
	}
	sort.Ints(out)
	return out
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
