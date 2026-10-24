package config

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestOldConfigGetsDefaultNotifications(t *testing.T) {
	cfg, err := Parse([]byte(`{"language":"en"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Notifications, DefaultNotifications()) {
		t.Fatalf("notifications = %+v", cfg.Notifications)
	}
	n := cfg.Notifications
	if !n.Sound.Enabled || n.Sound.Volume != 100 || n.QuietHours.Enabled || n.Events.Speed.Desktop || !n.Events.Down.Webhook {
		t.Fatalf("defaults changed current behaviour: %+v", n)
	}
}

func TestParseMergesPartialNotifications(t *testing.T) {
	cfg, err := Parse([]byte(`{"notifications":{"sound":{"volume":40},"events":{"down":{"telegram":false}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	n := cfg.Notifications
	if !n.Sound.Enabled || n.Sound.Volume != 40 {
		t.Errorf("sound = %+v", n.Sound)
	}
	if !n.Events.Down.Desktop || !n.Events.Down.Webhook || n.Events.Down.Telegram {
		t.Errorf("down = %+v", n.Events.Down)
	}
}

func TestNotificationsSanitize(t *testing.T) {
	n := Notifications{
		Sound:            SoundSettings{Volume: 250},
		QuietHours:       QuietHours{Start: "25:00", End: "7:5", Days: []int{6, 1, 1, 9, -1}, Mode: "loud"},
		MinOutageSeconds: 99999,
		ReminderMinutes:  -3,
	}
	n.Sanitize()
	if n.Sound.Volume != 100 {
		t.Errorf("volume = %d", n.Sound.Volume)
	}
	if n.QuietHours.Start != "22:00" || n.QuietHours.End != "07:00" {
		t.Errorf("clock = %s-%s", n.QuietHours.Start, n.QuietHours.End)
	}
	if !reflect.DeepEqual(n.QuietHours.Days, []int{1, 6}) {
		t.Errorf("days = %v", n.QuietHours.Days)
	}
	if n.QuietHours.Mode != QuietModeSound || n.MinOutageSeconds != MaxMinOutageSeconds || n.ReminderMinutes != 0 {
		t.Errorf("sanitized = %+v", n)
	}
}

func TestNotificationsDaysEmptyStaysEmpty(t *testing.T) {
	cfg, err := Parse([]byte(`{"notifications":{"quiet_hours":{"days":[]}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Notifications.QuietHours.Days == nil || len(cfg.Notifications.QuietHours.Days) != 0 {
		t.Fatalf("days = %#v", cfg.Notifications.QuietHours.Days)
	}
	out, _ := json.Marshal(cfg.Notifications.QuietHours)
	if !json.Valid(out) || !strings.Contains(string(out), `"days":[]`) {
		t.Fatalf("marshal = %s", out)
	}
}

func TestParseClock(t *testing.T) {
	for in, want := range map[string]int{"00:00": 0, "7:05": 425, "23:59": 1439} {
		if got, ok := ParseClock(in); !ok || got != want {
			t.Errorf("ParseClock(%q) = %d,%v", in, got, ok)
		}
	}
	for _, in := range []string{"", "24:00", "12:60", "12", "1:2", "aa:bb", "123:00"} {
		if _, ok := ParseClock(in); ok {
			t.Errorf("ParseClock(%q) accepted", in)
		}
	}
}

func TestStoreUpdateOfDaysKeepsPreviousSnapshot(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	before := s.Get()
	if _, _, err := s.Update(0, func(c *Config) error {
		return json.Unmarshal([]byte(`{"notifications":{"quiet_hours":{"days":[2]}}}`), c)
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.Notifications.QuietHours.Days, allDays()) {
		t.Fatalf("previous snapshot mutated: %v", before.Notifications.QuietHours.Days)
	}
	if got := s.Get().Notifications.QuietHours.Days; !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("days = %v", got)
	}
}
