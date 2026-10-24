import {
  ArrowDownToLine,
  ArrowUpFromLine,
  CircleAlert,
  Gauge as GaugeIcon,
  type LucideIcon,
  Play,
  RefreshCw,
  Square,
  Timer,
} from "lucide-react";
import { type CSSProperties, useEffect, useState } from "react";
import { Gauge } from "../components/Gauge";
import { toast } from "../components/Toast";
import { Banner, Button, Card, CardHead, Empty, Skeleton } from "../components/ui";
import { ApiError, api } from "../lib/api";
import { speedErrorKey } from "../lib/errors";
import { formatNumber, formatTime, latencyParts, localDateKey } from "../lib/format";
import { useAsync } from "../lib/hooks";
import { type MessageKey, useI18n } from "../lib/i18n";
import { live, useLive } from "../lib/live";
import type { SpeedProgress } from "../lib/types";
import { tone } from "../styles/categories";

const PHASE_KEY: Record<NonNullable<SpeedProgress["phase"]>, MessageKey> = {
  ping: "phase_ping",
  download: "phase_download",
  upload: "phase_upload",
  done: "phase_done",
  error: "phase_error",
  cancelled: "phase_cancelled",
};

function Phase({
  icon: Icon,
  label,
  value,
  unit,
  flare,
  active,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
  unit: string;
  flare: string;
  active: boolean;
}) {
  return (
    <div className="phase" data-active={active} style={{ "--flare": flare } as CSSProperties}>
      <span className="icon-well">
        <Icon size={17} aria-hidden />
      </span>
      <span className="stack" style={{ gap: 0 }}>
        <span className="muted" style={{ fontSize: "var(--text-xs)" }}>
          {label}
        </span>
        <span className="phase-value">
          {value} <small className="muted">{unit}</small>
        </span>
      </span>
    </div>
  );
}

export function Speed() {
  const { t, lang } = useI18n();
  const speed = useLive((s) => s.speed);
  const running = useLive((s) => s.snapshot?.speed_test_running ?? false);
  const status = useLive((s) => s.snapshot?.status);
  const [starting, setStarting] = useState(false);
  const today = localDateKey(new Date());
  const history = useAsync((signal) => api.speedHistory(today, signal), [today]);
  const config = useAsync(() => api.config(), []);

  const busy = running || (speed !== null && !speed.done);
  const doneKey = speed?.done ? speed.phase : undefined;

  // biome-ignore lint/correctness/useExhaustiveDependencies: reload only when a test finishes
  useEffect(() => {
    if (doneKey === "done") history.reload();
  }, [doneKey]);

  const start = async () => {
    setStarting(true);
    live.beginSpeedTest();
    try {
      await api.startSpeedTest();
    } catch (e) {
      if (e instanceof ApiError && e.code === "test_already_running") toast(t("speed_running_elsewhere"));
      else {
        live.abortSpeedTest();
        toast(t("phase_error"), "error");
      }
    } finally {
      setStarting(false);
    }
  };

  const cancel = async () => {
    try {
      await api.cancelSpeedTest();
    } catch {
      toast(t("phase_error"), "error");
    }
  };

  const phase = speed?.phase;
  const result = speed?.result;
  const current = busy ? (speed?.current_mbps ?? 0) : (result?.download_mbps ?? speed?.download_mbps ?? 0);
  const gaugeTone = phase === "upload" ? tone.upload : tone.download;
  const uploadOff = config.data ? !config.data.config.speed_test.upload_target : false;
  const fmt = (v: number | null | undefined) => (v && v > 0 ? formatNumber(lang, v, v < 10 ? 1 : 0) : "—");

  const download =
    result?.download_mbps ?? speed?.download_mbps ?? (phase === "download" ? speed?.current_mbps : undefined);
  const upload = result?.upload_mbps ?? (phase === "upload" ? speed?.current_mbps : undefined);
  const latency = result?.latency_ms ?? speed?.latency_ms;
  const statusText = busy
    ? phase
      ? t(PHASE_KEY[phase])
      : t("speed_starting")
    : phase
      ? t(PHASE_KEY[phase])
      : t("speed_idle");

  const alert = config.data?.config.speed_test.alert_threshold_mbps ?? 0;
  const schedule = config.data?.config.speed_test.schedule_minutes ?? 0;

  return (
    <>
      <div className="page-head">
        <div>
          <h1>{t("speed_title")}</h1>
          <p>
            {t("speed_sub")}
            {schedule > 0 ? ` · ${t("speed_schedule_note", { n: schedule })}` : ""}
          </p>
        </div>
      </div>

      {status === "disconnected" && !busy ? (
        <Banner tone="warn" icon={CircleAlert}>
          {t("speed_offline")}
        </Banner>
      ) : null}

      <Card flare={gaugeTone} labelledBy="speed-status">
        <div className="stack" style={{ alignItems: "center", gap: 20 }}>
          <Gauge
            value={current}
            display={fmt(current)}
            unit={t("speed_unit")}
            label={phase === "upload" ? t("speed_upload") : t("speed_download")}
            tone={gaugeTone}
          />
          <p id="speed-status" className="secondary" role="status" aria-live="polite">
            {statusText}
            {phase === "error" ? ` — ${t(speedErrorKey(speed?.error))}` : ""}
          </p>
          {busy ? (
            <Button variant="edge" size="lg" icon={Square} onClick={cancel}>
              {t("speed_cancel")}
            </Button>
          ) : (
            <Button variant="aurora" size="lg" icon={Play} busy={starting} onClick={start}>
              {t("speed_start")}
            </Button>
          )}
          {!busy && result && alert > 0 && result.download_mbps < alert ? (
            <Banner tone="warn" icon={CircleAlert}>
              {t("speed_alert", { value: formatNumber(lang, alert) })}
            </Banner>
          ) : null}
        </div>
      </Card>

      <div className="phases">
        <Phase
          icon={Timer}
          label={t("speed_ping")}
          value={latency ? latencyParts(t, lang, latency).value : "—"}
          unit={latency ? latencyParts(t, lang, latency).unit : t("ms_unit")}
          flare={tone.ping}
          active={busy && phase === "ping"}
        />
        <Phase
          icon={ArrowDownToLine}
          label={t("speed_download")}
          value={fmt(download)}
          unit={t("speed_unit")}
          flare={tone.download}
          active={busy && phase === "download"}
        />
        <Phase
          icon={ArrowUpFromLine}
          label={t("speed_upload")}
          value={uploadOff ? t("speed_upload_off") : fmt(upload)}
          unit={uploadOff ? "" : t("speed_unit")}
          flare={tone.upload}
          active={busy && phase === "upload"}
        />
      </div>

      <Card labelledBy="speed-history">
        <CardHead id="speed-history" title={t("speed_history")} />
        {history.error && !history.data ? (
          <Banner
            tone="danger"
            icon={CircleAlert}
            actions={
              <Button size="sm" icon={RefreshCw} onClick={history.reload}>
                {t("retry")}
              </Button>
            }
          >
            {t("load_failed")}
          </Banner>
        ) : history.loading && !history.data ? (
          <div className="stack">
            <Skeleton height={44} />
            <Skeleton height={44} />
          </div>
        ) : !history.data?.length ? (
          <Empty icon={GaugeIcon}>{t("speed_history_empty")}</Empty>
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th scope="col">{t("col_time")}</th>
                  <th scope="col">{t("speed_download")}</th>
                  <th scope="col">{t("speed_upload")}</th>
                  <th scope="col">{t("speed_ping")}</th>
                  <th scope="col" className="hide-sm">
                    {t("speed_source")}
                  </th>
                </tr>
              </thead>
              <tbody>
                {history.data.map((h) => (
                  <tr key={h.timestamp}>
                    <td className="num">{formatTime(lang, h.timestamp, false)}</td>
                    <td className="num">
                      {fmt(h.download_mbps)} <span className="muted">{t("speed_unit")}</span>
                    </td>
                    <td className="num">
                      {h.upload_mbps === null ? "—" : fmt(h.upload_mbps)}{" "}
                      {h.upload_mbps === null ? null : <span className="muted">{t("speed_unit")}</span>}
                    </td>
                    <td className="num">
                      {latencyParts(t, lang, h.latency_ms).value}{" "}
                      <span className="muted">{latencyParts(t, lang, h.latency_ms).unit}</span>
                    </td>
                    <td className="hide-sm secondary">
                      {t(h.triggered_by === "schedule" ? "speed_by_schedule" : "speed_by_user")}
                    </td>
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
