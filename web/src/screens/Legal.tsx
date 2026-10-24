import { MessagesSquare } from "lucide-react";
import { useEffect } from "react";
import { RouteLink } from "../components/RouteLink";
import { PRIVACY, TERMS } from "../content/legal";
import { pick } from "../content/text";
import { formatDate } from "../lib/format";
import { useI18n } from "../lib/i18n";

const DATE_OPTS: Intl.DateTimeFormatOptions = { year: "numeric", month: "long", day: "numeric" };

export function Legal({ kind }: { kind: "privacy" | "terms" }) {
  const { t, lang } = useI18n();
  const doc = kind === "privacy" ? PRIVACY : TERMS;
  const prefix = `${kind}-`;

  useEffect(() => {
    const id = decodeURIComponent(window.location.hash.slice(1));
    if (id.startsWith(`${kind}-`)) document.getElementById(id)?.scrollIntoView();
  }, [kind]);

  return (
    <div className="doc-page legal">
      <header className="doc-hero">
        <h1>{t(kind === "privacy" ? "privacy_title" : "terms_title")}</h1>
        <p className="secondary">{pick(doc.intro, lang)}</p>
        <p className="legal-updated">
          {t("legal_updated", { date: formatDate(lang, doc.updated, DATE_OPTS) })}
        </p>
      </header>

      <div className="legal-layout">
        <nav className="legal-toc" aria-label={t("legal_toc")}>
          <span className="legal-toc-title">{t("legal_toc")}</span>
          {doc.sections.map((s) => (
            <a key={s.id} href={`#${prefix}${s.id}`}>
              {pick(s.title, lang)}
            </a>
          ))}
        </nav>
        <article className="card legal-body">
          {doc.sections.map((s) => (
            <section key={s.id} id={`${prefix}${s.id}`} aria-labelledby={`${prefix}${s.id}-h`}>
              <h2 id={`${prefix}${s.id}-h`}>{pick(s.title, lang)}</h2>
              {s.body.map((p) => (
                <p key={p.en}>{pick(p, lang)}</p>
              ))}
              {s.list ? (
                <ul>
                  {s.list.map((li) => (
                    <li key={li.en}>{pick(li, lang)}</li>
                  ))}
                </ul>
              ) : null}
              {s.quote ? <blockquote>{pick(s.quote, lang)}</blockquote> : null}
              {s.quote && lang === "ar" ? (
                <blockquote lang="en" dir="ltr">
                  {s.quote.en}
                </blockquote>
              ) : null}
              {s.id === "contact" ? (
                <RouteLink route="contact" className="btn btn-edge btn-sm">
                  <MessagesSquare size={14} aria-hidden />
                  {t("nav_contact")}
                </RouteLink>
              ) : null}
            </section>
          ))}
        </article>
      </div>
    </div>
  );
}
