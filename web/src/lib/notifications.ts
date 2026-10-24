import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "./api";
import { describeCause, formatDuration, formatNumber } from "./format";
import type { Lang, MessageKey, Translate } from "./i18n";
import { useLive } from "./live";
import type { AlertChannel, AlertKind, MuteReason, NotificationList, NotificationRecord } from "./types";

export const KIND_LABEL: Record<AlertKind, MessageKey> = {
  down: "n_kind_down",
  degraded: "n_kind_degraded",
  restored: "n_kind_restored",
  reminder: "n_kind_reminder",
  speed: "n_kind_speed",
  test: "n_kind_test",
};

export const KIND_TONE: Record<AlertKind, string> = {
  down: "var(--accent-coral)",
  degraded: "var(--accent-amber)",
  restored: "var(--accent-green)",
  reminder: "var(--accent-coral)",
  speed: "var(--accent-blue)",
  test: "var(--accent-violet)",
};

export const CHANNEL_NAME: Record<AlertChannel, MessageKey> = {
  desktop: "n_ch_desktop",
  sound: "n_ch_sound",
  webhook: "n_ch_webhook",
  telegram: "n_ch_telegram",
};

export const REASON_LABEL: Record<MuteReason, MessageKey> = {
  disabled: "n_reason_disabled",
  quiet_hours: "n_reason_quiet_hours",
  sound_off: "n_reason_sound_off",
  short_outage: "n_reason_short_outage",
  startup: "n_reason_startup",
};

const CHANNEL_ORDER: AlertChannel[] = ["desktop", "sound", "webhook", "telegram"];

export function mutedChannels(r: NotificationRecord): { channel: AlertChannel; reason: MuteReason }[] {
  const muted = r.muted ?? {};
  return CHANNEL_ORDER.flatMap((channel) => {
    const reason = muted[channel];
    return reason ? [{ channel, reason }] : [];
  });
}

export function sentChannels(r: NotificationRecord): AlertChannel[] {
  return CHANNEL_ORDER.filter((c) => r.sent.includes(c));
}

export function noticeDetail(t: Translate, lang: Lang, r: NotificationRecord): string {
  if (r.kind === "speed" && r.speed) {
    const parts = [`↓ ${formatNumber(lang, r.speed.download_mbps, 1)} ${t("speed_unit")}`];
    if (r.speed.upload_mbps) parts.push(`↑ ${formatNumber(lang, r.speed.upload_mbps, 1)} ${t("speed_unit")}`);
    parts.push(`${formatNumber(lang, r.speed.latency_ms)} ${t("ms_unit")}`);
    return parts.join(" · ");
  }
  if (r.kind === "restored" || r.kind === "reminder") {
    return r.duration_seconds ? t("n_lasted", { duration: formatDuration(t, r.duration_seconds) }) : "";
  }
  if ((r.kind === "down" || r.kind === "degraded") && r.reason) {
    return describeCause(t, lang, {
      tcp: r.reason.tcp_ping_failed,
      http: r.reason.http_failed,
      dns: r.reason.dns_failed,
      loss: r.reason.packet_loss_pct,
      latency: r.reason.avg_latency_ms,
      type: r.kind,
    });
  }
  return "";
}

export function isUnread(r: NotificationRecord, lastReadAt: string | null): boolean {
  if (!lastReadAt) return true;
  return new Date(r.time).getTime() > new Date(lastReadAt).getTime();
}

const REFRESH_DELAY = 250;
const RETRY_DELAY = 5_000;

export function useNotifications(limit: number) {
  const revision = useLive((s) => s.notice.revision);
  const newest = useLive((s) => s.notice.last?.time ?? null);
  const [data, setData] = useState<NotificationList | null>(null);
  const [error, setError] = useState(false);
  const [nonce, setNonce] = useState(0);
  const seen = useRef<number | null>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: refetch on every live notification frame
  useEffect(() => {
    const ctrl = new AbortController();
    const load = () =>
      api.notifications(limit, ctrl.signal).then(
        (next) => {
          setData(next);
          setError(false);
        },
        () => {
          if (!ctrl.signal.aborted) setError(true);
        },
      );
    const live = seen.current !== null && seen.current !== revision;
    seen.current = revision;
    if (!live) {
      void load();
      return () => ctrl.abort();
    }
    const id = window.setTimeout(load, REFRESH_DELAY);
    return () => {
      window.clearTimeout(id);
      ctrl.abort();
    };
  }, [limit, revision, nonce]);

  const reload = useCallback(() => setNonce((n) => n + 1), []);

  // biome-ignore lint/correctness/useExhaustiveDependencies: schedule another retry after each failed attempt
  useEffect(() => {
    if (!error) return;
    const id = window.setTimeout(reload, RETRY_DELAY);
    return () => window.clearTimeout(id);
  }, [error, nonce, reload]);

  useEffect(() => {
    const onVisible = () => {
      if (document.visibilityState === "visible") reload();
    };
    document.addEventListener("visibilitychange", onVisible);
    return () => document.removeEventListener("visibilitychange", onVisible);
  }, [reload]);

  const markAllRead = useCallback(async (): Promise<boolean> => {
    if (!data) return false;
    const candidates = [data.items[0]?.time, newest].filter((v): v is string => !!v);
    const until = candidates.sort((a, b) => new Date(b).getTime() - new Date(a).getTime())[0];
    try {
      const r = await api.markNotificationsRead(until);
      setData((d) => (d ? { ...d, unread: r.unread, last_read_at: until ?? new Date().toISOString() } : d));
      reload();
      return true;
    } catch {
      return false;
    }
  }, [data, newest, reload]);

  return { data, error, reload, markAllRead };
}
