import { Copy, ExternalLink, HeartHandshake, ShieldAlert } from "lucide-react";
import type { CSSProperties } from "react";
import { toast } from "../components/Toast";
import { Banner, Button } from "../components/ui";
import { type DonateAccent, donateMethods } from "../content/donate";
import { useI18n } from "../lib/i18n";

const TONE: Record<DonateAccent, string> = {
  blue: "var(--accent-blue)",
  coral: "var(--accent-coral)",
  violet: "var(--accent-violet)",
};

export function Support() {
  const { t } = useI18n();
  const methods = donateMethods();

  const copy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast(t("contact_copied"), "ok");
    } catch {
      toast(t("contact_copy_failed"), "error");
    }
  };

  return (
    <div className="doc-page">
      <header className="doc-hero">
        <p className="doc-eyebrow">{t("support_eyebrow")}</p>
        <h1>{t("support_title")}</h1>
        <p className="secondary">{t("support_intro")}</p>
      </header>

      {methods.length === 0 ? (
        <Banner tone="info" icon={HeartHandshake}>
          {t("support_none")}
        </Banner>
      ) : (
        <ul className="contact-grid">
          {methods.map((m) => {
            const Icon = m.icon;
            const name = t(m.name);
            return (
              <li
                key={m.id}
                className="card card-flare contact-card"
                style={{ "--flare": TONE[m.accent], "--tone": TONE[m.accent] } as CSSProperties}
              >
                <span className="contact-icon">
                  <Icon size={20} aria-hidden />
                </span>
                <h2>{name}</h2>
                <p className="contact-handle ltr" dir="ltr">
                  {m.handle}
                </p>
                <p className="contact-best">{t(m.how)}</p>
                <div className="contact-actions">
                  {m.href ? (
                    <a
                      className="btn btn-edge btn-sm"
                      href={m.href}
                      target="_blank"
                      rel="noreferrer noopener"
                    >
                      {t("contact_open", { name })}
                      <ExternalLink size={14} aria-hidden />
                      <span className="sr-only"> {t("opens_new_tab")}</span>
                    </a>
                  ) : null}
                  <Button
                    variant={m.href ? "ghost" : "edge"}
                    size="sm"
                    iconOnly={!!m.href}
                    icon={Copy}
                    aria-label={t("contact_copy", { name })}
                    title={t("contact_copy", { name })}
                    onClick={() => void copy(m.handle)}
                  >
                    {m.href ? null : t("contact_copy", { name })}
                  </Button>
                </div>
              </li>
            );
          })}
        </ul>
      )}

      <p className="secondary">{t("support_thanks")}</p>

      <Banner tone="warn" icon={ShieldAlert}>
        {t("support_safety")}
      </Banner>
    </div>
  );
}
