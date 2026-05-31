package config

import (
	"encoding/json"
	"os"
	"time"
)

// SpeedTestConfig holds speed test parameters (all optional; defaults applied on load).
type SpeedTestConfig struct {
	DownloadTargets     []string `json:"download_targets"`
	ParallelConnections int      `json:"parallel_connections"`
	TimeoutSeconds      int      `json:"timeout_seconds"`
	UploadTarget        string   `json:"upload_target"` // POST endpoint for the upload test; empty disables it
	AlertThresholdMbps  float64  `json:"alert_threshold_mbps"`
	ScheduleMinutes     int      `json:"schedule_minutes"` // 0 = disabled; otherwise run an automatic speed test every N minutes
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
	UseICMP             bool            `json:"use_icmp"` // use ICMP echo for ping (falls back to TCP)
	SpeedTest           SpeedTestConfig `json:"speed_test,omitempty"`
}

var Default = Config{
	CheckIntervalSec:    5,
	PingTargets:         []string{"8.8.8.8:53", "1.1.1.1:53"},
	HTTPTargets:         []string{"https://connectivitycheck.gstatic.com/generate_204"},
	DNSTargets:          []string{"www.google.com", "www.cloudflare.com"},
	FailThreshold:       3,
	PacketLossThreshold: 20.0,
	LatencyThreshold:    500,
	LogDir:              "logs",
	WebhookURL:          "",
	DashboardPort:       8765,
	SpeedTest: SpeedTestConfig{
		DownloadTargets:     []string{"https://speed.cloudflare.com/__down"},
		ParallelConnections: 4,
		TimeoutSeconds:      10,
		UploadTarget:        "https://speed.cloudflare.com/__up",
	},
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default
		writeDefault(path, cfg)
		return &cfg, nil
	}
	if err != nil {
		return nil, err
	}

	cfg := Default
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Migrate from old single-value fields (http_target, dns_target → http_targets, dns_targets)
	var old struct {
		HTTPTarget string `json:"http_target"`
		DNSTarget  string `json:"dns_target"`
	}
	if json.Unmarshal(data, &old) == nil {
		if len(cfg.HTTPTargets) == 0 && old.HTTPTarget != "" {
			cfg.HTTPTargets = []string{old.HTTPTarget}
		}
		if len(cfg.DNSTargets) == 0 && old.DNSTarget != "" {
			cfg.DNSTargets = []string{old.DNSTarget}
		}
	}

	// Apply speed test defaults for existing configs without speed_test key
	if len(cfg.SpeedTest.DownloadTargets) == 0 {
		cfg.SpeedTest.DownloadTargets = Default.SpeedTest.DownloadTargets
	}
	if cfg.SpeedTest.ParallelConnections == 0 {
		cfg.SpeedTest.ParallelConnections = Default.SpeedTest.ParallelConnections
	}
	if cfg.SpeedTest.TimeoutSeconds == 0 {
		cfg.SpeedTest.TimeoutSeconds = Default.SpeedTest.TimeoutSeconds
	}

	cfg.Sanitize()
	return &cfg, nil
}

// Sanitize clamps out-of-range values to safe defaults so a malformed config
// (e.g. check_interval_sec=0, which would panic time.NewTicker) can never crash
// the process or be persisted from the dashboard.
func (c *Config) Sanitize() {
	if c.CheckIntervalSec < 1 {
		c.CheckIntervalSec = Default.CheckIntervalSec
	}
	if c.FailThreshold < 1 {
		c.FailThreshold = Default.FailThreshold
	}
	if c.PacketLossThreshold < 0 {
		c.PacketLossThreshold = Default.PacketLossThreshold
	}
	if c.LatencyThreshold < 0 {
		c.LatencyThreshold = Default.LatencyThreshold
	}
	if c.DashboardPort < 1 || c.DashboardPort > 65535 {
		c.DashboardPort = Default.DashboardPort
	}
	if c.LogDir == "" {
		c.LogDir = Default.LogDir
	}
	if c.SpeedTest.ParallelConnections < 1 {
		c.SpeedTest.ParallelConnections = Default.SpeedTest.ParallelConnections
	}
	if c.SpeedTest.TimeoutSeconds < 1 {
		c.SpeedTest.TimeoutSeconds = Default.SpeedTest.TimeoutSeconds
	}
	if len(c.SpeedTest.DownloadTargets) == 0 {
		c.SpeedTest.DownloadTargets = Default.SpeedTest.DownloadTargets
	}
	if c.SpeedTest.UploadTarget == "" {
		c.SpeedTest.UploadTarget = Default.SpeedTest.UploadTarget
	}
}

func writeDefault(path string, cfg Config) {
	pretty, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, pretty, 0644)
}

func (c *Config) CheckInterval() time.Duration {
	if c.CheckIntervalSec < 1 {
		return time.Second
	}
	return time.Duration(c.CheckIntervalSec) * time.Second
}
