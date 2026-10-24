import type { CSSProperties } from "react";
import { useEffect, useId, useMemo, useRef, useState } from "react";
import { CHANGE_KIND } from "../content/changeKinds";
import { isReleasedVersion, KINDS, RELEASES, releaseOfVersion } from "../content/changelog";
import { formatDate } from "../lib/format";
import { useI18n } from "../lib/i18n";
import { useLive } from "../lib/live";
import { navigate } from "../lib/router";
import { Button } from "./ui";

const SEEN_KEY = "im.seenVersion";

function readSeen(): string | null {
  try {
    return localStorage.getItem(SEEN_KEY);
  } catch {
    return null;
  }
}

function writeSeen(v: string) {
  try {
    localStorage.setItem(SEEN_KEY, v);
  } catch {}
}

export function WhatsNew() {
  const { t, lang } = useI18n();
  const running = useLive((s) => s.snapshot?.version);
  const [dismissed, setDismissed] = useState<string | null>(null);
  const ref = useRef<HTMLDialogElement>(null);
  const titleId = useId();

  const release = useMemo(() => {
    if (!isReleasedVersion(running)) return null;
    const id = releaseOfVersion(running);
    return RELEASES.find((r) => r.id === id) ?? null;
  }, [running]);

  const open = !!release && !!running && dismissed !== running && readSeen() !== running;

  useEffect(() => {
    const d = ref.current;
    if (!d) return;
    if (open && !d.open) {
      if (typeof d.showModal === "function") d.showModal();
      else d.setAttribute("open", "");
    }
    if (!open && d.open) {
      if (typeof d.close === "function") d.close();
      else d.removeAttribute("open");
    }
  }, [open]);

  if (!release || !running) return null;

  const close = (then?: () => void) => {
    writeSeen(running);
    setDismissed(running);
    then?.();
  };

  const groups = KINDS.map((kind) => ({ kind, items: release.items.filter((i) => i.kind === kind) })).filter(
    (g) => g.items.length > 0,
  );

  return (
    <dialog
      ref={ref}
      className="dialog whats-new"
      aria-labelledby={titleId}
      onCancel={(e) => {
        e.preventDefault();
        close();
      }}
    >
      <div className="dialog-body">
        <p className="whats-new-eyebrow">{t("wn_eyebrow")}</p>
        <h2 id={titleId}>{t("wn_title", { v: release.version ?? running })}</h2>
        {release.date ? (
          <p className="secondary">
            {formatDate(lang, release.date, { year: "numeric", month: "long", day: "numeric" })}
          </p>
        ) : null}
        <div className="whats-new-list">
          {groups.map(({ kind, items }) => {
            const meta = CHANGE_KIND[kind];
            const Icon = meta.icon;
            return (
              <section
                key={kind}
                className="whats-new-group"
                style={{ "--tone": meta.tone } as CSSProperties}
              >
                <h3>
                  <Icon size={15} aria-hidden />
                  {t(meta.label)}
                </h3>
                <ul>
                  {items.map((i) => (
                    <li key={i.text}>{t(i.text)}</li>
                  ))}
                </ul>
              </section>
            );
          })}
        </div>
      </div>
      <div className="dialog-actions">
        <Button variant="ghost" onClick={() => close(() => navigate("changelog"))}>
          {t("wn_all")}
        </Button>
        <Button variant="solid" onClick={() => close()} autoFocus>
          {t("wn_close")}
        </Button>
      </div>
    </dialog>
  );
}
