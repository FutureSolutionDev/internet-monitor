// Package report aggregates connectivity events and metric samples into a rich
// monthly outage summary for the dashboard's printable report. Pure Go, so the
// aggregation logic is unit-tested independently of HTTP/IO.
package report

import (
	"internet-monitor/types"
	"sort"
	"time"
)

// CauseStat counts outages attributable to a single check layer.
type CauseStat struct {
	Outages      int     `json:"outages"`
	DowntimeSecs float64 `json:"downtime_seconds"`
}

// DayStat is the per-day row of the report table.
type DayStat struct {
	Date            string  `json:"date"`
	DowntimeSecs    float64 `json:"downtime_seconds"`
	Outages         int     `json:"outages"`
	WorstOutageSecs float64 `json:"worst_outage_seconds"`
	AvgLatencyMs    int64   `json:"avg_latency_ms"`
	AvgLossPct      float64 `json:"avg_loss_pct"`
	UptimePct       float64 `json:"uptime_pct"`
}

// TrendPoint is a per-day series point for the report charts.
type TrendPoint struct {
	Date         string  `json:"date"`
	AvgLatencyMs int64   `json:"avg_latency_ms"`
	MaxLatencyMs int64   `json:"max_latency_ms"`
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
	Causes            map[string]CauseStat `json:"causes"`
	Days              []DayStat            `json:"days"`
	Trend             []TrendPoint         `json:"trend"`
}

type dayAcc struct {
	downtime     float64
	outages      int
	worst        float64
	latWeighted  int64
	lossWeighted float64
	samples      int
	up           int
	maxLat       int64
}

// Summarize builds a monthly report from connectivity events and metric samples.
// `now` is used to bound an outage that is still ongoing at report time.
func Summarize(events []types.Event, samples []types.MetricSample, month string, now time.Time) MonthlySummary {
	sort.SliceStable(events, func(i, j int) bool { return events[i].Timestamp.Before(events[j].Timestamp) })

	sum := MonthlySummary{
		Month:       month,
		GeneratedAt: now.Format(time.RFC3339),
		Causes:      map[string]CauseStat{"tcp": {}, "http": {}, "dns": {}},
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
			a := getDay(ev.Timestamp)
			a.downtime += dur
			a.outages++
			if dur > a.worst {
				a.worst = dur
			}
		case "degraded":
			sum.DegradedEpisodes++
			sum.TotalDegradedSecs += segmentDuration(events, k, now)
		}
	}
	if sum.Disconnections > 0 {
		sum.MTTRSecs = sum.TotalDowntimeSecs / float64(sum.Disconnections)
	}

	// Trend, uptime and per-day metrics from one-minute samples.
	var totalSamples, upSamples int
	for _, s := range samples {
		totalSamples += s.Samples
		upSamples += s.UpSamples
		a := getDay(s.Timestamp)
		a.latWeighted += s.AvgLatencyMs * int64(s.Samples)
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
			ds.AvgLossPct = a.lossWeighted / float64(a.samples)
			ds.UptimePct = float64(a.up) / float64(a.samples) * 100
		}
		sum.Days = append(sum.Days, ds)
		sum.Trend = append(sum.Trend, TrendPoint{
			Date:         d,
			AvgLatencyMs: ds.AvgLatencyMs,
			MaxLatencyMs: a.maxLat,
			AvgLossPct:   ds.AvgLossPct,
			UptimePct:    ds.UptimePct,
			DowntimeSecs: a.downtime,
		})
	}

	return sum
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
