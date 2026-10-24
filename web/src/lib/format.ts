import { type Lang, localeOf, type Translate } from "./i18n";
import type { EventEntry, LogEvent, Status } from "./types";

export function formatNumber(lang: Lang, value: number, digits = 0): string {
  return new Intl.NumberFormat(localeOf(lang), {
    maximumFractionDigits: digits,
    minimumFractionDigits: digits,
  }).format(value);
}

export function formatPct(lang: Lang, value: number | null | undefined): string {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  if (value >= 100) return `${formatNumber(lang, 100, 0)}%`;
  const digits = value >= 99 ? 2 : 1;
  const scale = 10 ** digits;
  return `${formatNumber(lang, Math.floor(value * scale) / scale, digits)}%`;
}

export function formatDuration(t: Translate, seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return "—";
  const s = Math.round(seconds);
  if (s < 60) return t("dur_s", { n: s });
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  const rest = s % 60;
  const parts: string[] = [];
  if (d) parts.push(t("dur_d", { n: d }));
  if (h) parts.push(t("dur_h", { n: h }));
  if (m && !d) parts.push(t("dur_m", { n: m }));
  if (rest && !d && !h) parts.push(t("dur_s", { n: rest }));
  return parts.slice(0, 2).join(" ");
}

export function latencyParts(t: Translate, lang: Lang, ms: number): { value: string; unit: string } {
  if (!Number.isFinite(ms)) return { value: "—", unit: "" };
  const whole = Math.max(0, Math.round(ms));
  if (whole < 1000) return { value: formatNumber(lang, whole), unit: t("ms_unit") };
  const digits = ms < 10_000 ? 2 : 1;
  const s = Number((ms / 1000).toFixed(digits));
  if (s < 60) {
    const value = new Intl.NumberFormat(localeOf(lang), { maximumFractionDigits: digits }).format(s);
    return { value, unit: t("s_unit") };
  }
  return { value: formatDuration(t, Math.max(60, ms / 1000)), unit: "" };
}

export function formatLatency(t: Translate, lang: Lang, ms: number): string {
  const { value, unit } = latencyParts(t, lang, ms);
  return unit ? `${value} ${unit}` : value;
}

export function formatAgo(t: Translate, fromMs: number, nowMs: number): string {
  const diff = Math.max(0, (nowMs - fromMs) / 1000);
  if (diff < 5) return t("ago_now");
  return t("ago", { duration: formatDuration(t, diff) });
}

export function formatTime(lang: Lang, value: number | string | Date, withSeconds = true): string {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat(localeOf(lang), {
    hour: "2-digit",
    minute: "2-digit",
    second: withSeconds ? "2-digit" : undefined,
  }).format(date);
}

export function formatDate(
  lang: Lang,
  value: number | string | Date,
  opts?: Intl.DateTimeFormatOptions,
): string {
  const date =
    typeof value === "string" && /^\d{4}-\d{2}-\d{2}$/.test(value) ? parseLocalDate(value) : new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat(
    localeOf(lang),
    opts ?? { weekday: "short", day: "numeric", month: "short" },
  ).format(date);
}

export function formatMonth(lang: Lang, month: string): string {
  const [y, m] = month.split("-").map(Number);
  if (!y || !m) return month;
  return new Intl.DateTimeFormat(localeOf(lang), { month: "long", year: "numeric" }).format(
    new Date(y, m - 1, 1),
  );
}

export function parseLocalDate(value: string): Date {
  const [y, m, d] = value.split("-").map(Number);
  return new Date(y ?? 1970, (m ?? 1) - 1, d ?? 1);
}

export function localDateKey(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

export function monthKey(date: Date): string {
  return localDateKey(date).slice(0, 7);
}

export function shiftMonth(month: string, delta: number): string {
  const [y, m] = month.split("-").map(Number);
  return monthKey(new Date(y ?? 1970, (m ?? 1) - 1 + delta, 1));
}

export function statusKey(status: Status | string) {
  switch (status) {
    case "connected":
      return "status_connected" as const;
    case "degraded":
      return "status_degraded" as const;
    case "disconnected":
      return "status_disconnected" as const;
    default:
      return "status_checking" as const;
  }
}

export function eventKey(type: string) {
  switch (type) {
    case "connected":
      return "ev_connected" as const;
    case "degraded":
      return "ev_degraded" as const;
    default:
      return "ev_disconnected" as const;
  }
}

interface CauseInput {
  tcp: boolean;
  http: boolean;
  dns: boolean;
  loss: number;
  latency: number;
  type: string;
}

export function causeOfEntry(e: EventEntry): CauseInput {
  return {
    tcp: e.tcp_failed,
    http: e.http_failed,
    dns: e.dns_failed,
    loss: e.packet_loss_pct,
    latency: e.latency_ms,
    type: e.event_type,
  };
}

export function causeOfLog(e: LogEvent): CauseInput {
  return {
    tcp: e.reason?.tcp_ping_failed ?? false,
    http: e.reason?.http_failed ?? false,
    dns: e.reason?.dns_failed ?? false,
    loss: e.reason?.packet_loss_pct ?? 0,
    latency: e.reason?.avg_latency_ms ?? 0,
    type: e.event,
  };
}

export function describeCause(t: Translate, lang: Lang, c: CauseInput): string {
  if (c.type === "connected") return t("cause_none");
  const parts: string[] = [];
  if (c.tcp) parts.push(t("cause_tcp"));
  if (c.http) parts.push(t("cause_http"));
  if (c.dns) parts.push(t("cause_dns"));
  if (!parts.length && c.loss > 0) parts.push(t("cause_loss", { value: formatNumber(lang, c.loss) }));
  if (!parts.length && c.latency > 0)
    parts.push(t("cause_latency", { value: formatLatency(t, lang, c.latency) }));
  return parts.length ? parts.join(" · ") : t("cause_none");
}

export function describeReportCause(t: Translate, cause: string): string {
  if (cause.startsWith("carried")) return t("cause_carried");
  if (cause === "latency/loss") return `${t("cause_name_latency")} / ${t("cause_name_loss")}`;
  if (!cause) return t("cause_name_unknown");
  return cause
    .split("+")
    .map((c) => {
      switch (c) {
        case "tcp":
          return t("cause_name_tcp");
        case "http":
          return t("cause_name_http");
        case "dns":
          return t("cause_name_dns");
        default:
          return t("cause_name_unknown");
      }
    })
    .join(" + ");
}
