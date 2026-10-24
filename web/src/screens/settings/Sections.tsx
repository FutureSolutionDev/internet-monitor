import { Bell, Send, Zap } from "lucide-react";
import { useState } from "react";
import { type ItemResult, ListEditor, TextField } from "../../components/Field";
import { toast } from "../../components/Toast";
import { Button, Card, CardHead, Switch } from "../../components/ui";
import { api } from "../../lib/api";
import { errorKey } from "../../lib/errors";
import { formatNumber } from "../../lib/format";
import { type Lang, type MessageKey, type Translate, useI18n } from "../../lib/i18n";
import { type Draft, type FieldKey, MAX_TARGETS } from "../../lib/settingsForm";
import type { ConfigView, TestTargetResult } from "../../lib/types";
import { SecretField } from "./SecretField";

export interface SectionProps {
  draft: Draft;
  set: <K extends keyof Draft>(key: K, value: Draft[K]) => void;
  err: (key: FieldKey) => string | undefined;
}

export function MonitoringSection({ draft, set, err }: SectionProps) {
  const { t } = useI18n();
  return (
    <Card id="monitoring" className="settings-section" labelledBy="monitoring-title">
      <CardHead id="monitoring-title" title={t("sec_monitoring")} />
      <div className="form-grid">
        <TextField
          label={t("f_interval")}
          hint={t("f_interval_hint")}
          error={err("check_interval_sec")}
          value={draft.check_interval_sec}
          onChange={(v) => set("check_interval_sec", v)}
          inputMode="numeric"
          suffix={t("sec_unit")}
          narrow
        />
        <TextField
          label={t("f_fail")}
          hint={t("f_fail_hint")}
          error={err("fail_threshold")}
          value={draft.fail_threshold}
          onChange={(v) => set("fail_threshold", v)}
          inputMode="numeric"
          narrow
        />
        <TextField
          label={t("f_loss")}
          hint={t("f_loss_hint")}
          error={err("packet_loss_threshold")}
          value={draft.packet_loss_threshold}
          onChange={(v) => set("packet_loss_threshold", v)}
          inputMode="decimal"
          suffix={t("pct_unit")}
          narrow
        />
        <TextField
          label={t("f_latency")}
          hint={t("f_latency_hint")}
          error={err("latency_threshold_ms")}
          value={draft.latency_threshold_ms}
          onChange={(v) => set("latency_threshold_ms", v)}
          inputMode="numeric"
          suffix={t("ms_unit")}
          narrow
        />
        <div className="switch-row full">
          <div className="field">
            <span className="field-label" id="icmp-label">
              {t("f_icmp")}
            </span>
            <span className="field-hint" id="icmp-hint">
              {t("f_icmp_hint")}
            </span>
          </div>
          <Switch
            checked={draft.use_icmp}
            onChange={(v) => set("use_icmp", v)}
            label={t("f_icmp")}
            describedBy="icmp-hint"
          />
        </div>
      </div>
    </Card>
  );
}

const PROBE_KEYS: Record<string, MessageKey> = {
  timeout: "probe_timeout",
  refused: "probe_refused",
  not_found: "probe_not_found",
  unreachable: "probe_unreachable",
  no_permission: "probe_no_permission",
};

function probeText(t: Translate, lang: Lang, r: TestTargetResult): ItemResult {
  if (r.ok) return { ok: true, text: `${formatNumber(lang, r.latency_ms)} ${t("ms_unit")}` };
  const code = r.error ?? "error";
  if (code.startsWith("http_")) return { ok: false, text: `HTTP ${code.slice(5)}` };
  return { ok: false, text: t(PROBE_KEYS[code] ?? "probe_error") };
}

export function TargetsSection({ draft, set, err }: SectionProps) {
  const { t, lang } = useI18n();
  const [testing, setTesting] = useState(false);
  const [results, setResults] = useState<Record<string, ItemResult>>({});
  const [summary, setSummary] = useState<string | null>(null);

  const test = async () => {
    setTesting(true);
    try {
      const r = await api.testTargets({
        ping_targets: draft.ping_targets.map((v) => v.trim()).filter(Boolean),
        http_targets: draft.http_targets.map((v) => v.trim()).filter(Boolean),
        dns_targets: draft.dns_targets.map((v) => v.trim()).filter(Boolean),
      });
      const all = [...(r.ping_targets ?? []), ...(r.http_targets ?? []), ...(r.dns_targets ?? [])];
      const map: Record<string, ItemResult> = {};
      for (const item of all) map[item.target] = probeText(t, lang, item);
      setResults(map);
      setSummary(t("test_ok", { ok: all.filter((x) => x.ok).length, total: all.length }));
    } catch {
      toast(t("probe_error"), "error");
    } finally {
      setTesting(false);
    }
  };

  const common = {
    max: MAX_TARGETS,
    addLabel: t("add"),
    removeLabel: (item: string) => t("remove_item", { item }),
    results,
  };

  return (
    <Card id="targets" className="settings-section" labelledBy="targets-title-s">
      <CardHead
        id="targets-title-s"
        title={t("sec_targets")}
        sub={summary ?? t("max_items", { n: MAX_TARGETS })}
      >
        <Button variant="edge" size="sm" icon={Zap} busy={testing} onClick={test}>
          {testing ? t("testing") : t("test_targets")}
        </Button>
      </CardHead>
      <div className="stack" style={{ gap: 20 }}>
        {err("targets") ? (
          <span className="field-error" role="alert">
            {err("targets")}
          </span>
        ) : null}
        <ListEditor
          {...common}
          label={t("f_ping")}
          hint={t("f_ping_hint")}
          error={err("ping_targets")}
          items={draft.ping_targets}
          onChange={(v) => set("ping_targets", v)}
          placeholder="example.com:443"
        />
        <ListEditor
          {...common}
          label={t("f_http")}
          hint={t("f_http_hint")}
          error={err("http_targets")}
          items={draft.http_targets}
          onChange={(v) => set("http_targets", v)}
          placeholder="https://"
        />
        <ListEditor
          {...common}
          label={t("f_dns")}
          hint={t("f_dns_hint")}
          error={err("dns_targets")}
          items={draft.dns_targets}
          onChange={(v) => set("dns_targets", v)}
          placeholder="www.example.com"
        />
      </div>
    </Card>
  );
}

export function AlertsSection({ draft, set, err, base }: SectionProps & { base: ConfigView }) {
  const { t, lang } = useI18n();
  const [busy, setBusy] = useState<"webhook" | "notify" | null>(null);

  const testWebhook = async () => {
    setBusy("webhook");
    try {
      await api.testWebhook(draft.webhook.mode === "set" ? draft.webhook.value.trim() : undefined);
      toast(t("webhook_ok"), "ok");
    } catch (e) {
      toast(t(errorKey(e, "err_webhook_failed")), "error");
    } finally {
      setBusy(null);
    }
  };

  const testNotify = async () => {
    setBusy("notify");
    try {
      await api.testNotification(lang, draft.sound_enabled ? draft.sound_volume : 0);
      toast(t("notification_sent"), "ok");
    } catch (e) {
      toast(t(errorKey(e, "err_notify_failed")), "error");
    } finally {
      setBusy(null);
    }
  };

  const canTestWebhook =
    (draft.webhook.mode === "set" && !!draft.webhook.value.trim() && !err("webhook")) ||
    (draft.webhook.mode === "keep" && base.secret_set.webhook_url);

  return (
    <Card id="alerts" className="settings-section" labelledBy="alerts-title">
      <CardHead id="alerts-title" title={t("sec_alerts")}>
        <Button variant="edge" size="sm" icon={Bell} busy={busy === "notify"} onClick={testNotify}>
          {t("test_notification")}
        </Button>
      </CardHead>
      <div className="form-grid">
        <SecretField
          t={t}
          label={t("f_webhook")}
          hint={t("f_webhook_hint")}
          error={err("webhook")}
          saved={base.secret_set.webhook_url}
          value={draft.webhook}
          onChange={(v) => set("webhook", v)}
          placeholder="https://discord.com/api/webhooks/…"
        />
        <div className="full row">
          <Button
            variant="ghost"
            size="sm"
            icon={Send}
            busy={busy === "webhook"}
            disabled={!canTestWebhook}
            onClick={testWebhook}
          >
            {t("test_webhook")}
          </Button>
        </div>
        <hr className="divider full" />
        <SecretField
          t={t}
          label={t("f_tg_token")}
          hint={t("f_tg_hint")}
          error={err("telegram_token")}
          saved={base.secret_set.telegram_bot_token}
          value={draft.telegram_token}
          onChange={(v) => set("telegram_token", v)}
          placeholder="123456:ABC…"
          masked
        />
        <TextField
          label={t("f_tg_chat")}
          value={draft.telegram_chat_id}
          onChange={(v) => set("telegram_chat_id", v)}
          ltr
          narrow
          placeholder="-1001234567890"
        />
      </div>
    </Card>
  );
}

export function SpeedSection({ draft, set, err }: SectionProps) {
  const { t } = useI18n();
  return (
    <Card id="speed" className="settings-section" labelledBy="speed-title-s">
      <CardHead id="speed-title-s" title={t("sec_speed")} />
      <div className="form-grid">
        <TextField
          label={t("f_speed_duration")}
          hint={t("f_speed_duration_hint")}
          error={err("speed_timeout")}
          value={draft.speed_timeout}
          onChange={(v) => set("speed_timeout", v)}
          inputMode="numeric"
          suffix={t("sec_unit")}
          narrow
        />
        <TextField
          label={t("f_speed_parallel")}
          hint={t("f_speed_parallel_hint")}
          error={err("speed_parallel")}
          value={draft.speed_parallel}
          onChange={(v) => set("speed_parallel", v)}
          inputMode="numeric"
          narrow
        />
        <TextField
          label={t("f_speed_alert")}
          hint={t("f_speed_alert_hint")}
          error={err("speed_alert")}
          value={draft.speed_alert}
          onChange={(v) => set("speed_alert", v)}
          inputMode="decimal"
          suffix="Mbps"
          narrow
        />
        <TextField
          label={t("f_speed_schedule")}
          hint={t("f_speed_schedule_hint")}
          error={err("speed_schedule")}
          value={draft.speed_schedule}
          onChange={(v) => set("speed_schedule", v)}
          inputMode="numeric"
          suffix={t("min_unit")}
          narrow
        />
        <ListEditor
          className="full"
          label={t("f_speed_download")}
          error={err("speed_download_targets")}
          items={draft.speed_download_targets}
          onChange={(v) => set("speed_download_targets", v)}
          max={MAX_TARGETS}
          placeholder="https://"
          addLabel={t("add")}
          removeLabel={(item) => t("remove_item", { item })}
        />
        <TextField
          className="full"
          label={t("f_speed_upload")}
          hint={t("f_speed_upload_hint")}
          error={err("speed_upload_target")}
          value={draft.speed_upload_target}
          onChange={(v) => set("speed_upload_target", v)}
          ltr
          placeholder="https://speed.cloudflare.com/__up"
        />
      </div>
    </Card>
  );
}
