import { MessagesSquare } from "lucide-react";
import { useEffect } from "react";
import { RouteLink } from "../components/RouteLink";
import { PRIVACY, TERMS } from "../content/legal";
import { formatDate } from "../lib/format";
import { DEFAULT_LANG, translate, useI18n } from "../lib/i18n";

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
        <p className="secondary">{t(doc.intro)}</p>
        <p className="legal-updated">
          {t("legal_updated", { date: formatDate(lang, doc.updated, DATE_OPTS) })}
        </p>
      </header>

      <div className="legal-layout">
        <nav className="legal-toc" aria-label={t("legal_toc")}>
          <span className="legal-toc-title">{t("legal_toc")}</span>
          {doc.sections.map((s) => (
            <a key={s.id} href={`#${prefix}${s.id}`}>
              {t(s.title)}
            </a>
          ))}
        </nav>
        <article className="card legal-body">
          {doc.sections.map((s) => (
            <section key={s.id} id={`${prefix}${s.id}`} aria-labelledby={`${prefix}${s.id}-h`}>
              <h2 id={`${prefix}${s.id}-h`} style={{
                fontSize: "1.2rem",
                fontWeight: "bolder"
              }}>{t(s.title)}</h2>
              {s.body.map((p) => (
                <p key={p}>{t(p)}</p>
              ))}
              {s.list ? (
                <ul>
                  {s.list.map((li) => (
                    <li key={li}>{t(li)}</li>
                  ))}
                </ul>
              ) : null}
              {s.quote ? <blockquote>{t(s.quote)}</blockquote> : null}
              {s.quote && lang !== DEFAULT_LANG ? (
                <blockquote lang={DEFAULT_LANG} dir="ltr">
                  {translate(DEFAULT_LANG, s.quote)}
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
