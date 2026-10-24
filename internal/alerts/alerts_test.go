package alerts

import (
	"internet-monitor/internal/config"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/types"
	"reflect"
	"sync"
	"testing"
	"time"
)

type fakeTimer struct {
	at      time.Time
	fn      func()
	stopped bool
}

func (t *fakeTimer) Stop() bool {
	was := !t.stopped
	t.stopped = true
	return was
}

type fakeClock struct {
	now    time.Time
	timers []*fakeTimer
}

func (c *fakeClock) after(d time.Duration, fn func()) stopper {
	t := &fakeTimer{at: c.now.Add(d), fn: fn}
	c.timers = append(c.timers, t)
	return t
}

func (c *fakeClock) advance(d time.Duration) {
	c.now = c.now.Add(d)
	for i := 0; i < len(c.timers); i++ {
		t := c.timers[i]
		if !t.stopped && !t.at.After(c.now) {
			t.stopped = true
			t.fn()
		}
	}
}

type sentEvent struct {
	kind              string
	duration          float64
	webhook, telegram bool
}

type fakeRemote struct {
	mu     sync.Mutex
	events []sentEvent
	speeds int
}

func (r *fakeRemote) SendEvent(_ string, ev types.Event, kind string, webhook, telegram bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, sentEvent{kind, ev.DurationSeconds, webhook, telegram})
}

func (r *fakeRemote) SendSpeed(string, logger.SpeedTestEvent, float64, bool, bool, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.speeds++
}

type fakeDesktop struct {
	mu     sync.Mutex
	toasts []Toast
}

func (d *fakeDesktop) Show(t Toast) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.toasts = append(d.toasts, t)
}

type harness struct {
	a       *Alerts
	clock   *fakeClock
	remote  *fakeRemote
	desktop *fakeDesktop
	records []Record
}

var monday = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

func baseConfig() config.Config {
	cfg := config.Default()
	cfg.WebhookURL = "https://discord.com/api/webhooks/1/abc"
	cfg.TelegramBotToken = "token"
	cfg.TelegramChatID = "chat"
	return cfg
}

func newHarness(t *testing.T, cfg config.Config) *harness {
	t.Helper()
	h := &harness{clock: &fakeClock{now: monday}, remote: &fakeRemote{}, desktop: &fakeDesktop{}}
	h.a = New(cfg, Options{
		Desktop: h.desktop,
		Remote:  h.remote,
		Publish: func(r Record) { h.records = append(h.records, r) },
	})
	h.a.now = func() time.Time { return h.clock.now }
	h.a.after = h.clock.after
	t.Cleanup(h.a.Close)
	return h
}

func (h *harness) event(eventType string) types.Event {
	return types.Event{Timestamp: h.clock.now, EventType: eventType, Reason: types.EventReason{TCPPingFailed: eventType == "disconnected"}}
}

func (h *harness) kinds() []Kind {
	out := []Kind{}
	for _, r := range h.records {
		out = append(out, r.Kind)
	}
	return out
}

func (h *harness) drain() []Toast {
	h.a.Close()
	h.desktop.mu.Lock()
	defer h.desktop.mu.Unlock()
	return h.desktop.toasts
}

func TestDefaultsSendEverythingImmediately(t *testing.T) {
	h := newHarness(t, baseConfig())
	h.a.OnEvent(h.event("disconnected"))
	h.clock.advance(90 * time.Second)
	h.a.OnEvent(h.event("connected"))

	if got := h.kinds(); !reflect.DeepEqual(got, []Kind{KindDown, KindRestored}) {
		t.Fatalf("kinds = %v", got)
	}
	want := []Channel{ChannelDesktop, ChannelSound, ChannelWebhook, ChannelTelegram}
	if !reflect.DeepEqual(h.records[0].Sent, want) || len(h.records[0].Muted) != 0 {
		t.Fatalf("down sent=%v muted=%v", h.records[0].Sent, h.records[0].Muted)
	}
	if h.records[1].DurationSeconds != 90 {
		t.Errorf("restored duration = %v", h.records[1].DurationSeconds)
	}
	if len(h.remote.events) != 2 || h.remote.events[1].duration != 90 {
		t.Errorf("remote = %+v", h.remote.events)
	}
	toasts := h.drain()
	if len(toasts) != 2 || toasts[0].Volume != 100 || toasts[0].Status != types.StatusDisconnected {
		t.Fatalf("toasts = %+v", toasts)
	}
}

func TestFastFlapIsNotDropped(t *testing.T) {
	h := newHarness(t, baseConfig())
	h.a.OnEvent(h.event("disconnected"))
	h.clock.advance(time.Second)
	h.a.OnEvent(h.event("connected"))
	if toasts := h.drain(); len(toasts) != 2 {
		t.Fatalf("restored toast lost to a cooldown: %d toasts", len(toasts))
	}
}

func TestInitialProblemGoesRemoteOnly(t *testing.T) {
	h := newHarness(t, baseConfig())
	h.a.OnInitialEvent(h.event("disconnected"))
	if len(h.records) != 1 || h.records[0].Muted[ChannelDesktop] != ReasonStartup {
		t.Fatalf("records = %+v", h.records)
	}
	if len(h.remote.events) != 1 || !h.remote.events[0].webhook {
		t.Fatalf("remote = %+v", h.remote.events)
	}
	h.a.OnEvent(h.event("connected"))
	if toasts := h.drain(); len(toasts) != 1 || toasts[0].Kind != KindRestored {
		t.Fatalf("toasts = %+v", toasts)
	}
	h2 := newHarness(t, baseConfig())
	h2.a.OnInitialEvent(h2.event("connected"))
	if len(h2.records) != 0 {
		t.Fatalf("initial connected recorded: %+v", h2.records)
	}
}

func TestDecideMatrix(t *testing.T) {
	quietAll := config.DefaultNotifications()
	quietAll.QuietHours = config.QuietHours{Enabled: true, Start: "00:00", End: "23:59", Days: []int{1}, Mode: config.QuietModeAll}
	quietSound := quietAll.Clone()
	quietSound.QuietHours.Mode = config.QuietModeSound
	noTelegram := config.DefaultNotifications()
	noTelegram.Events.Down.Telegram = false
	mutedSound := config.DefaultNotifications()
	mutedSound.Sound.Volume = 0
	soundOff := config.DefaultNotifications()
	soundOff.Sound.Enabled = false
	all := Configured{Webhook: true, Telegram: true}

	cases := []struct {
		name   string
		kind   Kind
		n      config.Notifications
		has    Configured
		sent   []Channel
		muted  map[Channel]Reason
		volume int
	}{
		{"defaults", KindDown, config.DefaultNotifications(), all, []Channel{ChannelDesktop, ChannelSound, ChannelWebhook, ChannelTelegram}, nil, 100},
		{"unconfigured remotes are not listed", KindDown, config.DefaultNotifications(), Configured{}, []Channel{ChannelDesktop, ChannelSound}, nil, 100},
		{"channel switched off", KindDown, noTelegram, all, []Channel{ChannelDesktop, ChannelSound, ChannelWebhook}, map[Channel]Reason{ChannelTelegram: ReasonDisabled}, 100},
		{"reminder follows down", KindReminder, noTelegram, all, []Channel{ChannelDesktop, ChannelSound, ChannelWebhook}, map[Channel]Reason{ChannelTelegram: ReasonDisabled}, 100},
		{"speed desktop off by default", KindSpeed, config.DefaultNotifications(), all, []Channel{ChannelWebhook, ChannelTelegram}, map[Channel]Reason{ChannelDesktop: ReasonDisabled}, 0},
		{"quiet hours mute sound only", KindDown, quietSound, all, []Channel{ChannelDesktop, ChannelWebhook, ChannelTelegram}, map[Channel]Reason{ChannelSound: ReasonQuietHours}, 0},
		{"quiet hours mute the desktop", KindDown, quietAll, all, []Channel{ChannelWebhook, ChannelTelegram}, map[Channel]Reason{ChannelDesktop: ReasonQuietHours}, 0},
		{"volume zero is sound off", KindDown, mutedSound, Configured{}, []Channel{ChannelDesktop}, map[Channel]Reason{ChannelSound: ReasonSoundOff}, 0},
		{"sound disabled", KindRestored, soundOff, Configured{}, []Channel{ChannelDesktop}, map[Channel]Reason{ChannelSound: ReasonSoundOff}, 0},
		{"test ignores quiet hours", KindTest, quietAll, all, []Channel{ChannelDesktop, ChannelSound}, nil, 100},
		{"test respects sound off", KindTest, soundOff, all, []Channel{ChannelDesktop}, map[Channel]Reason{ChannelSound: ReasonSoundOff}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := Decide(c.kind, c.n, c.has, monday)
			if !reflect.DeepEqual(d.Sent(), c.sent) {
				t.Errorf("sent = %v, want %v", d.Sent(), c.sent)
			}
			if len(d.Muted) != len(c.muted) || (len(c.muted) > 0 && !reflect.DeepEqual(d.Muted, c.muted)) {
				t.Errorf("muted = %v, want %v", d.Muted, c.muted)
			}
			if d.Volume != c.volume {
				t.Errorf("volume = %d, want %d", d.Volume, c.volume)
			}
		})
	}
}

func TestQuietHours(t *testing.T) {
	night := config.QuietHours{Enabled: true, Start: "22:00", End: "07:00", Days: []int{1}, Mode: config.QuietModeSound}
	day := config.QuietHours{Enabled: true, Start: "09:00", End: "17:00", Days: []int{1}}
	cairo := time.FixedZone("EET", 2*3600)
	at := func(d, h, m int) time.Time { return time.Date(2026, 10, d, h, m, 0, 0, time.UTC) }
	cases := []struct {
		name string
		q    config.QuietHours
		now  time.Time
		want bool
	}{
		{"monday night", night, at(19, 23, 0), true},
		{"after midnight belongs to monday", night, at(20, 3, 0), true},
		{"end is exclusive", night, at(20, 7, 0), false},
		{"tuesday night is not listed", night, at(20, 23, 0), false},
		{"monday early morning belongs to sunday", night, at(19, 3, 0), false},
		{"same-day window", day, at(19, 12, 0), true},
		{"same-day window end", day, at(19, 17, 0), false},
		{"disabled", config.QuietHours{Start: "00:00", End: "23:59", Days: []int{1}}, at(19, 12, 0), false},
		{"start equals end is off", config.QuietHours{Enabled: true, Start: "08:00", End: "08:00", Days: []int{1}}, at(19, 8, 0), false},
		{"no days", config.QuietHours{Enabled: true, Start: "00:00", End: "23:59", Days: []int{}}, at(19, 12, 0), false},
		{"uses local wall clock", night, time.Date(2026, 10, 19, 21, 30, 0, 0, time.UTC).In(cairo), true},
	}
	for _, c := range cases {
		if got := InQuietHours(c.q, c.now); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func deferConfig(seconds int) config.Config {
	cfg := baseConfig()
	cfg.Notifications.MinOutageSeconds = seconds
	return cfg
}

func TestShortOutageIsSuppressed(t *testing.T) {
	h := newHarness(t, deferConfig(30))
	h.a.OnEvent(h.event("disconnected"))
	h.clock.advance(10 * time.Second)
	h.a.OnEvent(h.event("connected"))
	h.clock.advance(time.Minute)

	if len(h.records) != 1 || h.records[0].Kind != KindDown || h.records[0].Muted[ChannelDesktop] != ReasonShortOutage {
		t.Fatalf("records = %+v", h.records)
	}
	if h.records[0].DurationSeconds != 10 || len(h.records[0].Sent) != 0 {
		t.Errorf("short record = %+v", h.records[0])
	}
	if len(h.remote.events) != 0 || len(h.drain()) != 0 {
		t.Fatal("short outage was announced")
	}
}

func TestLongOutageIsAnnouncedAfterTheDelay(t *testing.T) {
	h := newHarness(t, deferConfig(30))
	h.a.OnEvent(h.event("degraded"))
	h.clock.advance(5 * time.Second)
	h.a.OnEvent(h.event("disconnected"))
	h.clock.advance(20 * time.Second)
	if len(h.records) != 0 {
		t.Fatalf("announced too early: %+v", h.records)
	}
	h.clock.advance(5 * time.Second)
	if got := h.kinds(); !reflect.DeepEqual(got, []Kind{KindDown}) {
		t.Fatalf("kinds = %v", got)
	}
	h.clock.advance(time.Minute)
	h.a.OnEvent(h.event("connected"))
	if got := h.kinds(); !reflect.DeepEqual(got, []Kind{KindDown, KindRestored}) {
		t.Fatalf("kinds = %v", got)
	}
	if h.records[1].DurationSeconds != 90 {
		t.Errorf("outage duration = %v", h.records[1].DurationSeconds)
	}
}

func TestLoweringTheDelayFiresAPendingOutage(t *testing.T) {
	h := newHarness(t, deferConfig(300))
	h.a.OnEvent(h.event("disconnected"))
	h.clock.advance(40 * time.Second)
	h.a.ApplyConfig(deferConfig(30))
	if got := h.kinds(); !reflect.DeepEqual(got, []Kind{KindDown}) {
		t.Fatalf("kinds = %v", got)
	}
	h.clock.advance(10 * time.Minute)
	if len(h.records) != 1 {
		t.Fatalf("old timer fired again: %v", h.kinds())
	}
}

func TestRemindersRepeatUntilRestored(t *testing.T) {
	cfg := baseConfig()
	cfg.Notifications.ReminderMinutes = 5
	h := newHarness(t, cfg)
	h.a.OnEvent(h.event("disconnected"))
	h.clock.advance(5 * time.Minute)
	h.clock.advance(5 * time.Minute)
	h.a.OnEvent(h.event("connected"))
	h.clock.advance(30 * time.Minute)

	want := []Kind{KindDown, KindReminder, KindReminder, KindRestored}
	if got := h.kinds(); !reflect.DeepEqual(got, want) {
		t.Fatalf("kinds = %v", got)
	}
	if h.records[1].DurationSeconds != 300 || h.records[2].DurationSeconds != 600 {
		t.Errorf("reminder durations = %v, %v", h.records[1].DurationSeconds, h.records[2].DurationSeconds)
	}
	if h.remote.events[1].kind != "reminder" {
		t.Errorf("remote reminder = %+v", h.remote.events[1])
	}
}

func TestNoReminderWhileDegraded(t *testing.T) {
	cfg := baseConfig()
	cfg.Notifications.ReminderMinutes = 1
	h := newHarness(t, cfg)
	h.a.OnEvent(h.event("degraded"))
	h.clock.advance(10 * time.Minute)
	if got := h.kinds(); !reflect.DeepEqual(got, []Kind{KindDegraded}) {
		t.Fatalf("kinds = %v", got)
	}
}

func TestChangingTheReminderReArmsIt(t *testing.T) {
	cfg := baseConfig()
	cfg.Notifications.ReminderMinutes = 60
	h := newHarness(t, cfg)
	h.a.OnEvent(h.event("disconnected"))
	h.clock.advance(time.Minute)
	cfg.Notifications.ReminderMinutes = 2
	h.a.ApplyConfig(cfg)
	h.clock.advance(2 * time.Minute)
	if got := h.kinds(); !reflect.DeepEqual(got, []Kind{KindDown, KindReminder}) {
		t.Fatalf("kinds = %v", got)
	}
	cfg.Notifications.ReminderMinutes = 0
	h.a.ApplyConfig(cfg)
	h.clock.advance(time.Hour)
	if len(h.records) != 2 {
		t.Fatalf("reminder kept firing: %v", h.kinds())
	}
}

func TestSpeedAndTest(t *testing.T) {
	h := newHarness(t, baseConfig())
	h.a.OnSpeedResult(logger.SpeedTestEvent{DownloadMbps: 4}, 50, true)
	h.a.Test("en", nil)
	if got := h.kinds(); !reflect.DeepEqual(got, []Kind{KindSpeed, KindTest}) {
		t.Fatalf("kinds = %v", got)
	}
	if h.records[0].Speed == nil || !h.records[0].Speed.Below || h.remote.speeds != 1 {
		t.Fatalf("speed record = %+v, remote=%d", h.records[0], h.remote.speeds)
	}
	toasts := h.drain()
	if len(toasts) != 1 || toasts[0].Kind != KindTest {
		t.Fatalf("toasts = %+v", toasts)
	}
}

func TestTestUsesOnScreenVolume(t *testing.T) {
	h := newHarness(t, baseConfig())
	quiet := 30
	h.a.Test("en", &quiet)
	off := 0
	h.a.Test("en", &off)
	toasts := h.drain()
	if len(toasts) != 2 || toasts[0].Volume != 30 || toasts[1].Volume != 0 {
		t.Fatalf("toasts = %+v", toasts)
	}
	if h.records[1].Muted[ChannelSound] != ReasonSoundOff {
		t.Fatalf("muted = %v", h.records[1].Muted)
	}
}

func TestClosedAlertsIgnoreEvents(t *testing.T) {
	h := newHarness(t, deferConfig(30))
	h.a.OnEvent(h.event("disconnected"))
	h.a.Close()
	h.clock.advance(time.Minute)
	h.a.OnEvent(h.event("connected"))
	if len(h.records) != 0 || len(h.remote.events) != 0 {
		t.Fatalf("closed alerts still acted: %+v", h.records)
	}
}

func TestStaleTimerAfterRaisingTheDelayDoesNotFireEarly(t *testing.T) {
	h2 := newHarness(t, deferConfig(30))
	h2.a.OnEvent(h2.event("disconnected"))
	stale := h2.clock.timers[0]
	h2.clock.now = h2.clock.now.Add(30 * time.Second)
	h2.a.ApplyConfig(deferConfig(300))
	stale.fn()
	if len(h2.records) != 0 {
		t.Fatalf("stale timer announced early: %v", h2.kinds())
	}
	h2.clock.advance(270 * time.Second)
	if got := h2.kinds(); !reflect.DeepEqual(got, []Kind{KindDown}) {
		t.Fatalf("kinds = %v", got)
	}
}
