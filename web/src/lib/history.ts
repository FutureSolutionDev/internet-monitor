import type { LogEvent } from "./types";

export interface HistoryRow {
  event: LogEvent;
  ts: number;
  duration: number | null;
  ongoing: boolean;
}

export function buildRows(events: LogEvent[], now: number, isToday: boolean): HistoryRow[] {
  const sorted = [...events]
    .map((e) => ({ e, ts: Date.parse(e.timestamp) }))
    .filter((x) => Number.isFinite(x.ts))
    .sort((a, b) => a.ts - b.ts);
  const rows = sorted.map(({ e, ts }, i) => {
    const next = sorted[i + 1];
    let duration: number | null = null;
    let ongoing = false;
    if (next) duration = next.e.duration_seconds ?? (next.ts - ts) / 1000;
    else if (isToday) {
      duration = Math.max(0, (now - ts) / 1000);
      ongoing = true;
    }
    return { event: e, ts, duration, ongoing };
  });
  return rows.reverse();
}

export interface DaySummary {
  outages: number;
  downtime: number;
  unstable: number;
  longest: number;
}

export function summarize(rows: HistoryRow[]): DaySummary {
  const s: DaySummary = { outages: 0, downtime: 0, unstable: 0, longest: 0 };
  for (const r of rows) {
    if (r.event.event === "disconnected") {
      s.outages++;
      const d = r.duration ?? 0;
      s.downtime += d;
      s.longest = Math.max(s.longest, d);
    } else if (r.event.event === "degraded") {
      s.unstable++;
    }
  }
  return s;
}
