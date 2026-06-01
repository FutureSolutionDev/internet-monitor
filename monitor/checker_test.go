package monitor

import (
	"internet-monitor/config"
	"net"
	"strings"
	"testing"
	"time"
)

func TestICMPPingLoopback(t *testing.T) {
	lat, ok := icmpPing("127.0.0.1", time.Second)
	if !ok {
		t.Skip("ICMP unavailable here (expected without privileges/ping_group_range)")
	}
	if lat < 0 {
		t.Errorf("negative latency %d", lat)
	}
}

func TestParseProcRoute(t *testing.T) {
	// Gateway 0100A8C0 is little-endian for 192.168.0.1.
	data := "Iface\tDestination\tGateway\tFlags\tRefCnt\tUse\tMetric\tMask\n" +
		"eth0\t00010A0A\t00000000\t0001\t0\t0\t0\t00FFFFFF\n" + // not default
		"eth0\t00000000\t0100A8C0\t0003\t0\t0\t0\t00000000\n" // default + RTF_GATEWAY
	gw, ok := parseProcRoute(strings.NewReader(data))
	if !ok || gw != "192.168.0.1" {
		t.Fatalf("parseProcRoute = %q,%v; want 192.168.0.1,true", gw, ok)
	}
}

func TestParseProcRouteNoDefault(t *testing.T) {
	data := "Iface\tDestination\tGateway\tFlags\n" +
		"eth0\t00010A0A\t00000000\t0001\n"
	if _, ok := parseProcRoute(strings.NewReader(data)); ok {
		t.Error("expected no default gateway")
	}
}

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

	// Pick another ephemeral port, then close it immediately so it's a
	// deterministic "definitely closed" target (avoids flakiness from hardcoded
	// privileged port 1, which may behave differently per platform).
	closedLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := closedLn.Addr().String()
	closedLn.Close()

	c := NewChecker(&config.Config{PingTargets: []string{open, closed}})
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
		if tr.Target == closed {
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
