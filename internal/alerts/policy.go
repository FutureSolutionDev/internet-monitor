package alerts

import (
	"internet-monitor/internal/config"
	"time"
)

type Kind string

const (
	KindDown     Kind = "down"
	KindDegraded Kind = "degraded"
	KindRestored Kind = "restored"
	KindReminder Kind = "reminder"
	KindSpeed    Kind = "speed"
	KindTest     Kind = "test"
)

type Channel string

const (
	ChannelDesktop  Channel = "desktop"
	ChannelSound    Channel = "sound"
	ChannelWebhook  Channel = "webhook"
	ChannelTelegram Channel = "telegram"
)

type Reason string

const (
	ReasonDisabled    Reason = "disabled"
	ReasonQuietHours  Reason = "quiet_hours"
	ReasonSoundOff    Reason = "sound_off"
	ReasonShortOutage Reason = "short_outage"
	ReasonStartup     Reason = "startup"
)

type Configured struct {
	Webhook  bool
	Telegram bool
}

type Decision struct {
	Desktop  bool
	Volume   int
	Webhook  bool
	Telegram bool
	Muted    map[Channel]Reason
}

func (d Decision) Sent() []Channel {
	out := []Channel{}
	if d.Desktop {
		out = append(out, ChannelDesktop)
	}
	if d.Volume > 0 {
		out = append(out, ChannelSound)
	}
	if d.Webhook {
		out = append(out, ChannelWebhook)
	}
	if d.Telegram {
		out = append(out, ChannelTelegram)
	}
	return out
}

func (d *Decision) mute(c Channel, r Reason) {
	if d.Muted == nil {
		d.Muted = map[Channel]Reason{}
	}
	d.Muted[c] = r
}

func switchesFor(kind Kind, n config.Notifications) config.ChannelSwitches {
	switch kind {
	case KindDown, KindReminder:
		return n.Events.Down
	case KindDegraded:
		return n.Events.Degraded
	case KindRestored:
		return n.Events.Restored
	case KindSpeed:
		return n.Events.Speed
	case KindTest:
		return config.ChannelSwitches{Desktop: true}
	}
	return config.ChannelSwitches{}
}

func Decide(kind Kind, n config.Notifications, has Configured, now time.Time) Decision {
	var d Decision
	sw := switchesFor(kind, n)
	quiet := kind != KindTest && InQuietHours(n.QuietHours, now)

	switch {
	case !sw.Desktop:
		d.mute(ChannelDesktop, ReasonDisabled)
	case quiet && n.QuietHours.Mode == config.QuietModeAll:
		d.mute(ChannelDesktop, ReasonQuietHours)
	default:
		d.Desktop = true
		switch {
		case !n.Sound.Enabled || n.Sound.Volume <= 0:
			d.mute(ChannelSound, ReasonSoundOff)
		case quiet:
			d.mute(ChannelSound, ReasonQuietHours)
		default:
			d.Volume = n.Sound.Volume
		}
	}

	if kind == KindTest {
		return d
	}
	if has.Webhook {
		if sw.Webhook {
			d.Webhook = true
		} else {
			d.mute(ChannelWebhook, ReasonDisabled)
		}
	}
	if has.Telegram {
		if sw.Telegram {
			d.Telegram = true
		} else {
			d.mute(ChannelTelegram, ReasonDisabled)
		}
	}
	return d
}

func InQuietHours(q config.QuietHours, now time.Time) bool {
	if !q.Enabled {
		return false
	}
	start, ok1 := config.ParseClock(q.Start)
	end, ok2 := config.ParseClock(q.End)
	if !ok1 || !ok2 || start == end {
		return false
	}
	m := now.Hour()*60 + now.Minute()
	day := int(now.Weekday())
	switch {
	case start < end:
		if m < start || m >= end {
			return false
		}
	case m >= start:
	case m < end:
		day = (day + 6) % 7
	default:
		return false
	}
	for _, d := range q.Days {
		if d == day {
			return true
		}
	}
	return false
}
