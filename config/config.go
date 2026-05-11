package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	CheckIntervalSec    int      `json:"check_interval_sec"`
	PingTargets         []string `json:"ping_targets"`
	HTTPTarget          string   `json:"http_target"`
	DNSTarget           string   `json:"dns_target"`
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
	HTTPTarget:          "https://connectivitycheck.gstatic.com/generate_204",
	DNSTarget:           "www.google.com",
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
		return &Default, nil
	}
	if err != nil {
		return nil, err
	}
	cfg := Default
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) CheckInterval() time.Duration {
	return time.Duration(c.CheckIntervalSec) * time.Second
}
