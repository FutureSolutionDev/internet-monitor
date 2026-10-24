import { Bell, BellOff, CheckCheck } from "lucide-react";
import { type CSSProperties, type MouseEvent, useEffect, useId, useRef, useState } from "react";
import { formatAgo, formatDate, formatTime } from "../lib/format";
import { useNow } from "../lib/hooks";
import { useI18n } from "../lib/i18n";
import { isUnread, KIND_LABEL, KIND_TONE, noticeDetail, useNotifications } from "../lib/notifications";
import { hrefOf, navigate } from "../lib/router";
import { toast } from "./Toast";
import { Button } from "./ui";

const PREVIEW = 8;

export function NotificationBell() {
  const { t, lang } = useI18n();
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [announce, setAnnounce] = useState("");
  const { data, error, markAllRead } = useNotifications(PREVIEW);
  const now = useNow(30000);
  const root = useRef<HTMLDivElement>(null);
  const button = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLDivElement>(null);
  const previous = useRef<number | null>(null);
  const titleId = useId();
  const panelId = useId();

  const unread = data?.unread ?? 0;
  const items = open ? (data?.items ?? []) : [];
  const badge = unread > 99 ? "99+" : String(unread);

  useEffect(() => {
    if (!data) return;
    if (previous.current !== null && data.unread > previous.current) {
      setAnnounce(t("bell_label_unread", { n: data.unread }));
    }
    previous.current = data.unread;
  }, [data, t]);

  useEffect(() => {
    if (!open) return;
    panel.current?.focus({ preventScroll: true });
    const onDown = (e: PointerEvent) => {
      if (!root.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "Escape") return;
      setOpen(false);
      button.current?.focus();
    };
    const onFocus = (e: FocusEvent) => {
      if (!root.current?.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("pointerdown", onDown);
    document.addEventListener("keydown", onKey);
    document.addEventListener("focusin", onFocus);
    return () => {
      document.removeEventListener("pointerdown", onDown);
      document.removeEventListener("keydown", onKey);
      document.removeEventListener("focusin", onFocus);
    };
  }, [open]);

  const readAll = async () => {
    setBusy(true);
    const ok = await markAllRead();
    setBusy(false);
    if (!ok) toast(t("n_mark_failed"), "error");
    panel.current?.focus({ preventScroll: true });
  };

  const viewAll = (e: MouseEvent<HTMLAnchorElement>) => {
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
    e.preventDefault();
    setOpen(false);
    navigate("notifications");
  };

  return (
    <div ref={root} className="bell">
      <Button
        ref={button}
        variant="edge"
        size="sm"
        iconOnly
        icon={Bell}
        aria-label={unread > 0 ? t("bell_label_unread", { n: unread }) : t("bell_label")}
        title={t("bell_label")}
        aria-expanded={open}
        aria-controls={open ? panelId : undefined}
        aria-haspopup="dialog"
        onClick={() => setOpen((o) => !o)}
      />
      {unread > 0 ? (
        <span className="bell-badge num" aria-hidden>
          {badge}
        </span>
      ) : null}
      <span className="sr-only" role="status">
        {announce}
      </span>
      {open ? (
        <div
          ref={panel}
          id={panelId}
          className="bell-panel"
          role="dialog"
          aria-labelledby={titleId}
          tabIndex={-1}
        >
          <div className="bell-head">
            <strong id={titleId}>{t("bell_label")}</strong>
            <Button
              variant="ghost"
              size="sm"
              icon={CheckCheck}
              busy={busy}
              disabled={!data || unread === 0}
              onClick={readAll}
            >
              {t("bell_mark_read")}
            </Button>
          </div>
          {error && !data ? (
            <p className="bell-empty">{t("load_failed")}</p>
          ) : !data ? (
            <p className="bell-empty" aria-busy="true">
              {t("n_loading")}
            </p>
          ) : items.length === 0 ? (
            <p className="bell-empty">
              <BellOff size={20} aria-hidden />
              {t("bell_empty")}
            </p>
          ) : (
            <ul className="bell-list">
              {items.map((r) => {
                const detail = noticeDetail(t, lang, r);
                const fresh = isUnread(r, data.last_read_at);
                return (
                  <li key={r.id} className="bell-item" data-unread={fresh || undefined}>
                    <span
                      className="dot"
                      style={{ "--tone": KIND_TONE[r.kind] } as CSSProperties}
                      aria-hidden
                    />
                    <span className="bell-item-main">
                      <span className="bell-item-title">
                        {t(KIND_LABEL[r.kind])}
                        {fresh ? <span className="sr-only"> · {t("n_unread")}</span> : null}
                      </span>
                      {detail ? <span className="bell-item-sub">{detail}</span> : null}
                    </span>
                    <time
                      className="bell-item-time"
                      dateTime={r.time}
                      title={`${formatDate(lang, r.time)} ${formatTime(lang, r.time, false)}`}
                    >
                      {formatAgo(t, new Date(r.time).getTime(), now)}
                    </time>
                  </li>
                );
              })}
            </ul>
          )}
          <a className="bell-foot" href={hrefOf("notifications")} onClick={viewAll}>
            {t("bell_view_all")}
          </a>
        </div>
      ) : null}
    </div>
  );
}
