package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	CheckIntervalSec    int      `json:"check_interval_sec"`
	PingTargets         []string `json:"ping_targets"`
	HTTPTargets         []string `json:"http_targets"`
	DNSTargets          []string `json:"dns_targets"`
	FailThreshold       int      `json:"fail_threshold"`
	PacketLossThreshold float64  `json:"packet_loss_threshold"`
	LatencyThreshold    int      `json:"latency_threshold_ms"`
	LogDir              string   `json:"log_dir"`
	WebhookURL          string   `json:"webhook_url"`
	DashboardPort       int      `json:"dashboard_port"`
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

	return &cfg, nil
}

func writeDefault(path string, cfg Config) {
	pretty, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, pretty, 0644)
}

func (c *Config) CheckInterval() time.Duration {
	return time.Duration(c.CheckIntervalSec) * time.Second
}
