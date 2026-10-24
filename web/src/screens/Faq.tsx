import { MessagesSquare } from "lucide-react";
import { Fragment } from "react";
import { RouteLink } from "../components/RouteLink";
import { FAQ } from "../content/faq";
import { pick } from "../content/text";
import { formatNumber } from "../lib/format";
import { useI18n } from "../lib/i18n";

export function Faq() {
  const { t, lang } = useI18n();
  let index = 0;

  return (
    <div className="doc-page">
      <header className="doc-hero">
        <p className="doc-eyebrow">{t("faq_eyebrow")}</p>
        <h1>{t("faq_title")}</h1>
        <p className="secondary">{t("faq_intro")}</p>
      </header>

      <div className="faq-timeline">
        {FAQ.map((section) => {
          const Icon = section.icon;
          return (
            <Fragment key={section.id}>
              <h2 className="faq-seg" data-accent={section.accent} id={`faq-${section.id}`}>
                <span className="faq-seg-pill">
                  <span className="faq-seg-icon">
                    <Icon size={16} aria-hidden />
                  </span>
                  {pick(section.title, lang)}
                </span>
              </h2>
              {section.items.map((item) => {
                index += 1;
                const side = index % 2 === 1 ? "end" : "start";
                return (
                  <div key={item.id} className="faq-item" data-side={side} data-accent={section.accent}>
                    <span className="faq-node" aria-hidden />
                    <article className="faq-card" aria-labelledby={`faq-q-${item.id}`}>
                      <div className="faq-q">
                        <span className="faq-idx num">{t("faq_q", { n: formatNumber(lang, index) })}</span>
                        <h3 id={`faq-q-${item.id}`}>{pick(item.q, lang)}</h3>
                      </div>
                      <p className="faq-a">{pick(item.a, lang)}</p>
                    </article>
                  </div>
                );
              })}
            </Fragment>
          );
        })}
      </div>

      <div className="doc-cta">
        <RouteLink route="contact" className="btn btn-edge">
          <MessagesSquare size={16} aria-hidden />
          {t("faq_more")}
        </RouteLink>
      </div>
    </div>
  );
}
