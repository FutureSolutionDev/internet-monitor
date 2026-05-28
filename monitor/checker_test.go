package monitor

import (
	"internet-monitor/config"
	"net"
	"testing"
)

func TestProbePreservesOrderAndResults(t *testing.T) {
	got := probe("tcp", []string{"a", "b", "c"}, func(target string) (bool, int64) {
		return target == "b", 7
	})
	if len(got) != 3 {
		t.Fatalf("got %d results, want 3", len(got))
	}
	if got[0].Target != "a" || got[1].Target != "b" || got[2].Target != "c" {
		t.Errorf("order not preserved: %+v", got)
	}
	if got[0].OK || !got[1].OK || got[2].OK {
		t.Errorf("ok flags wrong: %+v", got)
	}
	if got[1].LatencyMs != 7 || got[1].Layer != "tcp" {
		t.Errorf("layer/latency wrong: %+v", got[1])
	}
}

func TestCheckTCPPerTarget(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	open := ln.Addr().String()

	c := NewChecker(&config.Config{PingTargets: []string{open, "127.0.0.1:1"}})
	r := c.Check()

	if !r.TCPPingOK {
		t.Fatal("expected TCPPingOK (one target is open)")
	}
	var openOK, closedOK *bool
	for i := range r.Targets {
		tr := r.Targets[i]
		if tr.Layer != "tcp" {
			continue
		}
		ok := tr.OK
		if tr.Target == open {
			openOK = &ok
		}
		if tr.Target == "127.0.0.1:1" {
			closedOK = &ok
		}
	}
	if openOK == nil || !*openOK {
		t.Errorf("open target should be OK; targets=%+v", r.Targets)
	}
	if closedOK == nil || *closedOK {
		t.Errorf("closed target should not be OK; targets=%+v", r.Targets)
	}
}
