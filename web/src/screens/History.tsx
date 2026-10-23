import {
  CalendarDays,
  CircleAlert,
  Download,
  FileText,
  History as HistoryIcon,
  RefreshCw,
} from "lucide-react";
import { type CSSProperties, useMemo, useState } from "react";
import { toast } from "../components/Toast";
import { Banner, Button, Card, CardHead, Empty, Segmented, Skeleton } from "../components/ui";
import { api } from "../lib/api";
import { downloadText, toCSV } from "../lib/csv";
import {
  causeOfLog,
  describeCause,
  eventKey,
  formatDate,
  formatDuration,
  formatNumber,
  formatTime,
  localDateKey,
  monthKey,
} from "../lib/format";
import { buildRows, summarize } from "../lib/history";
import { useAsync } from "../lib/hooks";
import { useI18n } from "../lib/i18n";
import { useLive } from "../lib/live";
import { hrefOf, navigate } from "../lib/router";
import type { LogEvent } from "../lib/types";
import { statusTone } from "../styles/categories";

type Filter = "all" | "disconnected" | "degraded" | "connected";

export function History() {
  const { t, lang } = useI18n();
  const today = localDateKey(new Date());
  const eventCount = useLive((s) => s.snapshot?.events?.length ?? 0);
  const dates = useAsync(() => api.logDates(), []);
  const [picked, setPicked] = useState<string | null>(null);
  const [filter, setFilter] = useState<Filter>("all");
  const [exporting, setExporting] = useState<number | null>(null);

  const available = dates.data ?? [];
  const date = picked ?? available[0] ?? today;
  const logs = useAsync((signal) => api.logs(date, 5000, signal), [date, date === today ? eventCount : 0]);

  const rows = useMemo(
    () => buildRows(logs.data ?? [], Date.now(), date === today),
    [logs.data, date, today],
  );
  const summary = useMemo(() => summarize(rows), [rows]);
  const visible = filter === "all" ? rows : rows.filter((r) => r.event.event === filter);

  const exportDays = async (days: number) => {
    setExporting(days);
    try {
      const cutoff = localDateKey(new Date(Date.now() - (days - 1) * 86400000));
      const wanted = days === 1 ? [date] : available.filter((d) => d >= cutoff);
      let partial = false;
      const all: LogEvent[] = [];
      for (const d of wanted) {
        try {
          all.push(...(await api.logs(d, 5000)));
        } catch {
          partial = true;
        }
      }
      const built = buildRows(all, Date.now(), wanted.includes(today)).reverse();
      const csv = toCSV([
        [
          "timestamp",
          "local_time",
          "event",
          "duration_seconds",
          "tcp_failed",
          "http_failed",
          "dns_failed",
          "packet_loss_pct",
          "latency_ms",
        ],
        ...built.map((r) => [
          r.event.timestamp,
          new Date(r.ts).toLocaleString(),
          r.event.event,
          r.duration === null ? "" : Math.round(r.duration),
          String(r.event.reason?.tcp_ping_failed ?? false),
          String(r.event.reason?.http_failed ?? false),
          String(r.event.reason?.dns_failed ?? false),
          r.event.reason?.packet_loss_pct ?? 0,
          r.event.reason?.avg_latency_ms ?? 0,
        ]),
      ]);
      downloadText(
        days === 1 ? `internet-monitor_${date}.csv` : `internet-monitor_last-${days}-days.csv`,
        csv,
      );
      toast(partial ? t("export_partial") : t("export_done", { n: built.length }), partial ? "error" : "ok");
    } finally {
      setExporting(null);
    }
  };

  return (
    <>
      <div className="page-head">
        <div>
          <h1>{t("history_title")}</h1>
          <p>{t("history_sub")}</p>
        </div>
        <div className="row">
          <a
            className="btn btn-edge"
            href={hrefOf("report", `?month=${monthKey(new Date())}`)}
            onClick={(e) => {
              e.preventDefault();
              navigate("report", `?month=${date.slice(0, 7)}`);
            }}
          >
            <FileText size={17} aria-hidden />
            {t("report_open")}
          </a>
        </div>
      </div>

      <Card>
        <div className="row" style={{ gap: 12 }}>
          <label className="field" style={{ minWidth: 220 }}>
            <span className="field-label">{t("history_date")}</span>
            <select
              className="input"
              value={date}
              onChange={(e) => setPicked(e.target.value)}
              disabled={!available.length}
            >
              {available.length === 0 ? <option value={today}>{t("history_no_dates")}</option> : null}
              {available.map((d) => (
                <option key={d} value={d}>
                  {formatDate(lang, d, { weekday: "long", day: "numeric", month: "long", year: "numeric" })}
                </option>
              ))}
            </select>
          </label>
          <div className="field">
            <span className="field-label">{t("history_filter")}</span>
            <Segmented<Filter>
              label={t("history_filter")}
              value={filter}
              onChange={setFilter}
              options={[
                { value: "all", label: t("filter_all") },
                { value: "disconnected", label: t("filter_disconnected") },
                { value: "degraded", label: t("filter_degraded") },
                { value: "connected", label: t("filter_connected") },
              ]}
            />
          </div>
          <span className="spacer" />
          <div className="row">
            <Button
              variant="edge"
              size="sm"
              icon={Download}
              busy={exporting === 1}
              onClick={() => exportDays(1)}
            >
              {t("export_day")}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              busy={exporting === 7}
              onClick={() => exportDays(7)}
              disabled={!available.length}
            >
              {t("export_7")}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              busy={exporting === 30}
              onClick={() => exportDays(30)}
              disabled={!available.length}
            >
              {t("export_30")}
            </Button>
          </div>
        </div>
      </Card>

      <div className="grid grid-stats">
        {(
          [
            ["history_outages", formatNumber(lang, summary.outages), "var(--accent-coral)"],
            [
              "history_downtime",
              summary.downtime ? formatDuration(t, summary.downtime) : "—",
              "var(--accent-coral)",
            ],
            ["history_unstable", formatNumber(lang, summary.unstable), "var(--accent-amber)"],
            [
              "history_longest",
              summary.longest ? formatDuration(t, summary.longest) : "—",
              "var(--accent-blue)",
            ],
          ] as const
        ).map(([key, value, flare]) => (
          <Card key={key} className="stat" flare={flare} as="div">
            <span className="stat-label">{t(key)}</span>
            <span className="stat-value">{logs.loading && !logs.data ? "…" : value}</span>
          </Card>
        ))}
      </div>

      <Card labelledBy="log-title">
        <CardHead
          id="log-title"
          title={formatDate(lang, date, { weekday: "long", day: "numeric", month: "long" })}
          sub={
            logs.data
              ? `${formatNumber(lang, visible.length)} / ${formatNumber(lang, rows.length)}`
              : undefined
          }
        />
        {logs.error ? (
          <Banner
            tone="danger"
            icon={CircleAlert}
            actions={
              <Button size="sm" icon={RefreshCw} onClick={logs.reload}>
                {t("retry")}
              </Button>
            }
          >
            {t("load_failed")}
          </Banner>
        ) : logs.loading && !logs.data ? (
          <div className="stack">
            <Skeleton height={44} />
            <Skeleton height={44} />
            <Skeleton height={44} />
          </div>
        ) : rows.length === 0 ? (
          <Empty icon={available.length ? CalendarDays : HistoryIcon}>{t("history_empty")}</Empty>
        ) : visible.length === 0 ? (
          <Empty icon={HistoryIcon}>{t("history_empty_filter")}</Empty>
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th scope="col">{t("col_time")}</th>
                  <th scope="col">{t("col_event")}</th>
                  <th scope="col">{t("col_duration")}</th>
                  <th scope="col">{t("col_cause")}</th>
                </tr>
              </thead>
              <tbody>
                {visible.map((r) => (
                  <tr key={`${r.ts}-${r.event.event}`}>
                    <td className="num">{formatTime(lang, r.ts)}</td>
                    <td>
                      <span className="row" style={{ gap: 8, flexWrap: "nowrap" }}>
                        <span
                          className="dot"
                          style={{ "--tone": statusTone(r.event.event) } as CSSProperties}
                          aria-hidden
                        />
                        {t(eventKey(r.event.event))}
                      </span>
                    </td>
                    <td className="num">
                      {r.duration === null ? "—" : formatDuration(t, r.duration)}
                      {r.ongoing ? " …" : ""}
                    </td>
                    <td className="wrap secondary">{describeCause(t, lang, causeOfLog(r.event))}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </>
  );
}
