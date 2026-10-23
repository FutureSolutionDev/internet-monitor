import { describe, expect, test } from "vitest";
import {
  buildChanges,
  draftFrom,
  isHostPort,
  isSafeLogDir,
  isSupportedWebhook,
  rebaseDraft,
  validateDraft,
} from "./settingsForm";
import type { ConfigView } from "./types";

const base: ConfigView = {
  check_interval_sec: 5,
  ping_targets: ["8.8.8.8:53", "1.1.1.1:53"],
  http_targets: ["https://connectivitycheck.gstatic.com/generate_204"],
  dns_targets: ["www.google.com"],
  fail_threshold: 3,
  packet_loss_threshold: 20,
  latency_threshold_ms: 500,
  log_dir: "logs",
  webhook_url: "",
  telegram_bot_token: "",
  telegram_chat_id: "",
  dashboard_port: 8765,
  use_icmp: false,
  language: "en",
  speed_test: {
    download_targets: ["https://speed.cloudflare.com/__down"],
    parallel_connections: 4,
    timeout_seconds: 30,
    upload_target: "https://speed.cloudflare.com/__up",
    alert_threshold_mbps: 0,
    schedule_minutes: 0,
  },
  secret_set: { webhook_url: true, telegram_bot_token: true },
};

describe("buildChanges", () => {
  test("an untouched form sends nothing, so saved secrets are never overwritten", () => {
    expect(buildChanges(draftFrom(base), base)).toEqual({});
  });

  test("unrelated edits omit secret keys entirely", () => {
    const d = { ...draftFrom(base), check_interval_sec: "10" };
    const changes = buildChanges(d, base);
    expect(changes).toEqual({ check_interval_sec: 10 });
    expect("webhook_url" in changes).toBe(false);
    expect("telegram_bot_token" in changes).toBe(false);
  });

  test("replacing a secret sends the new value", () => {
    const d = {
      ...draftFrom(base),
      webhook: { mode: "set" as const, value: " https://discord.com/api/webhooks/1/a " },
    };
    expect(buildChanges(d, base)).toEqual({ webhook_url: "https://discord.com/api/webhooks/1/a" });
  });

  test("removing a secret sends an empty string", () => {
    const d = { ...draftFrom(base), telegram_token: { mode: "clear" as const } };
    expect(buildChanges(d, base)).toEqual({ telegram_bot_token: "" });
  });

  test("zero is a real value, not a fallback to the default", () => {
    const d = { ...draftFrom(base), packet_loss_threshold: "0", latency_threshold_ms: "0" };
    expect(buildChanges(d, base)).toEqual({ packet_loss_threshold: 0, latency_threshold_ms: 0 });
  });

  test("speed test changes are nested and partial", () => {
    const d = { ...draftFrom(base), speed_schedule: "60", speed_upload_target: "" };
    expect(buildChanges(d, base)).toEqual({ speed_test: { schedule_minutes: 60, upload_target: "" } });
  });

  test("lists are trimmed and blank rows dropped", () => {
    const d = { ...draftFrom(base), dns_targets: [" www.google.com ", "", "example.com"] };
    expect(buildChanges(d, base)).toEqual({ dns_targets: ["www.google.com", "example.com"] });
  });
});

describe("validateDraft", () => {
  test("defaults are valid", () => {
    expect(validateDraft(draftFrom(base))).toEqual({});
  });

  test("flags out-of-range numbers and empty fields", () => {
    const d = { ...draftFrom(base), check_interval_sec: "0", fail_threshold: "", dashboard_port: "70000" };
    const issues = validateDraft(d);
    expect(issues.check_interval_sec?.key).toBe("err_range");
    expect(issues.fail_threshold?.key).toBe("err_required");
    expect(issues.dashboard_port?.key).toBe("err_range");
  });

  test("requires at least one target overall", () => {
    const d = { ...draftFrom(base), ping_targets: [], http_targets: [], dns_targets: [""] };
    expect(validateDraft(d).targets?.key).toBe("err_targets");
  });

  test("rejects unsupported webhook hosts only when replacing", () => {
    const d = {
      ...draftFrom(base),
      webhook: { mode: "set" as const, value: "https://evil.example/api/webhooks/1/x" },
    };
    expect(validateDraft(d).webhook?.key).toBe("err_webhook");
  });

  test("flags duplicates", () => {
    const d = { ...draftFrom(base), ping_targets: ["8.8.8.8:53", "8.8.8.8:53"] };
    expect(validateDraft(d).ping_targets?.key).toBe("err_duplicate");
  });
});

describe("validators", () => {
  test("webhook allowlist mirrors the backend", () => {
    expect(isSupportedWebhook("https://discord.com/api/webhooks/1/abc")).toBe(true);
    expect(isSupportedWebhook("https://hooks.slack.com/services/T/B/x")).toBe(true);
    expect(isSupportedWebhook("http://discord.com/api/webhooks/1/abc")).toBe(false);
    expect(isSupportedWebhook("https://discord.com.evil.io/api/webhooks/1/abc")).toBe(false);
    expect(isSupportedWebhook("https://discord.com:8443/api/webhooks/1/abc")).toBe(false);
    expect(isSupportedWebhook("https://user@discord.com/api/webhooks/1/abc")).toBe(false);
    expect(isSupportedWebhook("https://discord.com/api/webhooks/")).toBe(false);
  });

  test("host:port", () => {
    expect(isHostPort("8.8.8.8:53")).toBe(true);
    expect(isHostPort("[2606:4700:4700::1111]:53")).toBe(true);
    expect(isHostPort("example.com")).toBe(false);
    expect(isHostPort("example.com:0")).toBe(false);
  });

  test("log dir stays inside the app folder", () => {
    expect(isSafeLogDir("logs")).toBe(true);
    expect(isSafeLogDir("data/logs")).toBe(true);
    expect(isSafeLogDir("C:\\logs")).toBe(false);
    expect(isSafeLogDir("/var/log")).toBe(false);
    expect(isSafeLogDir("../logs")).toBe(false);
    expect(isSafeLogDir("a\\..\\..\\b")).toBe(false);
  });
});

describe("rebaseDraft", () => {
  test("a language change elsewhere keeps my edits without conflict", () => {
    const mine = { ...draftFrom(base), check_interval_sec: "9" };
    const remote = { ...base, language: "ar" as const };
    const r = rebaseDraft(mine, base, remote);
    expect(r.conflicts).toEqual([]);
    expect(r.draft.check_interval_sec).toBe("9");
    expect(buildChanges(r.draft, remote)).toEqual({ check_interval_sec: 9 });
  });

  test("untouched fields follow the remote value", () => {
    const remote = { ...base, fail_threshold: 7 };
    const r = rebaseDraft({ ...draftFrom(base), check_interval_sec: "9" }, base, remote);
    expect(r.draft.fail_threshold).toBe("7");
  });

  test("the same field changed on both sides is a conflict", () => {
    const remote = { ...base, check_interval_sec: 30 };
    const r = rebaseDraft({ ...draftFrom(base), check_interval_sec: "9" }, base, remote);
    expect(r.conflicts).toEqual(["check_interval_sec"]);
    expect(r.draft.check_interval_sec).toBe("9");
  });

  test("a secret replaced here and removed elsewhere is a conflict", () => {
    const remote = { ...base, secret_set: { webhook_url: false, telegram_bot_token: true } };
    const mine = {
      ...draftFrom(base),
      webhook: { mode: "set" as const, value: "https://discord.com/api/webhooks/1/a" },
    };
    expect(rebaseDraft(mine, base, remote).conflicts).toEqual(["webhook"]);
  });
});
