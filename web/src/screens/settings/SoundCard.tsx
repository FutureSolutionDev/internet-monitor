import { Music, Play, RotateCcw, Upload, Volume2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { ConfirmDialog } from "../../components/ConfirmDialog";
import { toast } from "../../components/Toast";
import { Button, Card, CardHead } from "../../components/ui";
import { ApiError, api } from "../../lib/api";
import { useI18n } from "../../lib/i18n";
import { useLive } from "../../lib/live";

export function SoundCard() {
  const { t } = useI18n();
  const revision = useLive((s) => s.soundRevision);
  const [custom, setCustom] = useState<boolean | null>(null);
  const [busy, setBusy] = useState<"upload" | "reset" | null>(null);
  const [confirm, setConfirm] = useState(false);
  const file = useRef<HTMLInputElement>(null);
  const audio = useRef<HTMLAudioElement | null>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: refetch when the server reports a sound change
  useEffect(() => {
    let alive = true;
    api.soundIsCustom().then(
      (v) => {
        if (alive) setCustom(v);
      },
      () => {
        if (alive) setCustom(false);
      },
    );
    return () => {
      alive = false;
    };
  }, [revision]);

  useEffect(() => () => audio.current?.pause(), []);

  const preview = () => {
    audio.current?.pause();
    const a = new Audio(`/notification-sound?t=${Date.now()}`);
    audio.current = a;
    a.play().catch(() => toast(t("probe_error"), "error"));
  };

  const upload = async (f: File) => {
    setBusy("upload");
    try {
      await api.uploadSound(f);
      setCustom(true);
      toast(t("sound_uploaded"), "ok");
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : String(e);
      toast(t("sound_failed", { error: msg }), "error");
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
      setCustom(false);
      toast(t("sound_reset_done"), "ok");
    } catch {
      toast(t("probe_error"), "error");
    } finally {
      setBusy(null);
    }
  };

  return (
    <Card id="sound" className="settings-section" labelledBy="sound-title">
      <CardHead id="sound-title" title={t("sound_title")} sub={t("sound_upload_hint")} />
      <div className="stack">
        <div className="list-row">
          <span
            className="icon-well"
            style={{ color: custom ? "var(--accent-violet)" : "var(--text-secondary)" }}
          >
            <Music size={17} aria-hidden />
          </span>
          <span className="list-row-main">
            <span className="list-row-title">{custom ? t("sound_custom") : t("sound_default")}</span>
            <span className="list-row-sub">
              {custom ? t("sound_using_custom") : t("sound_using_default")}
            </span>
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
          <Button variant="edge" icon={Upload} busy={busy === "upload"} onClick={() => file.current?.click()}>
            {t("sound_upload")}
          </Button>
          <Button variant="ghost" icon={Play} onClick={preview}>
            {t("sound_preview")}
          </Button>
          <Button
            variant="ghost"
            icon={Volume2}
            onClick={() => api.playSound().catch(() => toast(t("probe_error"), "error"))}
          >
            {t("sound_play_app")}
          </Button>
          {custom ? (
            <Button variant="ghost" icon={RotateCcw} busy={busy === "reset"} onClick={() => setConfirm(true)}>
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
