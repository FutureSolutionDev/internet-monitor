package config

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"
)

type SpeedTestConfig struct {
	DownloadTargets     []string `json:"download_targets"`
	ParallelConnections int      `json:"parallel_connections"`
	TimeoutSeconds      int      `json:"timeout_seconds"`
	UploadTarget        string   `json:"upload_target"`
	AlertThresholdMbps  float64  `json:"alert_threshold_mbps"`
	ScheduleMinutes     int      `json:"schedule_minutes"`
}

type Config struct {
	CheckIntervalSec    int             `json:"check_interval_sec"`
	PingTargets         []string        `json:"ping_targets"`
	HTTPTargets         []string        `json:"http_targets"`
	DNSTargets          []string        `json:"dns_targets"`
	FailThreshold       int             `json:"fail_threshold"`
	PacketLossThreshold float64         `json:"packet_loss_threshold"`
	LatencyThreshold    int             `json:"latency_threshold_ms"`
	LogDir              string          `json:"log_dir"`
	WebhookURL          string          `json:"webhook_url"`
	TelegramBotToken    string          `json:"telegram_bot_token"`
	TelegramChatID      string          `json:"telegram_chat_id"`
	DashboardPort       int             `json:"dashboard_port"`
	UseICMP             bool            `json:"use_icmp"`
	Language            string          `json:"language"`
	SpeedTest           SpeedTestConfig `json:"speed_test"`
}

const (
	MaxTargetsPerLayer = 10
	MaxTargetLength    = 512
)

var ErrInvalid = errors.New("invalid configuration")

func Default() Config {
	return Config{
		CheckIntervalSec:    5,
		PingTargets:         []string{"8.8.8.8:53", "1.1.1.1:53"},
		HTTPTargets:         []string{"https://connectivitycheck.gstatic.com/generate_204"},
		DNSTargets:          []string{"www.google.com", "www.cloudflare.com"},
		FailThreshold:       3,
		PacketLossThreshold: 20.0,
		LatencyThreshold:    500,
		LogDir:              "logs",
		DashboardPort:       8765,
		Language:            "en",
		SpeedTest: SpeedTestConfig{
			DownloadTargets:     []string{"https://speed.cloudflare.com/__down"},
			ParallelConnections: 4,
			TimeoutSeconds:      30,
			UploadTarget:        "https://speed.cloudflare.com/__up",
		},
	}
}

func (c Config) Clone() Config {
	out := c
	out.PingTargets = cloneStrings(c.PingTargets)
	out.HTTPTargets = cloneStrings(c.HTTPTargets)
	out.DNSTargets = cloneStrings(c.DNSTargets)
	out.SpeedTest.DownloadTargets = cloneStrings(c.SpeedTest.DownloadTargets)
	return out
}

func cloneStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func Parse(data []byte) (Config, error) {
	cfg := Default()
	cfg.PingTargets, cfg.HTTPTargets, cfg.DNSTargets = nil, nil, nil
	cfg.SpeedTest.DownloadTargets = nil
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}

	var legacy struct {
		HTTPTarget string `json:"http_target"`
		DNSTarget  string `json:"dns_target"`
	}
	if json.Unmarshal(data, &legacy) == nil {
		if len(cfg.HTTPTargets) == 0 && legacy.HTTPTarget != "" {
			cfg.HTTPTargets = []string{legacy.HTTPTarget}
		}
		if len(cfg.DNSTargets) == 0 && legacy.DNSTarget != "" {
			cfg.DNSTargets = []string{legacy.DNSTarget}
		}
	}

	def := Default()
	if cfg.PingTargets == nil {
		cfg.PingTargets = def.PingTargets
	}
	if cfg.HTTPTargets == nil {
		cfg.HTTPTargets = def.HTTPTargets
	}
	if cfg.DNSTargets == nil {
		cfg.DNSTargets = def.DNSTargets
	}
	cfg.Sanitize()
	return cfg, nil
}

func (c *Config) Sanitize() {
	def := Default()
	if c.CheckIntervalSec < 1 {
		c.CheckIntervalSec = def.CheckIntervalSec
	}
	if c.CheckIntervalSec > 3600 {
		c.CheckIntervalSec = 3600
	}
	if c.FailThreshold < 1 {
		c.FailThreshold = def.FailThreshold
	}
	if c.PacketLossThreshold < 0 || c.PacketLossThreshold > 100 {
		c.PacketLossThreshold = def.PacketLossThreshold
	}
	if c.LatencyThreshold < 0 {
		c.LatencyThreshold = def.LatencyThreshold
	}
	if c.DashboardPort < 1024 || c.DashboardPort > 65535 {
		c.DashboardPort = def.DashboardPort
	}
	if strings.TrimSpace(c.LogDir) == "" {
		c.LogDir = def.LogDir
	}
	if c.Language != "ar" && c.Language != "en" {
		c.Language = "en"
	}
	c.PingTargets = cleanTargets(c.PingTargets)
	c.HTTPTargets = cleanTargets(c.HTTPTargets)
	c.DNSTargets = cleanTargets(c.DNSTargets)
	c.SpeedTest.DownloadTargets = cleanTargets(c.SpeedTest.DownloadTargets)
	if c.SpeedTest.ParallelConnections < 1 {
		c.SpeedTest.ParallelConnections = def.SpeedTest.ParallelConnections
	}
	if c.SpeedTest.ParallelConnections > 16 {
		c.SpeedTest.ParallelConnections = 16
	}
	if c.SpeedTest.TimeoutSeconds < 5 {
		c.SpeedTest.TimeoutSeconds = def.SpeedTest.TimeoutSeconds
	}
	if c.SpeedTest.TimeoutSeconds > 120 {
		c.SpeedTest.TimeoutSeconds = 120
	}
	if len(c.SpeedTest.DownloadTargets) == 0 {
		c.SpeedTest.DownloadTargets = def.SpeedTest.DownloadTargets
	}
	if c.SpeedTest.AlertThresholdMbps < 0 {
		c.SpeedTest.AlertThresholdMbps = 0
	}
	if c.SpeedTest.ScheduleMinutes < 0 {
		c.SpeedTest.ScheduleMinutes = 0
	}
}

func cleanTargets(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]bool, len(in))
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t == "" || len(t) > MaxTargetLength || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
		if len(out) == MaxTargetsPerLayer {
			break
		}
	}
	return out
}

func (c Config) Validate() error {
	dir := filepath.Clean(c.LogDir)
	if filepath.IsAbs(c.LogDir) || filepath.VolumeName(c.LogDir) != "" || strings.HasPrefix(c.LogDir, "/") || strings.HasPrefix(c.LogDir, `\`) {
		return fieldError("log_dir", "must be a folder inside the app directory")
	}
	if dir == ".." || strings.HasPrefix(dir, ".."+string(filepath.Separator)) || strings.Contains(c.LogDir, "..") {
		return fieldError("log_dir", "must not leave the app directory")
	}
	if len(c.PingTargets) == 0 && len(c.HTTPTargets) == 0 && len(c.DNSTargets) == 0 {
		return fieldError("targets", "at least one target is required")
	}
	return nil
}

type FieldError struct {
	Field   string
	Message string
}

func (e *FieldError) Error() string { return e.Field + ": " + e.Message }

func (e *FieldError) Unwrap() error { return ErrInvalid }

func fieldError(field, msg string) error { return &FieldError{Field: field, Message: msg} }

func (c *Config) CheckInterval() time.Duration {
	if c.CheckIntervalSec < 1 {
		return time.Second
	}
	return time.Duration(c.CheckIntervalSec) * time.Second
}
