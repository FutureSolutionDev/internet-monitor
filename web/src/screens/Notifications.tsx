import { BellOff, Check, CheckCheck, CircleAlert, RefreshCw } from "lucide-react";
import { type CSSProperties, useMemo, useState } from "react";
import { SmartSelect } from "../components/SmartSelect";
import { toast } from "../components/Toast";
import { Banner, Button, Card, Empty, Segmented, Skeleton } from "../components/ui";
import { formatDate, formatTime, localDateKey } from "../lib/format";
import { type Lang, type Translate, useI18n } from "../lib/i18n";
import {
  CHANNEL_NAME,
  isUnread,
  KIND_LABEL,
  KIND_TONE,
  mutedChannels,
  noticeDetail,
  REASON_LABEL,
  sentChannels,
  useNotifications,
} from "../lib/notifications";
import type { AlertKind, NotificationRecord } from "../lib/types";

const LIMIT = 500;
const PAGE = 50;

type KindFilter = "all" | AlertKind;
type StateFilter = "all" | "sent" | "held";

const KINDS: AlertKind[] = ["down", "degraded", "restored", "reminder", "speed", "test"];

function dayLabel(t: Translate, lang: Lang, key: string, today: string, yesterday: string) {
  if (key === today) return t("n_today");
  if (key === yesterday) return t("n_yesterday");
  return formatDate(lang, key, { weekday: "long", day: "numeric", month: "long", year: "numeric" });
}

export function Notifications() {
  const { t, lang } = useI18n();
  const { data, error, reload, markAllRead } = useNotifications(LIMIT);
  const [kind, setKind] = useState<KindFilter>("all");
  const [state, setState] = useState<StateFilter>("all");
  const [shown, setShown] = useState(PAGE);
  const [busy, setBusy] = useState(false);

  const filtered = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((r) => {
      if (kind !== "all" && r.kind !== kind) return false;
      if (state === "sent") return r.sent.length > 0;
      if (state === "held") return mutedChannels(r).length > 0;
      return true;
    });
  }, [data, kind, state]);

  const groups = useMemo(() => {
    const out: { key: string; items: NotificationRecord[] }[] = [];
    for (const r of filtered.slice(0, shown)) {
      const key = localDateKey(new Date(r.time));
      const last = out[out.length - 1];
      if (last && last.key === key) last.items.push(r);
      else out.push({ key, items: [r] });
    }
    return out;
  }, [filtered, shown]);

  const now = new Date();
  const today = localDateKey(now);
  const yesterday = localDateKey(new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1));
  const unread = data?.unread ?? 0;

  const readAll = async () => {
    setBusy(true);
    const ok = await markAllRead();
    setBusy(false);
    if (!ok) toast(t("n_mark_failed"), "error");
  };

  const kindOptions = [
    { value: "all" as KindFilter, label: t("n_filter_all_kinds") },
    ...KINDS.map((k) => ({ value: k as KindFilter, label: t(KIND_LABEL[k]) })),
  ];

  return (
    <>
      <div className="page-head">
        <div>
          <h1>{t("n_page_title")}</h1>
          <p>{t("n_page_sub")}</p>
        </div>
        {data && data.items.length > 0 ? (
          <Button variant="edge" icon={CheckCheck} busy={busy} disabled={unread === 0} onClick={readAll}>
            {t("bell_mark_read")}
          </Button>
        ) : null}
      </div>

      {error ? (
        <Banner
          tone="danger"
          icon={CircleAlert}
          actions={
            <Button size="sm" icon={RefreshCw} onClick={reload}>
              {t("retry")}
            </Button>
          }
        >
          {t("load_failed")}
        </Banner>
      ) : null}

      {data && data.items.length > 0 ? (
        <div className="row notice-filters">
          <SmartSelect
            variant="pill"
            label={t("n_filter_kind")}
            value={kind}
            options={kindOptions}
            onChange={(v) => {
              setKind(v);
              setShown(PAGE);
            }}
          />
          <Segmented<StateFilter>
            label={t("n_filter_state")}
            value={state}
            onChange={(v) => {
              setState(v);
              setShown(PAGE);
            }}
            options={[
              { value: "all", label: t("n_filter_all") },
              { value: "sent", label: t("n_filter_sent") },
              { value: "held", label: t("n_filter_held") },
            ]}
          />
        </div>
      ) : null}

      {!data && !error ? (
        <div className="stack" aria-busy="true">
          <Skeleton height={72} />
          <Skeleton height={72} />
          <Skeleton height={72} />
        </div>
      ) : null}

      {data && filtered.length === 0 ? (
        <Card>
          <Empty icon={BellOff}>{data.items.length === 0 ? t("bell_empty") : t("n_filter_empty")}</Empty>
        </Card>
      ) : null}

      {groups.map((g) => (
        <section key={g.key} className="notice-day" aria-labelledby={`day-${g.key}`}>
          <h2 id={`day-${g.key}`} className="notice-day-title">
            {dayLabel(t, lang, g.key, today, yesterday)}
          </h2>
          <Card as="div">
            <ul className="notice-list">
              {g.items.map((r) => {
                const detail = noticeDetail(t, lang, r);
                const muted = mutedChannels(r);
                const sent = sentChannels(r);
                const fresh = isUnread(r, data?.last_read_at ?? null);
                return (
                  <li key={r.id} className="notice" data-unread={fresh || undefined}>
                    <span
                      className="notice-mark"
                      style={{ "--tone": KIND_TONE[r.kind] } as CSSProperties}
                      aria-hidden
                    />
                    <div className="notice-main">
                      <div className="notice-head">
                        <strong>{t(KIND_LABEL[r.kind])}</strong>
                        {fresh ? <span className="notice-new">{t("n_unread")}</span> : null}
                        <time className="notice-time num" dateTime={r.time} title={formatDate(lang, r.time)}>
                          {formatTime(lang, r.time, false)}
                        </time>
                      </div>
                      {detail ? <p className="notice-detail">{detail}</p> : null}
                      <div className="notice-channels">
                        {sent.length === 0 && muted.length === 0 ? (
                          <span className="notice-tag">{t("n_not_sent")}</span>
                        ) : null}
                        {sent.map((c) => (
                          <span key={c} className="notice-tag" data-sent>
                            <Check size={12} aria-hidden />
                            <span className="sr-only">{t("n_sent_to")} </span>
                            {t(CHANNEL_NAME[c])}
                          </span>
                        ))}
                        {muted.map(({ channel, reason }) => (
                          <span key={channel} className="notice-tag">
                            <BellOff size={12} aria-hidden />
                            <span className="sr-only">{t("n_held")} </span>
                            {t(CHANNEL_NAME[channel])} · {t(REASON_LABEL[reason])}
                          </span>
                        ))}
                      </div>
                    </div>
                  </li>
                );
              })}
            </ul>
          </Card>
        </section>
      ))}

      {data && data.items.length >= LIMIT && filtered.length <= shown ? (
        <p className="notice-cap">{t("n_cap", { n: LIMIT })}</p>
      ) : null}

      {data ? (
        <p className="sr-only" role="status">
          {t("n_showing", { n: Math.min(shown, filtered.length), total: filtered.length })}
        </p>
      ) : null}

      {filtered.length > shown ? (
        <div className="row notice-more">
          <Button variant="ghost" onClick={() => setShown((n) => n + PAGE)}>
            {t("n_show_more")}
          </Button>
        </div>
      ) : null}
    </>
  );
}
