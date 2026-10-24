import {
  ArrowLeft,
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  FileText,
  Printer,
  RefreshCw,
} from "lucide-react";
import { type CSSProperties, type KeyboardEvent, useEffect, useId, useMemo, useRef, useState } from "react";
import { ChartTooltip } from "../components/ChartTooltip";
import { Banner, Button, Card, CardHead, Empty, Meter, Skeleton } from "../components/ui";
import { api } from "../lib/api";
import {
  describeReportCause,
  eventKey,
  formatDate,
  formatDuration,
  formatMonth,
  formatNumber,
  formatPct,
  formatTime,
  monthKey,
  shiftMonth,
} from "../lib/format";
import { useAsync } from "../lib/hooks";
import { isRtl, type MessageKey, useI18n } from "../lib/i18n";
import { navigate } from "../lib/router";
import type { MonthlyReport, ReportDay } from "../lib/types";
import { availabilityTone, statusTone, tone } from "../styles/categories";

function daysOf(month: string, days: ReportDay[]): (ReportDay | { date: string; empty: true })[] {
  const [y, m] = month.split("-").map(Number);
  if (!y || !m) return days;
  const count = new Date(y, m, 0).getDate();
  const byDate = new Map(days.map((d) => [d.date, d]));
  return Array.from({ length: count }, (_, i) => {
    const date = `${month}-${String(i + 1).padStart(2, "0")}`;
    return byDate.get(date) ?? { date, empty: true as const };
  });
}

function DailyChart({ report }: { report: MonthlyReport }) {
  const { t, lang } = useI18n();
  const wrap = useRef<HTMLDivElement>(null);
  const tipId = useId();
  const [active, setActive] = useState<{ i: number; x: number; y: number; w: number } | null>(null);
  const days = useMemo(() => daysOf(report.month, report.days ?? []), [report]);
  const [focusIdx, setFocusIdx] = useState(() => {
    const last = days.findLastIndex((d) => !("empty" in d));
    return Math.max(last, 0);
  });
  const rtl = isRtl(lang);

  const onBarKey = (e: KeyboardEvent<HTMLButtonElement>, i: number) => {
    const next = rtl ? "ArrowLeft" : "ArrowRight";
    const prev = rtl ? "ArrowRight" : "ArrowLeft";
    let to: number | null = null;
    if (e.key === next) to = Math.min(days.length - 1, i + 1);
    else if (e.key === prev) to = Math.max(0, i - 1);
    else if (e.key === "Home") to = 0;
    else if (e.key === "End") to = days.length - 1;
    if (to === null) return;
    e.preventDefault();
    setFocusIdx(to);
    const target = e.currentTarget.parentElement?.children[to];
    if (target instanceof HTMLElement) target.focus();
  };
  const max = Math.max(60, ...days.map((d) => ("empty" in d ? 0 : d.downtime_seconds)));
  const current = active !== null ? days[active.i] : undefined;

  const show = (i: number, el: HTMLElement) => {
    const box = wrap.current?.getBoundingClientRect();
    const fill = el.firstElementChild?.getBoundingClientRect() ?? el.getBoundingClientRect();
    if (!box) return;
    setActive({ i, x: fill.left + fill.width / 2 - box.left, y: fill.top - box.top, w: box.width });
  };

  const tipRows = (d: (typeof days)[number]) =>
    "empty" in d
      ? [{ label: t("report_uptime"), value: t("avail_none") }]
      : [
          {
            label: t("report_uptime"),
            value: formatPct(lang, d.uptime_pct),
            tone: availabilityTone(d.uptime_pct),
          },
          { label: t("report_outages"), value: formatNumber(lang, d.outages), tone: tone.disconnected },
          {
            label: t("report_downtime"),
            value: formatDuration(t, d.downtime_seconds),
            tone: tone.disconnected,
          },
        ];

  return (
    <div className="chart" ref={wrap}>
      <div
        className="bars"
        onPointerLeave={(e) => {
          if (e.pointerType === "mouse") setActive(null);
        }}
      >
        {days.map((d, i) => {
          const empty = "empty" in d;
          const down = empty ? 0 : d.downtime_seconds;
          const h = empty ? 2 : down > 0 ? Math.max(6, (down / max) * 100) : 3;
          const label = empty
            ? `${formatDate(lang, d.date)}: —`
            : `${formatDate(lang, d.date)}: ${formatPct(lang, d.uptime_pct)} · ${formatDuration(t, down)}`;
          return (
            <button
              type="button"
              key={d.date}
              className="bar"
              aria-label={label}
              tabIndex={focusIdx === i ? 0 : -1}
              onKeyDown={(e) => onBarKey(e, i)}
              data-active={active?.i === i || undefined}
              data-dim={(active !== null && active.i !== i) || undefined}
              onPointerEnter={(e) => show(i, e.currentTarget)}
              onFocus={(e) => {
                setFocusIdx(i);
                show(i, e.currentTarget);
              }}
              onBlur={() => setActive(null)}
              style={
                {
                  "--tone": empty ? "var(--color-track)" : down > 0 ? tone.disconnected : tone.connected,
                } as CSSProperties
              }
            >
              <span className="bar-fill" style={{ height: `${h}%` }} />
            </button>
          );
        })}
      </div>
      {current && active ? (
        <ChartTooltip
          id={tipId}
          x={active.x}
          y={active.y}
          bounds={active.w}
          title={formatDate(lang, current.date, { weekday: "long", day: "numeric", month: "long" })}
          rows={tipRows(current)}
        />
      ) : null}
      <div className="row" style={{ justifyContent: "space-between", marginTop: 8 }}>
        <span className="chart-axis num muted">
          {days[0] ? formatDate(lang, days[0].date, { day: "numeric", month: "short" }) : ""}
        </span>
        <span className="chart-axis num muted">
          {days.length
            ? formatDate(lang, days[days.length - 1]?.date ?? "", { day: "numeric", month: "short" })
            : ""}
        </span>
      </div>
      <div className="legend" style={{ marginTop: 8 }}>
        <span>
          <span className="dot" style={{ "--tone": tone.disconnected } as CSSProperties} />
          {t("report_downtime")}
        </span>
        <span>
          <span className="dot" style={{ "--tone": tone.connected } as CSSProperties} />
          {t("status_connected")}
        </span>
      </div>
    </div>
  );
}

export function Report({ month: initial }: { month: string | null }) {
  const { t, lang } = useI18n();
  const month = initial && /^\d{4}-\d{2}$/.test(initial) ? initial : monthKey(new Date());
  const report = useAsync((signal) => api.report(month, signal), [month]);
  const thisMonth = monthKey(new Date());
  const rtl = isRtl(lang);

  useEffect(() => {
    document.title = `${t("report_title")} — ${formatMonth(lang, month)}`;
    return () => {
      document.title = t("app_name");
    };
  }, [t, lang, month]);

  const go = (m: string) => navigate("report", `?month=${m}`, true);
  const r = report.data;
  const kpis: [MessageKey, string, string][] = r
    ? [
        ["report_uptime", formatPct(lang, r.uptime_pct), availabilityTone(r.uptime_pct)],
        ["report_outages", formatNumber(lang, r.disconnections), tone.disconnected],
        ["report_downtime", formatDuration(t, r.total_downtime_seconds), tone.disconnected],
        [
          "report_longest",
          r.longest_outage_seconds ? formatDuration(t, r.longest_outage_seconds) : "—",
          tone.degraded,
        ],
        ["report_mttr", r.mttr_seconds ? formatDuration(t, r.mttr_seconds) : "—", tone.latency],
        ["report_degraded", formatDuration(t, r.total_degraded_seconds), tone.degraded],
      ]
    : [];
  const causes = r
    ? Object.entries(r.causes ?? {})
        .filter(([, c]) => c.outages > 0)
        .sort((a, b) => b[1].outages - a[1].outages)
    : [];

  return (
    <div className="report">
      <div className="report-head no-print">
        <Button variant="ghost" icon={rtl ? ArrowRight : ArrowLeft} onClick={() => navigate("history")}>
          {t("report_back")}
        </Button>
        <div className="row">
          <Button
            variant="edge"
            size="sm"
            iconOnly
            icon={rtl ? ChevronRight : ChevronLeft}
            aria-label={t("report_prev")}
            onClick={() => go(shiftMonth(month, -1))}
          />
          <strong className="num" style={{ minWidth: 140, textAlign: "center", fontWeight: 500 }}>
            {formatMonth(lang, month)}
          </strong>
          <Button
            variant="edge"
            size="sm"
            iconOnly
            icon={rtl ? ChevronLeft : ChevronRight}
            aria-label={t("report_next")}
            disabled={month >= thisMonth}
            onClick={() => go(shiftMonth(month, 1))}
          />
        </div>
        <Button variant="solid" icon={Printer} onClick={() => window.print()} disabled={!r}>
          {t("report_print")}
        </Button>
      </div>

      <header className="stack" style={{ gap: 4 }}>
        <div className="row">
          <img src="/favicon.png" alt="" width={28} height={28} />
          <span className="muted">{t("app_name")}</span>
        </div>
        <h1>
          {t("report_title")} — {formatMonth(lang, month)}
        </h1>
        {r ? (
          <p className="muted num">
            {t("report_generated", {
              date: formatDate(lang, r.generated_at, { dateStyle: "medium", timeStyle: "short" }),
            })}{" "}
            · {t("report_monitored", { duration: formatDuration(t, r.monitored_seconds) })}
          </p>
        ) : null}
      </header>

      {report.error ? (
        <Banner
          tone="danger"
          icon={CircleAlert}
          actions={
            <Button size="sm" icon={RefreshCw} onClick={report.reload}>
              {t("retry")}
            </Button>
          }
        >
          {t("load_failed")}
        </Banner>
      ) : !r ? (
        <div className="stack" aria-busy="true">
          <Skeleton height={110} />
          <Skeleton height={200} />
        </div>
      ) : r.monitored_seconds <= 0 && !r.events?.length ? (
        <Card>
          <Empty icon={FileText}>{t("report_no_data")}</Empty>
        </Card>
      ) : (
        <>
          <div className="kpis">
            {kpis.map(([key, value, flare]) => (
              <Card key={key} className="stat" flare={flare} as="div">
                <span className="stat-label">{t(key)}</span>
                <span className="stat-value">{value}</span>
              </Card>
            ))}
          </div>
          {r.carried_over_seconds > 0 ? (
            <p className="muted">
              {t("report_carried", { duration: formatDuration(t, r.carried_over_seconds) })}
            </p>
          ) : null}

          <Card labelledBy="daily-title">
            <CardHead id="daily-title" title={t("report_daily")} />
            <DailyChart report={r} />
          </Card>

          {causes.length ? (
            <Card labelledBy="causes-title">
              <CardHead id="causes-title" title={t("report_causes")} />
              <div className="stack" style={{ gap: 16 }}>
                {causes.map(([layer, c]) => (
                  <Meter
                    key={layer}
                    label={`${describeReportCause(t, layer)} · ${formatNumber(lang, c.outages)} · ${formatDuration(t, c.downtime_seconds)}`}
                    value={c.pct}
                    display={formatPct(lang, c.pct)}
                    tone={tone.disconnected}
                    tipTitle={describeReportCause(t, layer)}
                    tipRows={[
                      {
                        label: t("report_outages"),
                        value: formatNumber(lang, c.outages),
                        tone: tone.disconnected,
                      },
                      { label: t("report_downtime"), value: formatDuration(t, c.downtime_seconds) },
                      { label: t("report_share"), value: formatPct(lang, c.pct) },
                    ]}
                  />
                ))}
              </div>
            </Card>
          ) : null}

          <Card labelledBy="report-events">
            <CardHead
              id="report-events"
              title={t("report_events")}
              sub={formatNumber(lang, r.events?.length ?? 0)}
            />
            {!r.events?.length ? (
              <Empty icon={FileText}>{t("events_empty")}</Empty>
            ) : (
              <div className="table-wrap">
                <table className="table">
                  <thead>
                    <tr>
                      <th scope="col">{t("report_day")}</th>
                      <th scope="col">{t("col_time")}</th>
                      <th scope="col">{t("col_event")}</th>
                      <th scope="col">{t("col_duration")}</th>
                      <th scope="col">{t("col_cause")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {r.events.map((e) => (
                      <tr key={`${e.time}-${e.type}`}>
                        <td>{formatDate(lang, e.time)}</td>
                        <td className="num">{formatTime(lang, e.time, false)}</td>
                        <td>
                          <span className="row" style={{ gap: 8, flexWrap: "nowrap" }}>
                            <span
                              className="dot"
                              style={{ "--tone": statusTone(e.type) } as CSSProperties}
                              aria-hidden
                            />
                            {t(eventKey(e.type))}
                          </span>
                        </td>
                        <td className="num">{formatDuration(t, e.duration_seconds)}</td>
                        <td className="wrap secondary">{describeReportCause(t, e.cause)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        </>
      )}
    </div>
  );
}
