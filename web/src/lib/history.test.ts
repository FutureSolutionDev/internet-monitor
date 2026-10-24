import { describe, expect, test } from "vitest";
import { toCSV } from "./csv";
import { formatDuration, formatLatency } from "./format";
import { buildRows, summarize } from "./history";
import { translate } from "./i18n";
import type { LogEvent } from "./types";

const ev = (timestamp: string, event: LogEvent["event"], duration_seconds = 0): LogEvent => ({
  timestamp,
  event,
  duration_seconds,
  reason: {
    tcp_ping_failed: false,
    http_failed: false,
    dns_failed: false,
    packet_loss_pct: 0,
    avg_latency_ms: 0,
  },
});

describe("history rows", () => {
  const logs = [
    ev("2026-10-01T10:05:00Z", "connected", 300),
    ev("2026-10-01T10:00:00Z", "disconnected", 7200),
    ev("2026-10-01T11:00:00Z", "degraded", 3300),
  ];

  test("each row's duration comes from the next event, newest first", () => {
    const rows = buildRows(logs, Date.parse("2026-10-01T12:00:00Z"), false);
    expect(rows.map((r) => r.event.event)).toEqual(["degraded", "connected", "disconnected"]);
    expect(rows[2]?.duration).toBe(300);
    expect(rows[0]?.duration).toBeNull();
  });

  test("the last event today is ongoing", () => {
    const rows = buildRows(logs, Date.parse("2026-10-01T11:30:00Z"), true);
    expect(rows[0]?.ongoing).toBe(true);
    expect(rows[0]?.duration).toBe(1800);
  });

  test("summary counts outages and downtime", () => {
    const s = summarize(buildRows(logs, 0, false));
    expect(s).toEqual({ outages: 1, downtime: 300, unstable: 1, longest: 300 });
  });
});

describe("formatting", () => {
  const t = (key: Parameters<typeof translate>[1], vars?: Record<string, string | number>) =>
    translate("en", key, vars);

  test("durations are compact", () => {
    expect(formatDuration(t, 42)).toBe("42s");
    expect(formatDuration(t, 125)).toBe("2m 5s");
    expect(formatDuration(t, 3725)).toBe("1h 2m");
    expect(formatDuration(t, 90000)).toBe("1d 1h");
  });

  test("latency scales from ms to seconds and minutes", () => {
    expect(formatLatency(t, "en", 46)).toBe("46 ms");
    expect(formatLatency(t, "en", 1057)).toBe("1.06 s");
    expect(formatLatency(t, "en", 1000)).toBe("1 s");
    expect(formatLatency(t, "en", 12_345)).toBe("12.3 s");
    expect(formatLatency(t, "en", 125_000)).toBe("2m 5s");
    expect(formatLatency(t, "en", Number.NaN)).toBe("—");
  });

  test("arabic strings interpolate", () => {
    expect(translate("ar", "test_ok", { ok: 2, total: 3 })).toBe("2 من 3 شغّالين");
  });

  test("csv escapes quotes and neutralises formulas", () => {
    const csv = toCSV([
      ["a", 'he said "hi"'],
      ["=cmd()", -5],
    ]);
    expect(csv).toBe('﻿a,"he said ""hi"""\r\n\'=cmd(),-5\r\n');
  });
});
