package alerts

import (
	"fmt"
	"internet-monitor/internal/config"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/types"
	"sync"
	"sync/atomic"
	"time"
)

type Toast struct {
	Kind   Kind
	Status types.Status
	Lang   string
	Title  string
	Body   string
	Time   time.Time
	Volume int
}

type Desktop interface {
	Show(t Toast)
}

type Remote interface {
	SendEvent(lang string, ev types.Event, kind string, webhook, telegram bool)
	SendSpeed(lang string, ev logger.SpeedTestEvent, thresholdMbps float64, below, webhook, telegram bool)
}

type SpeedInfo struct {
	DownloadMbps  float64  `json:"download_mbps"`
	UploadMbps    *float64 `json:"upload_mbps,omitempty"`
	LatencyMs     int64    `json:"latency_ms"`
	ThresholdMbps float64  `json:"threshold_mbps,omitempty"`
	Below         bool     `json:"below,omitempty"`
}

type Record struct {
	ID              string             `json:"id"`
	Time            time.Time          `json:"time"`
	Kind            Kind               `json:"kind"`
	Reason          *types.EventReason `json:"reason,omitempty"`
	DurationSeconds float64            `json:"duration_seconds,omitempty"`
	Speed           *SpeedInfo         `json:"speed,omitempty"`
	Sent            []Channel          `json:"sent"`
	Muted           map[Channel]Reason `json:"muted,omitempty"`
}

type stopper interface{ Stop() bool }

type Options struct {
	Desktop Desktop
	Remote  Remote
	History *History
	Publish func(Record)
	Logf    func(format string, args ...interface{})
}

type pending struct {
	kind  Kind
	event types.Event
	timer stopper
}

type Alerts struct {
	opts Options

	mu           sync.Mutex
	settings     config.Notifications
	has          Configured
	lang         string
	pending      *pending
	announced    Kind
	problemSince time.Time
	lastProblem  types.Event
	reminder     stopper
	closed       bool

	queue chan func()
	done  chan struct{}
	seq   atomic.Uint64

	now   func() time.Time
	after func(time.Duration, func()) stopper
}

const queueSize = 64

func New(cfg config.Config, opts Options) *Alerts {
	a := &Alerts{
		opts:  opts,
		queue: make(chan func(), queueSize),
		done:  make(chan struct{}),
		now:   time.Now,
		after: func(d time.Duration, fn func()) stopper { return time.AfterFunc(d, fn) },
	}
	a.applyLocked(cfg)
	go a.worker()
	return a
}

func (a *Alerts) logf(format string, args ...interface{}) {
	if a.opts.Logf != nil {
		a.opts.Logf(format, args...)
	}
}

func (a *Alerts) worker() {
	defer close(a.done)
	for job := range a.queue {
		a.run(job)
	}
}

func (a *Alerts) run(job func()) {
	defer func() {
		if r := recover(); r != nil {
			a.logf("PANIC recovered in alerts: %v", r)
		}
	}()
	job()
}

func (a *Alerts) enqueue(job func()) {
	if a.closed {
		return
	}
	select {
	case a.queue <- job:
	default:
		a.logf("ALERTS queue full; desktop notification dropped")
	}
}

func (a *Alerts) Close() {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return
	}
	a.stopTimersLocked()
	a.closed = true
	close(a.queue)
	a.mu.Unlock()
	select {
	case <-a.done:
	case <-time.After(3 * time.Second):
	}
}

func (a *Alerts) stopTimersLocked() {
	if a.pending != nil {
		a.pending.timer.Stop()
		a.pending = nil
	}
	if a.reminder != nil {
		a.reminder.Stop()
		a.reminder = nil
	}
}

func (a *Alerts) Settings() config.Notifications {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.Clone()
}

func (a *Alerts) ApplyConfig(cfg config.Config) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	prevReminder := a.settings.ReminderMinutes
	a.applyLocked(cfg)
	if p := a.pending; p != nil {
		wait := time.Duration(a.settings.MinOutageSeconds)*time.Second - a.now().Sub(a.problemSince)
		p.timer.Stop()
		if wait <= 0 {
			a.firePendingLocked(p)
		} else {
			p.timer = a.after(wait, func() { a.fire(p) })
		}
	}
	if a.announced == KindDown && a.settings.ReminderMinutes != prevReminder {
		a.armReminderLocked()
	}
}

func (a *Alerts) applyLocked(cfg config.Config) {
	a.settings = cfg.Notifications.Clone()
	a.lang = cfg.Language
	a.has = Configured{
		Webhook:  cfg.WebhookURL != "" && config.IsSupportedWebhook(cfg.WebhookURL),
		Telegram: cfg.TelegramBotToken != "" && cfg.TelegramChatID != "",
	}
}

func kindOf(eventType string) Kind {
	switch eventType {
	case "connected":
		return KindRestored
	case "degraded":
		return KindDegraded
	default:
		return KindDown
	}
}

func (a *Alerts) OnTick(types.CheckResult, types.Status) {}

func (a *Alerts) OnInitialEvent(ev types.Event) {
	kind := kindOf(ev.EventType)
	if kind == KindRestored {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	a.problemSince = ev.Timestamp
	a.lastProblem = ev
	a.announced = kind
	d := Decide(kind, a.settings, a.has, a.now())
	if d.Desktop {
		d.Desktop, d.Volume = false, 0
		delete(d.Muted, ChannelSound)
		d.mute(ChannelDesktop, ReasonStartup)
	}
	a.dispatchLocked(kind, ev, d, 0)
	if kind == KindDown {
		a.armReminderLocked()
	}
}

func (a *Alerts) OnEvent(ev types.Event) {
	kind := kindOf(ev.EventType)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	if a.reminder != nil {
		a.reminder.Stop()
		a.reminder = nil
	}

	if kind == KindRestored {
		outage := ev.DurationSeconds
		if !a.problemSince.IsZero() {
			outage = ev.Timestamp.Sub(a.problemSince).Seconds()
		}
		p, announced := a.pending, a.announced
		a.pending, a.announced, a.problemSince = nil, "", time.Time{}
		if p != nil {
			p.timer.Stop()
			a.recordShortLocked(p, outage)
			if announced == "" {
				return
			}
		}
		a.dispatchLocked(kind, ev, Decide(kind, a.settings, a.has, a.now()), outage)
		return
	}

	if a.problemSince.IsZero() {
		a.problemSince = ev.Timestamp
	}
	a.lastProblem = ev
	if a.announced == "" && a.settings.MinOutageSeconds > 0 {
		if a.pending != nil {
			a.pending.kind, a.pending.event = kind, ev
			return
		}
		p := &pending{kind: kind, event: ev}
		wait := time.Duration(a.settings.MinOutageSeconds)*time.Second - a.now().Sub(a.problemSince)
		if wait < 0 {
			wait = 0
		}
		p.timer = a.after(wait, func() { a.fire(p) })
		a.pending = p
		return
	}
	a.announceLocked(kind, ev)
}

func (a *Alerts) fire(p *pending) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed || a.pending != p {
		return
	}
	if a.now().Sub(a.problemSince) < time.Duration(a.settings.MinOutageSeconds)*time.Second {
		return
	}
	a.firePendingLocked(p)
}

func (a *Alerts) firePendingLocked(p *pending) {
	a.pending = nil
	a.announceLocked(p.kind, p.event)
}

func (a *Alerts) announceLocked(kind Kind, ev types.Event) {
	a.announced = kind
	a.dispatchLocked(kind, ev, Decide(kind, a.settings, a.has, a.now()), 0)
	if kind == KindDown {
		a.armReminderLocked()
	}
}

func (a *Alerts) armReminderLocked() {
	if a.reminder != nil {
		a.reminder.Stop()
		a.reminder = nil
	}
	if a.settings.ReminderMinutes <= 0 {
		return
	}
	r := &reminderTimer{}
	r.t = a.after(time.Duration(a.settings.ReminderMinutes)*time.Minute, func() { a.remind(r) })
	a.reminder = r
}

type reminderTimer struct{ t stopper }

func (r *reminderTimer) Stop() bool { return r.t.Stop() }

func (a *Alerts) remind(r *reminderTimer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed || a.reminder != r || a.announced != KindDown {
		return
	}
	a.reminder = nil
	elapsed := a.now().Sub(a.problemSince).Seconds()
	a.dispatchLocked(KindReminder, a.lastProblem, Decide(KindReminder, a.settings, a.has, a.now()), elapsed)
	a.armReminderLocked()
}

func (a *Alerts) recordShortLocked(p *pending, outage float64) {
	reason := p.event.Reason
	rec := Record{
		ID:              a.nextID(),
		Time:            a.now(),
		Kind:            p.kind,
		Reason:          &reason,
		DurationSeconds: outage,
		Sent:            []Channel{},
		Muted:           map[Channel]Reason{ChannelDesktop: ReasonShortOutage},
	}
	if a.has.Webhook {
		rec.Muted[ChannelWebhook] = ReasonShortOutage
	}
	if a.has.Telegram {
		rec.Muted[ChannelTelegram] = ReasonShortOutage
	}
	a.record(rec)
}

func (a *Alerts) dispatchLocked(kind Kind, ev types.Event, d Decision, duration float64) {
	now := a.now()
	reason := ev.Reason
	a.record(Record{
		ID:              a.nextID(),
		Time:            now,
		Kind:            kind,
		Reason:          &reason,
		DurationSeconds: duration,
		Sent:            d.Sent(),
		Muted:           d.Muted,
	})

	lang := a.lang
	if d.Desktop && a.opts.Desktop != nil {
		title, body := EventText(lang, kind, ev, duration)
		a.show(Toast{Kind: kind, Status: statusOf(kind), Lang: lang, Title: title, Body: body, Time: now, Volume: d.Volume})
	}
	if (d.Webhook || d.Telegram) && a.opts.Remote != nil {
		sent := ev
		sent.DurationSeconds = duration
		a.opts.Remote.SendEvent(lang, sent, string(kind), d.Webhook, d.Telegram)
	}
}

func (a *Alerts) show(t Toast) {
	desktop := a.opts.Desktop
	a.enqueue(func() { desktop.Show(t) })
}

func (a *Alerts) record(rec Record) {
	if a.opts.History != nil {
		a.opts.History.Add(rec)
	}
	if a.opts.Publish != nil {
		a.opts.Publish(rec)
	}
}

func (a *Alerts) nextID() string {
	return fmt.Sprintf("%d-%d", a.now().UnixNano(), a.seq.Add(1))
}

func statusOf(kind Kind) types.Status {
	switch kind {
	case KindRestored, KindTest:
		return types.StatusConnected
	case KindDegraded, KindSpeed:
		return types.StatusDegraded
	default:
		return types.StatusDisconnected
	}
}

func (a *Alerts) OnSpeedResult(ev logger.SpeedTestEvent, thresholdMbps float64, below bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	now := a.now()
	d := Decide(KindSpeed, a.settings, a.has, now)
	info := SpeedInfo{DownloadMbps: ev.DownloadMbps, UploadMbps: ev.UploadMbps, LatencyMs: ev.LatencyMs, ThresholdMbps: thresholdMbps, Below: below}
	a.record(Record{ID: a.nextID(), Time: now, Kind: KindSpeed, Speed: &info, Sent: d.Sent(), Muted: d.Muted})

	lang := a.lang
	if d.Desktop && a.opts.Desktop != nil {
		title, body := SpeedText(lang, info)
		status := types.StatusConnected
		if below {
			status = types.StatusDegraded
		}
		a.show(Toast{Kind: KindSpeed, Status: status, Lang: lang, Title: title, Body: body, Time: now, Volume: d.Volume})
	}
	if (d.Webhook || d.Telegram) && a.opts.Remote != nil {
		a.opts.Remote.SendSpeed(lang, ev, thresholdMbps, below, d.Webhook, d.Telegram)
	}
}

func (a *Alerts) Test(lang string, volume *int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	now := a.now()
	settings := a.settings
	if volume != nil {
		settings.Sound = config.SoundSettings{Enabled: *volume > 0, Volume: *volume}
	}
	d := Decide(KindTest, settings, a.has, now)
	a.record(Record{ID: a.nextID(), Time: now, Kind: KindTest, Sent: d.Sent(), Muted: d.Muted})
	if a.opts.Desktop != nil {
		title, body := TestText(lang)
		a.show(Toast{Kind: KindTest, Status: types.StatusConnected, Lang: lang, Title: title, Body: body, Time: now, Volume: d.Volume})
	}
}
