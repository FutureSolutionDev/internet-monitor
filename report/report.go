// Package report aggregates connectivity events and metric samples into a rich
// monthly outage summary for the dashboard's printable report. Pure Go, so the
// aggregation logic is unit-tested independently of HTTP/IO.
package report

import (
	"fmt"
	"internet-monitor/types"
	"sort"
	"strings"
	"time"
)

// CauseStat counts outages attributable to a single check layer.
type CauseStat struct {
	Outages      int     `json:"outages"`
	DowntimeSecs float64 `json:"downtime_seconds"`
	Pct          float64 `json:"pct"` // share of disconnections this layer was involved in
}

// EventRow is one incident (disconnected/degraded) in the chronological log.
type EventRow struct {
	Time         string  `json:"time"` // RFC3339
	Type         string  `json:"type"` // degraded | disconnected
	DurationSecs float64 `json:"duration_seconds"`
	Cause        string  `json:"cause"` // failed layers, e.g. "tcp+dns"
}

// DayStat is the per-day row of the report table.
type DayStat struct {
	Date            string  `json:"date"`
	DowntimeSecs    float64 `json:"downtime_seconds"`
	Outages         int     `json:"outages"`
	WorstOutageSecs float64 `json:"worst_outage_seconds"`
	AvgLatencyMs    int64   `json:"avg_latency_ms"`
	AvgJitterMs     int64   `json:"avg_jitter_ms"`
	AvgLossPct      float64 `json:"avg_loss_pct"`
	UptimePct       float64 `json:"uptime_pct"`
}

// TrendPoint is a per-day series point for the report charts.
type TrendPoint struct {
	Date         string  `json:"date"`
	AvgLatencyMs int64   `json:"avg_latency_ms"`
	MaxLatencyMs int64   `json:"max_latency_ms"`
	AvgJitterMs  int64   `json:"avg_jitter_ms"`
	AvgLossPct   float64 `json:"avg_loss_pct"`
	UptimePct    float64 `json:"uptime_pct"`
	DowntimeSecs float64 `json:"downtime_seconds"`
}

// MonthlySummary is the full report payload for one month.
type MonthlySummary struct {
	Month             string               `json:"month"`
	GeneratedAt       string               `json:"generated_at"`
	MonitoredSeconds  float64              `json:"monitored_seconds"`
	TotalDowntimeSecs float64              `json:"total_downtime_seconds"`
	TotalDegradedSecs float64              `json:"total_degraded_seconds"`
	UptimePct         float64              `json:"uptime_pct"`
	Disconnections    int                  `json:"disconnections"`
	DegradedEpisodes  int                  `json:"degraded_episodes"`
	LongestOutageSecs float64              `json:"longest_outage_seconds"`
	LongestOutageAt   string               `json:"longest_outage_at"`
	CarriedOverSecs   float64              `json:"carried_over_seconds"` // downtime from an outage that started in a previous month
	MTTRSecs          float64              `json:"mttr_seconds"`
	AvgJitterMs       int64                `json:"avg_jitter_ms"`
	Causes            map[string]CauseStat `json:"causes"`
	OutageTypes       map[string]int       `json:"outage_types"`
	Events            []EventRow           `json:"events"`
	Days              []DayStat            `json:"days"`
	Trend             []TrendPoint         `json:"trend"`
}

type dayAcc struct {
	downtime       float64
	outages        int
	worst          float64
	latWeighted    int64
	jitterWeighted int64
	lossWeighted   float64
	samples        int
	up             int
	maxLat         int64
}

// Summarize builds a monthly report from connectivity events and metric samples.
// `now` is used to bound an outage that is still ongoing at report time.
// Inputs are filtered to events/samples whose timestamps fall in `month`
// (format "YYYY-MM"), so callers can safely pass broader history.
func Summarize(events []types.Event, samples []types.MetricSample, month string, now time.Time) MonthlySummary {
	// Samples are filtered to the month (trend/uptime are month-scoped). Events
	// are NOT filtered: an outage may start in the previous month or end in the
	// next, and the caller passes neighbor-month events so durations are exact.
	// Outage time is clipped to [monthStart, monthEnd) below.
	if month != "" {
		fs := make([]types.MetricSample, 0, len(samples))
		for _, s := range samples {
			if s.Timestamp.Format("2006-01") == month {
				fs = append(fs, s)
			}
		}
		samples = fs
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].Timestamp.Before(events[j].Timestamp) })

	// Month window, in now's location to match locally-stamped event times.
	var monthStart, monthEnd time.Time
	hasWindow := false
	if month != "" {
		var y, mo int
		if _, err := fmt.Sscanf(month, "%d-%d", &y, &mo); err == nil && mo >= 1 && mo <= 12 {
			monthStart = time.Date(y, time.Month(mo), 1, 0, 0, 0, 0, now.Location())
			monthEnd = monthStart.AddDate(0, 1, 0)
			hasWindow = true
		}
	}
	// clip intersects [start,end] with the month window; ok=false if disjoint.
	clip := func(start, end time.Time) (time.Time, time.Time, bool) {
		if hasWindow {
			if start.Before(monthStart) {
				start = monthStart
			}
			if end.After(monthEnd) {
				end = monthEnd
			}
			if !start.Before(monthEnd) || !end.After(monthStart) {
				return start, end, false
			}
		}
		if !start.Before(end) {
			return start, end, false
		}
		return start, end, true
	}
	inWindow := func(t time.Time) bool {
		return !hasWindow || (!t.Before(monthStart) && t.Before(monthEnd))
	}

	sum := MonthlySummary{
		Month:       month,
		GeneratedAt: now.Format(time.RFC3339),
		Causes:      map[string]CauseStat{"tcp": {}, "http": {}, "dns": {}},
		OutageTypes: map[string]int{},
	}

	days := map[string]*dayAcc{}
	getDay := func(t time.Time) *dayAcc {
		d := t.Format("2006-01-02")
		a := days[d]
		if a == nil {
			a = &dayAcc{}
			days[d] = a
		}
		return a
	}
	addCause := func(layer string, dur float64) {
		c := sum.Causes[layer]
		c.Outages++
		c.DowntimeSecs += dur
		sum.Causes[layer] = c
	}

	// Downtime from outages that STARTED this month — the MTTR numerator, kept
	// consistent with the Disconnections denominator (carried-over downtime is
	// excluded from both).
	var inWindowDowntime float64

	// Outages / degraded episodes from transition events.
	for k, ev := range events {
		switch ev.EventType {
		case "disconnected":
			dur := segmentDuration(events, k, now)
			start := ev.Timestamp
			cStart, cEnd, ok := clip(start, start.Add(time.Duration(dur*float64(time.Second))))
			if !ok {
				continue // outage doesn't overlap this month
			}
			clipped := cEnd.Sub(cStart).Seconds()
			sum.TotalDowntimeSecs += clipped

			// Split the clipped interval across day boundaries so per-day
			// downtime reflects what actually happened on each day. Advance by
			// calendar day (AddDate) rather than +24h so DST-transition days
			// (23h/25h) bucket correctly.
			remaining := clipped
			cursor := cStart
			for remaining > 1e-6 {
				dayStart := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, cursor.Location())
				dayEnd := dayStart.AddDate(0, 0, 1)
				slice := dayEnd.Sub(cursor).Seconds()
				if slice > remaining {
					slice = remaining
				}
				getDay(cursor).downtime += slice
				remaining -= slice
				cursor = dayEnd
			}

			startedThisMonth := inWindow(start)
			if startedThisMonth {
				// Outage started this month: count/attribute/log it here.
				sum.Disconnections++
				// "Longest" is the in-month portion so it can never exceed the
				// month's total downtime.
				if clipped > sum.LongestOutageSecs {
					sum.LongestOutageSecs = clipped
					sum.LongestOutageAt = start.Format(time.RFC3339)
				}
				inWindowDowntime += clipped
				if ev.Reason.TCPPingFailed {
					addCause("tcp", clipped)
				}
				if ev.Reason.HTTPFailed {
					addCause("http", clipped)
				}
				if ev.Reason.DNSFailed {
					addCause("dns", clipped)
				}
				sum.OutageTypes[types.Diagnose(types.CheckResult{
					TCPPingOK: !ev.Reason.TCPPingFailed,
					HTTPOK:    !ev.Reason.HTTPFailed,
					DNSOK:     !ev.Reason.DNSFailed,
				})]++
				sum.Events = append(sum.Events, EventRow{
					Time:         start.Format(time.RFC3339),
					Type:         "disconnected",
					DurationSecs: dur,
					Cause:        failedLayers(ev.Reason),
				})
				a := getDay(start)
				a.outages++
				if clipped > a.worst {
					a.worst = clipped
				}
			} else {
				// Outage started in a previous month and spilled into this one:
				// surface its in-month downtime as carried-over, and add a
				// synthetic event row so the log isn't empty-but-with-downtime.
				sum.CarriedOverSecs += clipped
				sum.Events = append(sum.Events, EventRow{
					Time:         cStart.Format(time.RFC3339),
					Type:         "disconnected",
					DurationSecs: clipped,
					Cause:        "carried over from previous month",
				})
			}
		case "degraded":
			ddur := segmentDuration(events, k, now)
			start := ev.Timestamp
			cStart, cEnd, ok := clip(start, start.Add(time.Duration(ddur*float64(time.Second))))
			if !ok {
				continue
			}
			sum.TotalDegradedSecs += cEnd.Sub(cStart).Seconds()
			if inWindow(start) {
				sum.DegradedEpisodes++
				cause := failedLayers(ev.Reason)
				if cause == "" {
					cause = "latency/loss"
				}
				sum.Events = append(sum.Events, EventRow{
					Time:         start.Format(time.RFC3339),
					Type:         "degraded",
					DurationSecs: ddur,
					Cause:        cause,
				})
			}
		}
	}
	if sum.Disconnections > 0 {
		sum.MTTRSecs = inWindowDowntime / float64(sum.Disconnections)
		for layer, c := range sum.Causes {
			c.Pct = float64(c.Outages) / float64(sum.Disconnections) * 100
			sum.Causes[layer] = c
		}
	}

	// Trend, uptime and per-day metrics from one-minute samples.
	var totalSamples, upSamples int
	var jitterWeighted int64
	for _, s := range samples {
		totalSamples += s.Samples
		upSamples += s.UpSamples
		jitterWeighted += s.JitterMs * int64(s.Samples)
		a := getDay(s.Timestamp)
		a.latWeighted += s.AvgLatencyMs * int64(s.Samples)
		a.jitterWeighted += s.JitterMs * int64(s.Samples)
		a.lossWeighted += s.AvgLossPct * float64(s.Samples)
		a.samples += s.Samples
		a.up += s.UpSamples
		if s.MaxLatencyMs > a.maxLat {
			a.maxLat = s.MaxLatencyMs
		}
	}
	sum.MonitoredSeconds = float64(len(samples)) * 60
	if totalSamples > 0 {
		sum.UptimePct = float64(upSamples) / float64(totalSamples) * 100
		sum.AvgJitterMs = jitterWeighted / int64(totalSamples)
	}

	dates := make([]string, 0, len(days))
	for d := range days {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	for _, d := range dates {
		a := days[d]
		ds := DayStat{Date: d, DowntimeSecs: a.downtime, Outages: a.outages, WorstOutageSecs: a.worst}
		if a.samples > 0 {
			ds.AvgLatencyMs = a.latWeighted / int64(a.samples)
			ds.AvgJitterMs = a.jitterWeighted / int64(a.samples)
			ds.AvgLossPct = a.lossWeighted / float64(a.samples)
			ds.UptimePct = float64(a.up) / float64(a.samples) * 100
		}
		sum.Days = append(sum.Days, ds)
		sum.Trend = append(sum.Trend, TrendPoint{
			Date:         d,
			AvgLatencyMs: ds.AvgLatencyMs,
			MaxLatencyMs: a.maxLat,
			AvgJitterMs:  ds.AvgJitterMs,
			AvgLossPct:   ds.AvgLossPct,
			UptimePct:    ds.UptimePct,
			DowntimeSecs: a.downtime,
		})
	}

	return sum
}

// failedLayers joins the failed check layers of an event reason, e.g. "tcp+dns".
func failedLayers(r types.EventReason) string {
	var p []string
	if r.TCPPingFailed {
		p = append(p, "tcp")
	}
	if r.HTTPFailed {
		p = append(p, "http")
	}
	if r.DNSFailed {
		p = append(p, "dns")
	}
	return strings.Join(p, "+")
}

// segmentDuration returns how long events[k]'s state lasted: the next event's
// DurationSeconds (which records the duration of the state that just ended), or
// (now - timestamp) if events[k] is the last, still-ongoing event.
func segmentDuration(events []types.Event, k int, now time.Time) float64 {
	if k+1 < len(events) {
		return events[k+1].DurationSeconds
	}
	if d := now.Sub(events[k].Timestamp).Seconds(); d > 0 {
		return d
	}
	return 0
}
