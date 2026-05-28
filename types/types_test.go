package types

import "testing"

func TestDiagnose(t *testing.T) {
	r := func(tcp, http, dns bool) CheckResult {
		return CheckResult{TCPPingOK: tcp, HTTPOK: http, DNSOK: dns}
	}
	cases := []struct {
		in   CheckResult
		want string
	}{
		{r(true, true, true), "ok"},
		{r(false, false, false), "down"},
		{r(false, true, true), "down"}, // no L4 reachability dominates
		{r(true, true, false), "dns"},
		{r(true, false, true), "http"},
	}
	for _, c := range cases {
		if got := Diagnose(c.in); got != c.want {
			t.Errorf("Diagnose(%+v) = %q, want %q", c.in, got, c.want)
		}
	}
}
