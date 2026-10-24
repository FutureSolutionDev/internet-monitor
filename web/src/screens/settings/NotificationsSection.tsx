import { Volume2 } from "lucide-react";
import { type CSSProperties, useId, useMemo } from "react";
import { TextField } from "../../components/Field";
import { toast } from "../../components/Toast";
import { Button, Card, CardHead, Segmented, Switch } from "../../components/ui";
import { api } from "../../lib/api";
import { formatNumber } from "../../lib/format";
import { type Lang, type MessageKey, useI18n } from "../../lib/i18n";
import { NOTIFY_EVENTS } from "../../lib/settingsForm";
import type { ConfigView, NotifyChannel, NotifyEvent, QuietMode } from "../../lib/types";
import type { SectionProps } from "./Sections";

const EVENT_LABEL: Record<NotifyEvent, MessageKey> = {
  down: "n_ev_down",
  degraded: "n_ev_degraded",
  restored: "n_ev_restored",
  speed: "n_ev_speed",
};

const CHANNEL_LABEL: Record<NotifyChannel, MessageKey> = {
  desktop: "n_ch_desktop",
  webhook: "n_ch_webhook",
  telegram: "n_ch_telegram",
};

const CHANNELS: NotifyChannel[] = ["desktop", "webhook", "telegram"];

const ARABIC_DIGITS = /[٠-٩۰-۹]/g;

function normalizeClock(raw: string): string {
  const v = raw.replace(ARABIC_DIGITS, (d) => String(d.charCodeAt(0) & 0xf)).replace(/[.٫،,\s]/g, ":");
  if (/^\d{4}$/.test(v)) return `${v.slice(0, 2)}:${v.slice(2)}`;
  return v;
}

function weekOrder(lang: Lang): number[] {
  return lang === "ar" ? [6, 0, 1, 2, 3, 4, 5] : [0, 1, 2, 3, 4, 5, 6];
}

function weekdayName(lang: Lang, day: number, width: "short" | "long"): string {
  return new Intl.DateTimeFormat(lang, { weekday: width, timeZone: "UTC" }).format(
    new Date(Date.UTC(2024, 0, 7 + day)),
  );
}

export function NotificationsSection({ draft, set, err, base }: SectionProps & { base: ConfigView }) {
  const { t, lang } = useI18n();
  const ids = useId();
  const volumeId = `${ids}-volume`;
  const daysError = `${ids}-days-error`;
  const daysHint = `${ids}-days-hint`;
  const soundHint = `${ids}-sound-hint`;
  const quietHint = `${ids}-quiet-hint`;
  const matrixTitle = `${ids}-matrix`;
  const offId = (ch: NotifyChannel) => `${ids}-off-${ch}`;

  const configured: Record<NotifyChannel, boolean> = useMemo(() => {
    const webhook =
      draft.webhook.mode === "set"
        ? draft.webhook.value.trim() !== ""
        : draft.webhook.mode === "keep" && base.secret_set.webhook_url;
    const token =
      draft.telegram_token.mode === "set"
        ? draft.telegram_token.value.trim() !== ""
        : draft.telegram_token.mode === "keep" && base.secret_set.telegram_bot_token;
    return { desktop: true, webhook, telegram: token && draft.telegram_chat_id.trim() !== "" };
  }, [draft.webhook, draft.telegram_token, draft.telegram_chat_id, base.secret_set]);

  const toggle = (ev: NotifyEvent, ch: NotifyChannel, on: boolean) => {
    set("notify_events", { ...draft.notify_events, [ev]: { ...draft.notify_events[ev], [ch]: on } });
  };

  const toggleDay = (day: number) => {
    const has = draft.quiet_days.includes(day);
    set(
      "quiet_days",
      has ? draft.quiet_days.filter((d) => d !== day) : [...draft.quiet_days, day].sort((a, b) => a - b),
    );
  };

  const testVolume = (volume = draft.sound_volume) => {
    if (volume === 0) return;
    api.playSound(volume).catch(() => toast(t("sound_play_failed"), "error"));
  };

  const previewOnRelease = (e: { currentTarget: HTMLInputElement }) =>
    testVolume(Number(e.currentTarget.value));

  const daysErr = err("quiet_days");
  const volumeText = `${formatNumber(lang, draft.sound_volume, 0)}${t("pct_unit")}`;

  return (
    <Card id="notifications" className="settings-section" labelledBy="notifications-title">
      <CardHead id="notifications-title" title={t("sec_notifications")} sub={t("n_sub")} />
      <div className="form-grid">
        <div className="switch-row full">
          <div className="field">
            <span className="field-label">{t("n_sound")}</span>
            <span className="field-hint" id={soundHint}>
              {t("n_sound_hint")}
            </span>
          </div>
          <Switch
            checked={draft.sound_enabled}
            onChange={(v) => set("sound_enabled", v)}
            label={t("n_sound")}
            describedBy={soundHint}
          />
        </div>
        <div className="field full">
          <label className="field-label" htmlFor={volumeId}>
            {t("n_volume")}
          </label>
          <div className="volume-row">
            <input
              id={volumeId}
              type="range"
              className="range"
              min={0}
              max={100}
              step={5}
              value={draft.sound_volume}
              disabled={!draft.sound_enabled}
              aria-valuetext={volumeText}
              style={{ "--fill": `${draft.sound_volume}%` } as CSSProperties}
              onChange={(e) => set("sound_volume", Number(e.target.value))}
              onPointerUp={previewOnRelease}
              onKeyUp={(e) => {
                if (e.key.startsWith("Arrow") || e.key === "Home" || e.key === "End") previewOnRelease(e);
              }}
            />
            <span className="volume-value num" aria-hidden>
              {volumeText}
            </span>
            <Button
              variant="ghost"
              size="sm"
              icon={Volume2}
              disabled={!draft.sound_enabled || draft.sound_volume === 0}
              onClick={() => testVolume()}
            >
              {t("n_test_volume")}
            </Button>
          </div>
        </div>

        <hr className="divider full" />

        <div className="switch-row full">
          <div className="field">
            <span className="field-label">{t("n_quiet")}</span>
            <span className="field-hint" id={quietHint}>
              {t("n_quiet_hint")}
            </span>
          </div>
          <Switch
            checked={draft.quiet_enabled}
            onChange={(v) => set("quiet_enabled", v)}
            label={t("n_quiet")}
            describedBy={quietHint}
          />
        </div>
        {draft.quiet_enabled ? (
          <>
            <TextField
              label={t("n_quiet_from")}
              error={err("quiet_start")}
              value={draft.quiet_start}
              onChange={(v) => set("quiet_start", normalizeClock(v))}
              placeholder="22:00"
              inputMode="numeric"
              ltr
              narrow
            />
            <TextField
              label={t("n_quiet_to")}
              error={err("quiet_end")}
              value={draft.quiet_end}
              onChange={(v) => set("quiet_end", normalizeClock(v))}
              placeholder="07:00"
              inputMode="numeric"
              ltr
              narrow
            />
            <fieldset
              className="field full field-set"
              aria-describedby={daysErr ? `${daysError} ${daysHint}` : daysHint}
            >
              <legend className="field-label">{t("n_quiet_days")}</legend>
              <div className="day-chips">
                {weekOrder(lang).map((day) => (
                  <button
                    key={day}
                    type="button"
                    className="chip"
                    aria-pressed={draft.quiet_days.includes(day)}
                    title={weekdayName(lang, day, "long")}
                    onClick={() => toggleDay(day)}
                  >
                    {weekdayName(lang, day, "short")}
                  </button>
                ))}
              </div>
              {daysErr ? (
                <span className="field-error" id={daysError}>
                  {daysErr}
                </span>
              ) : null}
              <span className="field-hint" id={daysHint}>
                {t("n_quiet_days_hint")}
              </span>
            </fieldset>
            <div className="field full">
              <span className="field-label">{t("n_quiet_mode")}</span>
              <Segmented<QuietMode>
                label={t("n_quiet_mode")}
                value={draft.quiet_mode}
                onChange={(v) => set("quiet_mode", v)}
                options={[
                  { value: "sound", label: t("n_quiet_mode_sound") },
                  { value: "all", label: t("n_quiet_mode_all") },
                ]}
              />
              <span className="field-hint">{t("n_quiet_mode_hint")}</span>
            </div>
          </>
        ) : null}

        <hr className="divider full" />

        <div className="field full">
          <span className="field-label" id={matrixTitle}>
            {t("n_matrix")}
          </span>
          <div className="matrix-wrap">
            <table className="matrix" aria-labelledby={matrixTitle}>
              <thead>
                <tr>
                  <th scope="col">
                    <span className="sr-only">{t("n_event")}</span>
                  </th>
                  {CHANNELS.map((ch) => (
                    <th key={ch} scope="col">
                      <span>{t(CHANNEL_LABEL[ch])}</span>
                      {configured[ch] ? null : (
                        <span className="matrix-off" id={offId(ch)}>
                          {t("n_ch_off")}
                        </span>
                      )}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {NOTIFY_EVENTS.map((ev) => (
                  <tr key={ev}>
                    <th scope="row">{t(EVENT_LABEL[ev])}</th>
                    {CHANNELS.map((ch) => (
                      <td key={ch}>
                        <Switch
                          checked={configured[ch] && draft.notify_events[ev][ch]}
                          disabled={!configured[ch]}
                          onChange={(v) => toggle(ev, ch, v)}
                          label={`${t(EVENT_LABEL[ev])} · ${t(CHANNEL_LABEL[ch])}`}
                          describedBy={configured[ch] ? undefined : offId(ch)}
                        />
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <span className="field-hint">{t("n_matrix_hint")}</span>
        </div>

        <hr className="divider full" />

        <TextField
          label={t("n_min_outage")}
          hint={t("n_min_outage_hint")}
          error={err("min_outage")}
          value={draft.min_outage}
          onChange={(v) => set("min_outage", v)}
          inputMode="numeric"
          suffix={t("sec_unit")}
          narrow
        />
        <TextField
          label={t("n_reminder")}
          hint={t("n_reminder_hint")}
          error={err("reminder")}
          value={draft.reminder}
          onChange={(v) => set("reminder", v)}
          inputMode="numeric"
          suffix={t("min_unit")}
          narrow
        />
      </div>
    </Card>
  );
}
