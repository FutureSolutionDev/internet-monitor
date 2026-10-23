package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestSanitizeClampsInvalidValues(t *testing.T) {
	c := Config{DashboardPort: 0, LogDir: "", CheckIntervalSec: 0, FailThreshold: -1, PacketLossThreshold: -5, LatencyThreshold: -10}
	c.Sanitize()
	def := Default()
	if c.CheckIntervalSec != def.CheckIntervalSec || c.FailThreshold != def.FailThreshold || c.DashboardPort != def.DashboardPort || c.LogDir != def.LogDir {
		t.Errorf("sanitize did not restore defaults: %+v", c)
	}
	if c.SpeedTest.ParallelConnections < 1 || c.SpeedTest.TimeoutSeconds < 5 {
		t.Errorf("speed test defaults not applied: %+v", c.SpeedTest)
	}
}

func TestSanitizeKeepsZeroLossThreshold(t *testing.T) {
	c := Default()
	c.PacketLossThreshold = 0
	c.LatencyThreshold = 0
	c.Sanitize()
	if c.PacketLossThreshold != 0 || c.LatencyThreshold != 0 {
		t.Errorf("explicit zero thresholds were overwritten: loss=%v latency=%v", c.PacketLossThreshold, c.LatencyThreshold)
	}
}

func TestSanitizeDedupesAndCapsTargets(t *testing.T) {
	c := Default()
	c.PingTargets = []string{" a:1 ", "a:1", "", "b:2"}
	for i := 0; i < 20; i++ {
		c.DNSTargets = append(c.DNSTargets, strings.Repeat("x", i+1))
	}
	c.Sanitize()
	if len(c.PingTargets) != 2 || c.PingTargets[0] != "a:1" {
		t.Errorf("ping targets = %q", c.PingTargets)
	}
	if len(c.DNSTargets) != MaxTargetsPerLayer {
		t.Errorf("dns targets len = %d, want %d", len(c.DNSTargets), MaxTargetsPerLayer)
	}
}

func TestParseDoesNotMutateDefaults(t *testing.T) {
	if _, err := Parse([]byte(`{"ping_targets":["9.9.9.9:53"],"speed_test":{"download_targets":["https://x.test"]}}`)); err != nil {
		t.Fatal(err)
	}
	def := Default()
	if def.PingTargets[0] != "8.8.8.8:53" || def.SpeedTest.DownloadTargets[0] != "https://speed.cloudflare.com/__down" {
		t.Fatalf("defaults were mutated by Parse: %+v", def)
	}
}

func TestParseLegacySingleTargets(t *testing.T) {
	cfg, err := Parse([]byte(`{"http_target":"https://legacy.test","dns_target":"legacy.test","http_targets":[],"dns_targets":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.HTTPTargets) != 1 || cfg.HTTPTargets[0] != "https://legacy.test" || cfg.DNSTargets[0] != "legacy.test" {
		t.Errorf("legacy targets not migrated: http=%q dns=%q", cfg.HTTPTargets, cfg.DNSTargets)
	}
}

func TestCloneIsDeep(t *testing.T) {
	a := Default()
	b := a.Clone()
	b.PingTargets[0] = "changed"
	b.SpeedTest.DownloadTargets[0] = "changed"
	if a.PingTargets[0] == "changed" || a.SpeedTest.DownloadTargets[0] == "changed" {
		t.Error("Clone shares slice backing arrays")
	}
}

func TestValidateRejectsEscapingLogDir(t *testing.T) {
	for _, dir := range []string{`C:\Windows`, "/etc", "../outside", `logs\..\..\x`, `\\server\share`} {
		c := Default()
		c.LogDir = dir
		err := c.Validate()
		var fe *FieldError
		if !errors.As(err, &fe) || fe.Field != "log_dir" {
			t.Errorf("log_dir %q: err = %v, want log_dir field error", dir, err)
		}
	}
	c := Default()
	c.LogDir = "data/logs"
	if err := c.Validate(); err != nil {
		t.Errorf("nested relative log_dir rejected: %v", err)
	}
}

func TestCheckIntervalNeverZero(t *testing.T) {
	c := Config{CheckIntervalSec: 0}
	if got := c.CheckInterval(); got <= 0 {
		t.Errorf("CheckInterval() = %v, want > 0", got)
	}
}

func TestStoreCreatesDefaultFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("default config not written: %v", err)
	}
	if s.Revision() != 1 {
		t.Errorf("revision = %d, want 1", s.Revision())
	}
}

func TestStoreBacksUpInvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.LoadError() == nil {
		t.Fatal("expected LoadError for invalid config")
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "config.json.invalid-*"))
	if len(matches) != 1 {
		t.Fatalf("expected one backup, got %v", matches)
	}
	if data, _ := os.ReadFile(path); string(data) != "{not json" {
		t.Error("invalid user file was overwritten on load")
	}
}

func TestStoreUpdateRevisionAndConflict(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var notified Config
	s.Subscribe(func(c Config) { notified = c })

	_, rev, err := s.Update(1, func(c *Config) error { c.CheckIntervalSec = 9; return nil })
	if err != nil || rev != 2 {
		t.Fatalf("update: rev=%d err=%v", rev, err)
	}
	if notified.CheckIntervalSec != 9 {
		t.Error("subscriber not notified with new config")
	}
	if _, _, err := s.Update(1, func(c *Config) error { c.CheckIntervalSec = 10; return nil }); !errors.Is(err, ErrConflict) {
		t.Errorf("stale revision: err = %v, want ErrConflict", err)
	}
	reloaded, err := Open(s.Path())
	if err != nil || reloaded.Get().CheckIntervalSec != 9 {
		t.Errorf("persisted value = %d, err=%v", reloaded.Get().CheckIntervalSec, err)
	}
}

func TestStoreRejectsInvalidUpdateWithoutWriting(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Update(0, func(c *Config) error { c.LogDir = "../../x"; return nil }); !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	if s.Get().LogDir != "logs" || s.Revision() != 1 {
		t.Errorf("invalid update leaked: %+v rev=%d", s.Get().LogDir, s.Revision())
	}
}

func TestStoreConcurrentUpdatesAreSerialized(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = s.Update(0, func(c *Config) error { c.FailThreshold++; return nil })
			_ = s.Get()
		}()
	}
	wg.Wait()
	if got := s.Get().FailThreshold; got != 3+20 {
		t.Errorf("FailThreshold = %d, want 23", got)
	}
	if s.Revision() != 21 {
		t.Errorf("revision = %d, want 21", s.Revision())
	}
}

func TestStoreKeepsLegacyAbsoluteLogDirWritable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"log_dir":"D:\monitor-logs"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Update(0, func(c *Config) error { c.Language = "ar"; return nil }); err != nil {
		t.Fatalf("unrelated save blocked by legacy log_dir: %v", err)
	}
	if _, _, err := s.Update(0, func(c *Config) error { c.LogDir = `E:\monitor`; return nil }); !errors.Is(err, ErrInvalid) {
		t.Errorf("changing to a new absolute log_dir: err = %v, want ErrInvalid", err)
	}
}

func TestWebhookValidation(t *testing.T) {
	good := []string{
		"https://discord.com/api/webhooks/1/abc",
		"https://hooks.slack.com/services/T/B/x",
	}
	bad := []string{
		"https://evil.example/?x=discord.com/api/webhooks",
		"http://discord.com/api/webhooks/1/abc",
		"https://discord.com.evil.example/api/webhooks/1",
		"https://user@hooks.slack.com/services/T",
		"https://hooks.slack.com/",
	}
	for _, u := range good {
		if !IsSupportedWebhook(u) {
			t.Errorf("%q rejected", u)
		}
	}
	for _, u := range bad {
		if IsSupportedWebhook(u) {
			t.Errorf("%q accepted", u)
		}
	}
	s, err := Open(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Update(0, func(c *Config) error { c.WebhookURL = bad[0]; return nil }); !errors.Is(err, ErrInvalid) {
		t.Errorf("unsupported webhook saved: err = %v", err)
	}
}

func TestValidateAllowsDotsInsideNames(t *testing.T) {
	c := Default()
	c.LogDir = "my..logs"
	if err := c.Validate(); err != nil {
		t.Errorf("my..logs rejected: %v", err)
	}
}
