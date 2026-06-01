package config

import "testing"

func TestSanitizeClampsInvalidValues(t *testing.T) {
	c := Config{
		CheckIntervalSec:    0, // would panic time.NewTicker
		FailThreshold:       -1,
		PacketLossThreshold: -5,
		LatencyThreshold:    -10,
		DashboardPort:       0,
		LogDir:              "",
	}
	c.Sanitize()

	if c.CheckIntervalSec != Default.CheckIntervalSec {
		t.Errorf("CheckIntervalSec = %d, want %d", c.CheckIntervalSec, Default.CheckIntervalSec)
	}
	if c.FailThreshold != Default.FailThreshold {
		t.Errorf("FailThreshold = %d, want %d", c.FailThreshold, Default.FailThreshold)
	}
	if c.DashboardPort != Default.DashboardPort {
		t.Errorf("DashboardPort = %d, want %d", c.DashboardPort, Default.DashboardPort)
	}
	if c.LogDir != Default.LogDir {
		t.Errorf("LogDir = %q, want %q", c.LogDir, Default.LogDir)
	}
	if c.SpeedTest.ParallelConnections < 1 || c.SpeedTest.TimeoutSeconds < 1 {
		t.Errorf("speed test defaults not applied: %+v", c.SpeedTest)
	}
}

func TestSanitizeKeepsValidValues(t *testing.T) {
	c := Config{
		CheckIntervalSec: 30,
		FailThreshold:    5,
		DashboardPort:    9000,
		LogDir:           "custom-logs",
	}
	c.Sanitize()
	if c.CheckIntervalSec != 30 || c.FailThreshold != 5 || c.DashboardPort != 9000 || c.LogDir != "custom-logs" {
		t.Errorf("Sanitize mutated valid values: %+v", c)
	}
}

func TestCheckIntervalNeverZero(t *testing.T) {
	c := Config{CheckIntervalSec: 0}
	if got := c.CheckInterval(); got <= 0 {
		t.Errorf("CheckInterval() = %v, want > 0 (time.NewTicker panics on <= 0)", got)
	}
}
