import { Copy, ExternalLink, ShieldAlert } from "lucide-react";
import type { CSSProperties } from "react";
import { toast } from "../components/Toast";
import { Banner, Button } from "../components/ui";
import { CHANNELS, type ContactAccent } from "../content/contact";
import { pick } from "../content/text";
import { useI18n } from "../lib/i18n";

const TONE: Record<ContactAccent, string> = {
  green: "var(--accent-green)",
  violet: "var(--accent-violet)",
  blue: "var(--accent-blue)",
  coral: "var(--accent-coral)",
};

export function Contact() {
  const { t, lang } = useI18n();

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
        <p className="doc-eyebrow">{t("contact_eyebrow")}</p>
        <h1>{t("contact_title")}</h1>
        <p className="secondary">{t("contact_intro")}</p>
      </header>

      <ul className="contact-grid">
        {CHANNELS.map((c) => {
          const Icon = c.icon;
          const name = pick(c.name, lang);
          const external = !c.href.startsWith("mailto:");
          return (
            <li
              key={c.id}
              className="card card-flare contact-card"
              style={{ "--flare": TONE[c.accent], "--tone": TONE[c.accent] } as CSSProperties}
            >
              <span className="contact-icon">
                <Icon size={20} aria-hidden />
              </span>
              <h2>{name}</h2>
              <p className="contact-handle ltr" dir="ltr">
                {c.handle}
              </p>
              <p className="contact-best">{pick(c.best, lang)}</p>
              <div className="contact-actions">
                <a
                  className="btn btn-edge btn-sm"
                  href={c.href}
                  {...(external ? { target: "_blank", rel: "noreferrer noopener" } : {})}
                >
                  {t("contact_open", { name })}
                  {external ? (
                    <>
                      <ExternalLink size={14} aria-hidden />
                      <span className="sr-only"> {t("opens_new_tab")}</span>
                    </>
                  ) : null}
                </a>
                <Button
                  variant="ghost"
                  size="sm"
                  iconOnly
                  icon={Copy}
                  aria-label={t("contact_copy", { name })}
                  title={t("contact_copy", { name })}
                  onClick={() => void copy(c.handle)}
                />
              </div>
            </li>
          );
        })}
      </ul>

      <Banner tone="warn" icon={ShieldAlert}>
        {t("contact_safety")}
      </Banner>
    </div>
  );
}
