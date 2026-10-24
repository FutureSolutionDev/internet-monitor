import { CircleAlert, Music, Play, RotateCcw, Upload, Volume2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { ConfirmDialog } from "../../components/ConfirmDialog";
import { toast } from "../../components/Toast";
import { Button, Card, CardHead } from "../../components/ui";
import { ApiError, api } from "../../lib/api";
import { formatNumber } from "../../lib/format";
import { useI18n } from "../../lib/i18n";
import { useLive } from "../../lib/live";

type Info = { custom: boolean; duration: number } | "error" | null;

export function SoundCard() {
  const { t, lang } = useI18n();
  const revision = useLive((s) => s.soundRevision);
  const [info, setInfo] = useState<Info>(null);
  const [busy, setBusy] = useState<"upload" | "reset" | null>(null);
  const [confirm, setConfirm] = useState(false);
  const file = useRef<HTMLInputElement>(null);
  const uploadButton = useRef<HTMLButtonElement>(null);
  const audio = useRef<HTMLAudioElement | null>(null);

  const refresh = async () => {
    try {
      const next = await api.soundInfo();
      setInfo(next);
      return next;
    } catch {
      setInfo("error");
      return null;
    }
  };

  // biome-ignore lint/correctness/useExhaustiveDependencies: refetch when the server reports a sound change
  useEffect(() => {
    void refresh();
  }, [revision]);

  useEffect(() => () => audio.current?.pause(), []);

  const preview = (quiet = false) => {
    audio.current?.pause();
    const a = new Audio(`/notification-sound?t=${Date.now()}`);
    audio.current = a;
    a.play().catch(() => {
      if (!quiet) toast(t("sound_preview_failed"), "error");
    });
  };

  const playInApp = () => {
    api.playSound().catch(() => toast(t("sound_play_failed"), "error"));
  };

  const upload = async (f: File) => {
    setBusy("upload");
    try {
      await api.uploadSound(f);
      await refresh();
      toast(t("sound_uploaded"), "ok");
      preview(true);
    } catch (e) {
      const key = e instanceof ApiError && e.status === 422 ? "sound_err_format" : "sound_err_size";
      toast(
        e instanceof ApiError && (e.status === 422 || e.status === 400) ? t(key) : t("save_failed"),
        "error",
      );
    } finally {
      setBusy(null);
      if (file.current) file.current.value = "";
    }
  };

  const reset = async () => {
    setConfirm(false);
    setBusy("reset");
    try {
      await api.resetSound();
      await refresh();
      toast(t("sound_reset_done"), "ok");
    } catch {
      toast(t("save_failed"), "error");
    } finally {
      setBusy(null);
      uploadButton.current?.focus();
    }
  };

  const custom = info !== null && info !== "error" && info.custom;
  const title =
    info === "error"
      ? t("sound_check_failed")
      : info === null
        ? "…"
        : custom
          ? t("sound_custom")
          : t("sound_default");
  const sub =
    info === "error"
      ? ""
      : custom && info !== null && info.duration > 0
        ? `${t("sound_using_custom")} · ${t("sound_duration", { n: formatNumber(lang, info.duration, 1) })}`
        : custom
          ? t("sound_using_custom")
          : info === null
            ? ""
            : t("sound_using_default");

  return (
    <Card id="sound" className="settings-section" labelledBy="sound-title">
      <CardHead id="sound-title" title={t("sound_title")} sub={t("sound_upload_hint")} />
      <div className="stack">
        <div className="list-row" aria-live="polite">
          <span
            className="icon-well"
            style={{
              color:
                info === "error"
                  ? "var(--accent-coral)"
                  : custom
                    ? "var(--accent-violet)"
                    : "var(--text-secondary)",
            }}
          >
            {info === "error" ? <CircleAlert size={17} aria-hidden /> : <Music size={17} aria-hidden />}
          </span>
          <span className="list-row-main">
            <span className="list-row-title">{title}</span>
            {sub ? <span className="list-row-sub">{sub}</span> : null}
          </span>
        </div>
        <div className="row">
          <input
            ref={file}
            type="file"
            accept=".mp3,.wav,audio/mpeg,audio/wav,audio/x-wav"
            className="sr-only"
            tabIndex={-1}
            aria-hidden
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) void upload(f);
            }}
          />
          <Button
            ref={uploadButton}
            variant="edge"
            icon={Upload}
            busy={busy === "upload"}
            disabled={busy !== null}
            onClick={() => file.current?.click()}
          >
            {t("sound_upload")}
          </Button>
          <Button variant="ghost" icon={Play} onClick={() => preview()} disabled={busy !== null}>
            {t("sound_preview")}
          </Button>
          <Button variant="ghost" icon={Volume2} onClick={playInApp} disabled={busy !== null}>
            {t("sound_play_app")}
          </Button>
          {custom ? (
            <Button
              variant="ghost"
              icon={RotateCcw}
              busy={busy === "reset"}
              disabled={busy !== null}
              onClick={() => setConfirm(true)}
            >
              {t("sound_reset")}
            </Button>
          ) : null}
        </div>
      </div>
      <ConfirmDialog
        open={confirm}
        title={t("sound_reset_title")}
        body={t("sound_reset_body")}
        confirmLabel={t("sound_reset")}
        cancelLabel={t("cancel")}
        danger
        onConfirm={reset}
        onCancel={() => setConfirm(false)}
      />
    </Card>
  );
}
