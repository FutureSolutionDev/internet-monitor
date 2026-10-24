import { useMemo, useState } from "react";
import { type ChangeKind, KINDS, RELEASES, type Release, releaseOfVersion } from "../content/changelog";
import { formatDate, formatNumber } from "../lib/format";
import { type MessageKey, useI18n } from "../lib/i18n";
import { useLive } from "../lib/live";

type Filter = ChangeKind | "all";

const KIND_LABEL: Record<ChangeKind, MessageKey> = {
  new: "kind_new",
  improved: "kind_improved",
  fixed: "kind_fixed",
  security: "kind_security",
};

const DATE_OPTS: Intl.DateTimeFormatOptions = { year: "numeric", month: "long", day: "numeric" };

function versionLabel(release: Release, t: ReturnType<typeof useI18n>["t"]): string {
  if (release.version === null) return t("cl_next");
  if (release.from) return t("cl_versions", { from: release.from, to: release.version });
  return `v${release.version}`;
}

export function Changelog() {
  const { t, lang } = useI18n();
  const running = useLive((s) => s.snapshot?.version);
  const installed = releaseOfVersion(running);
  const [filter, setFilter] = useState<Filter>("all");

  const counts = useMemo(() => {
    const by: Record<ChangeKind, number> = { new: 0, improved: 0, fixed: 0, security: 0 };
    for (const r of RELEASES) for (const i of r.items) by[i.kind] += 1;
    return by;
  }, []);

  const stats: { key: MessageKey; value: number }[] = [
    { key: "cl_stat_releases", value: RELEASES.length },
    { key: "cl_stat_new", value: counts.new },
    { key: "cl_stat_improved", value: counts.improved },
    { key: "cl_stat_fixed", value: counts.fixed + counts.security },
  ];

  return (
    <div className="doc-page">
      <header className="doc-hero">
        <p className="doc-eyebrow">{t("cl_eyebrow")}</p>
        <h1>{t("cl_title")}</h1>
        <p className="secondary">{t("cl_intro")}</p>
        <dl className="cl-stats">
          {stats.map((s) => (
            <div key={s.key} className="cl-stat">
              <dt>{t(s.key)}</dt>
              <dd className="num">{formatNumber(lang, s.value)}</dd>
            </div>
          ))}
        </dl>
      </header>

      <fieldset className="cl-filter" aria-label={t("cl_filter")}>
        {(["all", ...KINDS] as Filter[]).map((k) => (
          <button
            key={k}
            type="button"
            className="chip"
            data-kind={k}
            aria-pressed={filter === k}
            onClick={() => setFilter(k)}
          >
            {k === "all" ? t("cl_all") : t(KIND_LABEL[k])}
          </button>
        ))}
      </fieldset>

      <ol className="cl-timeline">
        {RELEASES.map((r) => {
          const items = filter === "all" ? r.items : r.items.filter((i) => i.kind === filter);
          if (items.length === 0) return null;
          const mine = installed === r.id;
          return (
            <li key={r.id} id={r.id} className="cl-release" data-current={mine || undefined}>
              <span className="cl-node" aria-hidden />
              <div className="cl-meta">
                <span className="cl-version num">{versionLabel(r, t)}</span>
                {r.date ? <time dateTime={r.date}>{formatDate(lang, r.date, DATE_OPTS)}</time> : null}
                {mine ? <span className="tag cl-current">{t("cl_installed")}</span> : null}
              </div>
              <article className="card cl-card" aria-labelledby={`${r.id}-title`}>
                <h2 id={`${r.id}-title`}>{t(r.title)}</h2>
                <ul className="cl-items">
                  {items.map((i) => (
                    <li key={i.text}>
                      <span className="cl-kind" data-kind={i.kind}>
                        {t(KIND_LABEL[i.kind])}
                      </span>
                      <span>{t(i.text)}</span>
                    </li>
                  ))}
                </ul>
              </article>
            </li>
          );
        })}
      </ol>
    </div>
  );
}
