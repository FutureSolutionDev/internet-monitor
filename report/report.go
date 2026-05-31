// Package report aggregates connectivity events and metric samples into a rich
// monthly outage summary for the dashboard's printable report. Pure Go, so the
// aggregation logic is unit-tested independently of HTTP/IO.
package report

import (
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
	if month != "" {
		fe := make([]types.Event, 0, len(events))
		for _, ev := range events {
			if ev.Timestamp.Format("2006-01") == month {
				fe = append(fe, ev)
			}
		}
		events = fe
		fs := make([]types.MetricSample, 0, len(samples))
		for _, s := range samples {
			if s.Timestamp.Format("2006-01") == month {
				fs = append(fs, s)
			}
		}
		samples = fs
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].Timestamp.Before(events[j].Timestamp) })

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

	// Outages / degraded episodes from transition events.
	for k, ev := range events {
		switch ev.EventType {
		case "disconnected":
			sum.Disconnections++
			dur := segmentDuration(events, k, now)
			sum.TotalDowntimeSecs += dur
			if dur > sum.LongestOutageSecs {
				sum.LongestOutageSecs = dur
				sum.LongestOutageAt = ev.Timestamp.Format(time.RFC3339)
			}
			if ev.Reason.TCPPingFailed {
				addCause("tcp", dur)
			}
			if ev.Reason.HTTPFailed {
				addCause("http", dur)
			}
			if ev.Reason.DNSFailed {
				addCause("dns", dur)
			}
			// Classify the outage by type (reason booleans are failures, so the
			// ok flags are their negation).
			sum.OutageTypes[types.Diagnose(types.CheckResult{
				TCPPingOK: !ev.Reason.TCPPingFailed,
				HTTPOK:    !ev.Reason.HTTPFailed,
				DNSOK:     !ev.Reason.DNSFailed,
			})]++
			sum.Events = append(sum.Events, EventRow{
				Time:         ev.Timestamp.Format(time.RFC3339),
				Type:         "disconnected",
				DurationSecs: dur,
				Cause:        failedLayers(ev.Reason),
			})
			// Split the outage across day boundaries so per-day downtime
			// reflects what actually happened on each day. The outage count
			// and worst-outage stay attributed to the start day.
			start := ev.Timestamp
			a := getDay(start)
			a.outages++
			if dur > a.worst {
				a.worst = dur
			}
			remaining := dur
			cursor := start
			for remaining > 0 {
				dayEnd := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, cursor.Location()).Add(24 * time.Hour)
				slice := dayEnd.Sub(cursor).Seconds()
				if slice > remaining {
					slice = remaining
				}
				getDay(cursor).downtime += slice
				remaining -= slice
				cursor = dayEnd
			}
		case "degraded":
			sum.DegradedEpisodes++
			ddur := segmentDuration(events, k, now)
			sum.TotalDegradedSecs += ddur
			cause := failedLayers(ev.Reason)
			if cause == "" {
				cause = "latency/loss"
			}
			sum.Events = append(sum.Events, EventRow{
				Time:         ev.Timestamp.Format(time.RFC3339),
				Type:         "degraded",
				DurationSecs: ddur,
				Cause:        cause,
			})
		}
	}
	if sum.Disconnections > 0 {
		sum.MTTRSecs = sum.TotalDowntimeSecs / float64(sum.Disconnections)
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
