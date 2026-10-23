import {
  Activity,
  CircleAlert,
  CircleCheck,
  CircleX,
  Clock,
  Globe,
  type LucideIcon,
  Radar,
  RefreshCw,
  ShieldAlert,
  TriangleAlert,
  Waves,
  Wifi,
  WifiOff,
} from "lucide-react";
import { type CSSProperties, useState } from "react";
import { LatencyChart } from "../components/LatencyChart";
import { Banner, Button, Card, CardHead, Empty, Ltr, Meter, Skeleton } from "../components/ui";
import {
  causeOfEntry,
  describeCause,
  eventKey,
  formatAgo,
  formatDuration,
  formatNumber,
  formatPct,
  formatTime,
  statusKey,
} from "../lib/format";
import { useNow } from "../lib/hooks";
import { useI18n } from "../lib/i18n";
import { useLive } from "../lib/live";
import { hrefOf, navigate } from "../lib/router";
import type { MessageKey } from "../lib/strings";
import type { Diagnosis, Snapshot, Status } from "../lib/types";
import { availabilityTone, statusTone, tone } from "../styles/categories";

const STATUS_ICON: Record<Status, LucideIcon> = {
  connected: Wifi,
  degraded: TriangleAlert,
  disconnected: WifiOff,
  checking: Radar,
};

const DIAG_KEY: Record<Diagnosis, MessageKey> = {
  ok: "diag_ok",
  checking: "diag_checking",
  down: "diag_down",
  dns: "diag_dns",
  http: "diag_http",
  partial: "diag_partial",
  lan: "diag_lan",
  isp: "diag_isp",
};

function Hero({ snap }: { snap: Snapshot }) {
  const { t } = useI18n();
  const now = useNow();
  const status = snap.status;
  const Icon = STATUS_ICON[status] ?? Radar;
  const since = Math.max(0, (now - snap.status_since) / 1000);
  return (
    <Card className={`hero tone-${status}`} flare={statusTone(status)} labelledBy="hero-status">
      <div className={`hero-orb ${status === "disconnected" || status === "checking" ? "pulse" : ""}`}>
        <Icon size={34} aria-hidden />
      </div>
      <div>
        <h1 id="hero-status" className="hero-status" aria-live="polite">
          {t(statusKey(status))}
        </h1>
        <p className="hero-sub">{t(DIAG_KEY[snap.diagnosis] ?? "diag_partial")}</p>
      </div>
      <div className="hero-meta">
        {status !== "checking" ? (
          <span className="num">{t("status_since", { duration: formatDuration(t, since) })}</span>
        ) : null}
        <span className="num">
          {snap.last_check
            ? t("last_check", { ago: formatAgo(t, snap.last_check, now) })
            : t("status_waiting")}
        </span>
        <span>{t("every_n_seconds", { n: snap.check_interval_sec })}</span>
        {snap.disconnections > 0 ? (
          <span className="num">
            {t("stat_drops")}: {snap.disconnections}
          </span>
        ) : null}
      </div>
    </Card>
  );
}

function Stat({
  icon: Icon,
  label,
  value,
  unit,
  foot,
  flare,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
  unit?: string;
  foot?: string;
  flare: string;
}) {
  return (
    <Card className="stat" flare={flare} as="div">
      <span className="stat-label">
        <Icon size={16} aria-hidden />
        {label}
      </span>
      <span className="stat-value">
        {value}
        {unit ? <small>{unit}</small> : null}
      </span>
      {foot ? <span className="stat-foot num">{foot}</span> : null}
    </Card>
  );
}

function Stats({ snap }: { snap: Snapshot }) {
  const { t, lang } = useI18n();
  const ms = t("ms_unit");
  const live = snap.total_checks > 0;
  return (
    <div className="grid grid-stats">
      <Stat
        icon={Activity}
        label={t("stat_latency")}
        value={live && snap.latency_ms > 0 ? formatNumber(lang, snap.latency_ms) : "—"}
        unit={ms}
        foot={
          snap.avg_latency_ms
            ? t("stat_avg", { value: `${formatNumber(lang, snap.avg_latency_ms)} ${ms}` })
            : undefined
        }
        flare={tone.latency}
      />
      <Stat
        icon={Waves}
        label={t("stat_jitter")}
        value={live ? formatNumber(lang, snap.jitter_ms) : "—"}
        unit={ms}
        flare={tone.jitter}
      />
      <Stat
        icon={ShieldAlert}
        label={t("stat_loss")}
        value={live ? formatNumber(lang, snap.packet_loss, snap.packet_loss % 1 ? 1 : 0) : "—"}
        unit="%"
        flare={tone.loss}
      />
      <Stat
        icon={Clock}
        label={t("stat_uptime")}
        value={live ? formatPct(lang, snap.uptime_pct) : "—"}
        foot={t("stat_checks", {
          up: formatNumber(lang, snap.up_checks),
          total: formatNumber(lang, snap.total_checks),
        })}
        flare={tone.uptime}
      />
    </div>
  );
}

function Availability({ snap }: { snap: Snapshot }) {
  const { t, lang } = useI18n();
  const a = snap.availability;
  const rows: [MessageKey, number | null | undefined][] = [
    ["avail_today", a?.today],
    ["avail_week", a?.week],
    ["avail_month", a?.month],
  ];
  return (
    <Card labelledBy="avail-title">
      <CardHead id="avail-title" title={t("avail_title")} />
      <div className="stack" style={{ gap: 16 }}>
        {rows.map(([key, v]) => (
          <Meter
            key={key}
            label={t(key)}
            value={v ?? null}
            display={v === null || v === undefined ? t("avail_none") : formatPct(lang, v)}
            tone={availabilityTone(v)}
          />
        ))}
      </div>
    </Card>
  );
}

function Targets({ snap }: { snap: Snapshot }) {
  const { t, lang } = useI18n();
  const targets = snap.targets ?? [];
  return (
    <Card labelledBy="targets-title">
      <CardHead id="targets-title" title={t("targets_title")} />
      {targets.length === 0 ? (
        <Empty icon={Globe}>{t("targets_empty")}</Empty>
      ) : (
        <ul className="list">
          {targets.map((tg) => (
            <li key={`${tg.layer}:${tg.target}`} className="list-row">
              <span className="tag">{tg.layer.toUpperCase()}</span>
              <span className="list-row-main">
                <Ltr className="list-row-title">{tg.target}</Ltr>
              </span>
              <span
                className="list-row-end num"
                style={{ color: tg.ok ? "var(--text-secondary)" : "var(--accent-coral)" }}
              >
                {tg.ok ? (
                  <>
                    <CircleCheck
                      size={14}
                      aria-hidden
                      style={{ verticalAlign: "-2px", color: tone.connected }}
                    />{" "}
                    {tg.latency_ms > 0
                      ? `${formatNumber(lang, tg.latency_ms)} ${t("ms_unit")}`
                      : t("target_ok")}
                  </>
                ) : (
                  <>
                    <CircleX size={14} aria-hidden style={{ verticalAlign: "-2px" }} /> {t("target_fail")}
                  </>
                )}
              </span>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}

function Events({ snap }: { snap: Snapshot }) {
  const { t, lang } = useI18n();
  const events = (snap.events ?? []).slice(0, 8);
  return (
    <Card labelledBy="events-title">
      <CardHead id="events-title" title={t("events_title")}>
        <a
          className="btn btn-ghost btn-sm"
          href={hrefOf("history")}
          onClick={(e) => {
            e.preventDefault();
            navigate("history");
          }}
        >
          {t("events_view_all")}
        </a>
      </CardHead>
      {events.length === 0 ? (
        <Empty icon={CircleCheck}>{t("events_empty")}</Empty>
      ) : (
        <ul className="list">
          {events.map((e) => {
            const Icon = STATUS_ICON[e.event_type] ?? Radar;
            return (
              <li key={`${e.ts}-${e.event_type}`} className="list-row">
                <span className="icon-well" style={{ "--flare": statusTone(e.event_type) } as CSSProperties}>
                  <Icon size={16} aria-hidden />
                </span>
                <span className="list-row-main">
                  <span className="list-row-title">{t(eventKey(e.event_type))}</span>
                  <span className="list-row-sub">{describeCause(t, lang, causeOfEntry(e))}</span>
                </span>
                <span className="list-row-end">
                  <span className="num">{formatTime(lang, e.ts)}</span>
                  {e.event_type === "connected" && e.duration_seconds > 0 ? (
                    <>
                      <br />
                      <span className="muted">
                        {t("ev_lasted", { duration: formatDuration(t, e.duration_seconds) })}
                      </span>
                    </>
                  ) : null}
                </span>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

function LoadingOverview() {
  const { t } = useI18n();
  const [started] = useState(() => Date.now());
  const now = useNow();
  if (now - started > 8000) {
    return (
      <Banner
        tone="warn"
        icon={CircleAlert}
        actions={
          <Button size="sm" icon={RefreshCw} onClick={() => window.location.reload()}>
            {t("retry")}
          </Button>
        }
      >
        {t("gate_offline_body")}
      </Banner>
    );
  }
  return (
    <div className="stack" aria-busy="true">
      <Skeleton height={124} />
      <div className="grid grid-stats">
        <Skeleton height={112} />
        <Skeleton height={112} />
        <Skeleton height={112} />
        <Skeleton height={112} />
      </div>
      <Skeleton height={260} />
    </div>
  );
}

export function Overview() {
  const { t, lang } = useI18n();
  const snap = useLive((s) => s.snapshot);
  if (!snap) return <LoadingOverview />;
  const history = snap.latency_history ?? [];
  return (
    <>
      <Hero snap={snap} />
      <Stats snap={snap} />
      <div className="grid grid-overview">
        <div className="stack">
          <Card labelledBy="chart-title" flare={tone.latency}>
            <CardHead
              id="chart-title"
              title={t("chart_latency")}
              sub={t("chart_latency_sub", { n: history.length })}
            />
            {history.length === 0 ? (
              <Empty icon={Activity}>{t("chart_empty")}</Empty>
            ) : (
              <LatencyChart
                values={history}
                lastCheck={snap.last_check || Date.now()}
                intervalSec={snap.check_interval_sec}
                lang={lang}
                t={t}
              />
            )}
          </Card>
          <Events snap={snap} />
        </div>
        <div className="stack">
          <Availability snap={snap} />
          <Targets snap={snap} />
        </div>
      </div>
    </>
  );
}
