import type { Vars } from "./i18n";
import type { MessageKey } from "./strings";
import type { ConfigView } from "./types";

export const MAX_TARGETS = 10;

export type SecretDraft = { mode: "keep" } | { mode: "set"; value: string } | { mode: "clear" };

export interface Draft {
  check_interval_sec: string;
  fail_threshold: string;
  packet_loss_threshold: string;
  latency_threshold_ms: string;
  use_icmp: boolean;
  ping_targets: string[];
  http_targets: string[];
  dns_targets: string[];
  webhook: SecretDraft;
  telegram_token: SecretDraft;
  telegram_chat_id: string;
  dashboard_port: string;
  log_dir: string;
  speed_timeout: string;
  speed_parallel: string;
  speed_alert: string;
  speed_schedule: string;
  speed_download_targets: string[];
  speed_upload_target: string;
}

export type FieldKey = keyof Draft | "targets";
export interface FieldIssue {
  key: MessageKey;
  vars?: Vars;
}
export type Issues = Partial<Record<FieldKey, FieldIssue>>;

export function draftFrom(c: ConfigView): Draft {
  return {
    check_interval_sec: String(c.check_interval_sec),
    fail_threshold: String(c.fail_threshold),
    packet_loss_threshold: String(c.packet_loss_threshold),
    latency_threshold_ms: String(c.latency_threshold_ms),
    use_icmp: c.use_icmp,
    ping_targets: [...(c.ping_targets ?? [])],
    http_targets: [...(c.http_targets ?? [])],
    dns_targets: [...(c.dns_targets ?? [])],
    webhook: { mode: "keep" },
    telegram_token: { mode: "keep" },
    telegram_chat_id: c.telegram_chat_id ?? "",
    dashboard_port: String(c.dashboard_port),
    log_dir: c.log_dir,
    speed_timeout: String(c.speed_test.timeout_seconds),
    speed_parallel: String(c.speed_test.parallel_connections),
    speed_alert: String(c.speed_test.alert_threshold_mbps),
    speed_schedule: String(c.speed_test.schedule_minutes),
    speed_download_targets: [...(c.speed_test.download_targets ?? [])],
    speed_upload_target: c.speed_test.upload_target ?? "",
  };
}

const WEBHOOK_HOSTS: Record<string, string> = {
  "discord.com": "/api/webhooks/",
  "discordapp.com": "/api/webhooks/",
  "canary.discord.com": "/api/webhooks/",
  "ptb.discord.com": "/api/webhooks/",
  "hooks.slack.com": "/services/",
};

export function isSupportedWebhook(raw: string): boolean {
  let u: URL;
  try {
    u = new URL(raw.trim());
  } catch {
    return false;
  }
  if (u.protocol !== "https:" || u.username || u.password || u.port) return false;
  const prefix = WEBHOOK_HOSTS[u.hostname.toLowerCase()];
  return !!prefix && u.pathname.startsWith(prefix) && u.pathname.length > prefix.length;
}

export function isHostPort(v: string): boolean {
  const m = /^(\[[0-9a-fA-F:.]+\]|[^\s:/[\]]+):(\d{1,5})$/.exec(v.trim());
  if (!m) return false;
  const port = Number(m[2]);
  return port >= 1 && port <= 65535;
}

export function isHttpURL(v: string, httpsOnly = false): boolean {
  try {
    const u = new URL(v.trim());
    return httpsOnly ? u.protocol === "https:" : u.protocol === "http:" || u.protocol === "https:";
  } catch {
    return false;
  }
}

export function isHostname(v: string): boolean {
  return /^(?=.{1,253}$)([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9-]{1,63}\.?$/.test(
    v.trim(),
  );
}

export function isSafeLogDir(v: string): boolean {
  const dir = v.trim();
  if (!dir || /^([a-zA-Z]:|[\\/])/.test(dir)) return false;
  return !dir.split(/[\\/]+/).some((p) => p === "..");
}

function numberIssue(raw: string, min: number, max?: number, integer = true): FieldIssue | undefined {
  if (raw.trim() === "") return { key: "err_required" };
  const n = Number(raw);
  if (!Number.isFinite(n) || (integer && !Number.isInteger(n)) || n < min || (max !== undefined && n > max)) {
    return max === undefined ? { key: "err_min", vars: { min } } : { key: "err_range", vars: { min, max } };
  }
  return undefined;
}

export function listIssue(
  items: string[],
  check: (v: string) => boolean,
  key: MessageKey,
): FieldIssue | undefined {
  const seen = new Set<string>();
  for (const item of items) {
    const v = item.trim();
    if (!v) continue;
    if (!check(v)) return { key };
    if (seen.has(v)) return { key: "err_duplicate" };
    seen.add(v);
  }
  return undefined;
}

export function validateDraft(d: Draft): Issues {
  const issues: Issues = {};
  const put = (k: FieldKey, issue: FieldIssue | undefined) => {
    if (issue) issues[k] = issue;
  };
  put("check_interval_sec", numberIssue(d.check_interval_sec, 1, 3600));
  put("fail_threshold", numberIssue(d.fail_threshold, 1, 100));
  put("packet_loss_threshold", numberIssue(d.packet_loss_threshold, 0, 100, false));
  put("latency_threshold_ms", numberIssue(d.latency_threshold_ms, 0, 60000));
  put("dashboard_port", numberIssue(d.dashboard_port, 1, 65535));
  put("speed_timeout", numberIssue(d.speed_timeout, 5, 120));
  put("speed_parallel", numberIssue(d.speed_parallel, 1, 16));
  put("speed_alert", numberIssue(d.speed_alert, 0, undefined, false));
  put("speed_schedule", numberIssue(d.speed_schedule, 0, 10080));
  put("ping_targets", listIssue(d.ping_targets, isHostPort, "err_hostport"));
  put(
    "http_targets",
    listIssue(d.http_targets, (v) => isHttpURL(v), "err_url"),
  );
  put("dns_targets", listIssue(d.dns_targets, isHostname, "err_hostname"));
  put(
    "speed_download_targets",
    listIssue(d.speed_download_targets, (v) => isHttpURL(v), "err_url"),
  );
  if (!d.speed_download_targets.some((v) => v.trim()))
    issues.speed_download_targets = { key: "err_required" };
  if (d.speed_upload_target.trim() && !isHttpURL(d.speed_upload_target))
    issues.speed_upload_target = { key: "err_url" };
  if (![d.ping_targets, d.http_targets, d.dns_targets].some((l) => l.some((v) => v.trim()))) {
    issues.targets = { key: "err_targets" };
  }
  if (d.webhook.mode === "set" && !isSupportedWebhook(d.webhook.value))
    issues.webhook = { key: "err_webhook" };
  if (d.telegram_token.mode === "set" && !d.telegram_token.value.trim())
    issues.telegram_token = { key: "err_required" };
  if (!isSafeLogDir(d.log_dir)) issues.log_dir = { key: "err_logdir" };
  return issues;
}

const clean = (items: string[]) => items.map((v) => v.trim()).filter(Boolean);
const sameList = (a: string[], b: string[] | null | undefined) => {
  const x = clean(a);
  const y = b ?? [];
  return x.length === y.length && x.every((v, i) => v === y[i]);
};

export function buildChanges(d: Draft, base: ConfigView): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  const num = (raw: string, current: number, key: string) => {
    const n = Number(raw);
    if (raw.trim() !== "" && Number.isFinite(n) && n !== current) out[key] = n;
  };
  num(d.check_interval_sec, base.check_interval_sec, "check_interval_sec");
  num(d.fail_threshold, base.fail_threshold, "fail_threshold");
  num(d.packet_loss_threshold, base.packet_loss_threshold, "packet_loss_threshold");
  num(d.latency_threshold_ms, base.latency_threshold_ms, "latency_threshold_ms");
  num(d.dashboard_port, base.dashboard_port, "dashboard_port");
  if (d.use_icmp !== base.use_icmp) out.use_icmp = d.use_icmp;
  if (!sameList(d.ping_targets, base.ping_targets)) out.ping_targets = clean(d.ping_targets);
  if (!sameList(d.http_targets, base.http_targets)) out.http_targets = clean(d.http_targets);
  if (!sameList(d.dns_targets, base.dns_targets)) out.dns_targets = clean(d.dns_targets);
  if (d.telegram_chat_id.trim() !== (base.telegram_chat_id ?? ""))
    out.telegram_chat_id = d.telegram_chat_id.trim();
  if (d.log_dir.trim() !== base.log_dir) out.log_dir = d.log_dir.trim();
  if (d.webhook.mode === "set") out.webhook_url = d.webhook.value.trim();
  if (d.webhook.mode === "clear") out.webhook_url = "";
  if (d.telegram_token.mode === "set") out.telegram_bot_token = d.telegram_token.value.trim();
  if (d.telegram_token.mode === "clear") out.telegram_bot_token = "";

  const st = base.speed_test;
  const speed: Record<string, unknown> = {};
  const snum = (raw: string, current: number, key: string) => {
    const n = Number(raw);
    if (raw.trim() !== "" && Number.isFinite(n) && n !== current) speed[key] = n;
  };
  snum(d.speed_timeout, st.timeout_seconds, "timeout_seconds");
  snum(d.speed_parallel, st.parallel_connections, "parallel_connections");
  snum(d.speed_alert, st.alert_threshold_mbps, "alert_threshold_mbps");
  snum(d.speed_schedule, st.schedule_minutes, "schedule_minutes");
  if (!sameList(d.speed_download_targets, st.download_targets)) {
    speed.download_targets = clean(d.speed_download_targets);
  }
  if (d.speed_upload_target.trim() !== (st.upload_target ?? ""))
    speed.upload_target = d.speed_upload_target.trim();
  if (Object.keys(speed).length) out.speed_test = speed;
  return out;
}

const SECTION_FIELDS: Record<string, FieldKey[]> = {
  monitoring: ["check_interval_sec", "fail_threshold", "packet_loss_threshold", "latency_threshold_ms"],
  targets: ["ping_targets", "http_targets", "dns_targets", "targets"],
  alerts: ["webhook", "telegram_token", "telegram_chat_id"],
  speed: [
    "speed_timeout",
    "speed_parallel",
    "speed_alert",
    "speed_schedule",
    "speed_download_targets",
    "speed_upload_target",
  ],
  app: ["dashboard_port", "log_dir"],
};

export function sectionOfField(field: FieldKey): string {
  for (const [section, fields] of Object.entries(SECTION_FIELDS)) {
    if (fields.includes(field)) return section;
  }
  return "monitoring";
}

export function serverFieldToDraft(field: string | undefined): FieldKey | undefined {
  switch (field) {
    case "webhook_url":
      return "webhook";
    case "log_dir":
      return "log_dir";
    case "targets":
      return "targets";
    default:
      return undefined;
  }
}

const same = (a: unknown, b: unknown) => JSON.stringify(a) === JSON.stringify(b);

export function rebaseDraft(
  draft: Draft,
  oldBase: ConfigView,
  newBase: ConfigView,
): { draft: Draft; conflicts: (keyof Draft)[] } {
  const before = draftFrom(oldBase);
  const after = draftFrom(newBase);
  const out: Draft = { ...after };
  const conflicts: (keyof Draft)[] = [];
  for (const key of Object.keys(draft) as (keyof Draft)[]) {
    if (same(draft[key], before[key])) continue;
    let theirs = !same(after[key], before[key]);
    if (key === "webhook") theirs = oldBase.secret_set.webhook_url !== newBase.secret_set.webhook_url;
    if (key === "telegram_token") {
      theirs = oldBase.secret_set.telegram_bot_token !== newBase.secret_set.telegram_bot_token;
    }
    if (theirs && !same(draft[key], after[key])) conflicts.push(key);
    Object.assign(out, { [key]: draft[key] });
  }
  return { draft: out, conflicts };
}
