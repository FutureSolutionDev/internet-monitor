import { ExternalLink, HeartHandshake, MessagesSquare } from "lucide-react";
import { RouteLink } from "../components/RouteLink";
import {
  ABOUT_STORY,
  DEVELOPER,
  DEVELOPER_URL,
  PRINCIPLES,
  PUBLISHER,
  PUBLISHER_URL,
} from "../content/about";
import { donateMethods } from "../content/donate";
import { REPO_URL } from "../content/text";
import { type MessageKey, useI18n } from "../lib/i18n";
import { useLive } from "../lib/live";
import type { Route } from "../lib/router";

const LINKS: { route: Route; key: MessageKey }[] = [
  { route: "changelog", key: "nav_changelog" },
  { route: "faq", key: "nav_faq" },
  { route: "privacy", key: "nav_privacy" },
  { route: "terms", key: "nav_terms" },
];

export function About() {
  const { t } = useI18n();
  const version = useLive((s) => s.snapshot?.version);

  return (
    <div className="doc-page">
      <header className="doc-hero">
        <img src="/favicon.png" alt="" width={64} height={64} className="about-logo" />
        <p className="doc-eyebrow">{t("about_eyebrow")}</p>
        <h1>{t("about_title")}</h1>
      </header>

      <section className="card about-story" aria-label={t("about_eyebrow")}>
        {ABOUT_STORY.map((p) => (
          <p key={p}>{t(p)}</p>
        ))}
      </section>

      <section aria-labelledby="about-principles">
        <h2 id="about-principles" className="about-heading">
          {t("about_principles")}
        </h2>
        <ul className="about-principles">
          {PRINCIPLES.map((p) => {
            const Icon = p.icon;
            return (
              <li key={p.id} className="card">
                <span className="icon-well">
                  <Icon size={18} aria-hidden />
                </span>
                <h3>{t(p.title)}</h3>
                <p>{t(p.body)}</p>
              </li>
            );
          })}
        </ul>
      </section>

      <section className="card about-details" aria-labelledby="about-details">
        <h2 id="about-details">{t("about_details")}</h2>
        <dl>
          <div>
            <dt>{t("about_publisher")}</dt>
            <dd>
              <a href={PUBLISHER_URL} target="_blank" rel="noreferrer noopener">
                {PUBLISHER}
                <span className="sr-only"> {t("opens_new_tab")}</span>
              </a>
            </dd>
          </div>
          <div>
            <dt>{t("about_developer")}</dt>
            <dd>
              <a href={DEVELOPER_URL} target="_blank" rel="noreferrer noopener">
                {DEVELOPER}
                <span className="sr-only"> {t("opens_new_tab")}</span>
              </a>
            </dd>
          </div>
          {version ? (
            <div>
              <dt>{t("about_version")}</dt>
              <dd className="num">
                <RouteLink route="changelog">{version}</RouteLink>
              </dd>
            </div>
          ) : null}
          <div>
            <dt>{t("about_license")}</dt>
            <dd>
              <RouteLink route="terms">Elastic License 2.0</RouteLink>
            </dd>
          </div>
          <div className="about-wide">
            <dt>{t("about_source")}</dt>
            <dd>
              <a href={REPO_URL} target="_blank" rel="noreferrer noopener" className="ltr" dir="ltr">
                {t("app_name")}
                <span className="sr-only"> {t("opens_new_tab")}</span>
              </a>
            </dd>
          </div>
        </dl>
        <nav className="about-links" aria-label={t("about_related")}>
          {LINKS.map((l) => (
            <RouteLink key={l.route} route={l.route}>
              {t(l.key)}
            </RouteLink>
          ))}
        </nav>
      </section>

      <div className="doc-cta">
        <RouteLink route="contact" className="btn btn-edge">
          <MessagesSquare size={17} aria-hidden />
          {t("about_contact_cta")}
        </RouteLink>
        {donateMethods().length > 0 ? (
          <RouteLink route="support" className="btn btn-ghost">
            <HeartHandshake size={17} aria-hidden />
            {t("nav_support")}
          </RouteLink>
        ) : null}
        <a className="btn btn-ghost" href={REPO_URL} target="_blank" rel="noreferrer noopener">
          GitHub
          <ExternalLink size={15} aria-hidden />
          <span className="sr-only"> {t("opens_new_tab")}</span>
        </a>
      </div>
    </div>
  );
}
