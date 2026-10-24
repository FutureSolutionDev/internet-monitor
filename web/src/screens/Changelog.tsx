import {
  ChevronDown,
  type LucideIcon,
  Search,
  ShieldCheck,
  Sparkles,
  TrendingUp,
  Wrench,
  X,
} from "lucide-react";
import { type CSSProperties, useEffect, useMemo, useRef, useState } from "react";
import { type ChangeKind, KINDS, RELEASES, type Release, releaseOfVersion } from "../content/changelog";
import { buildTimeline, type KindFilter, NEXT_MONTH } from "../content/changelogView";
import { formatDate, formatMonth, formatNumber } from "../lib/format";
import { type MessageKey, type Translate, useI18n } from "../lib/i18n";
import { useLive } from "../lib/live";

const KIND: Record<ChangeKind, { label: MessageKey; icon: LucideIcon; tone: string }> = {
  new: { label: "kind_new", icon: Sparkles, tone: "var(--accent-blue)" },
  improved: { label: "kind_improved", icon: TrendingUp, tone: "var(--accent-violet)" },
  fixed: { label: "kind_fixed", icon: Wrench, tone: "var(--accent-green)" },
  security: { label: "kind_security", icon: ShieldCheck, tone: "var(--accent-amber)" },
};

const DATE_OPTS: Intl.DateTimeFormatOptions = { year: "numeric", month: "long", day: "numeric" };

function versionLabel(release: Release, t: Translate): string {
  if (release.version === null) return t("cl_next");
  if (release.from) return t("cl_versions", { from: release.from, to: release.version });
  return `v${release.version}`;
}

function shortVersion(release: Release, t: Translate): string {
  return release.version === null ? t("cl_next") : `v${release.version}`;
}

const latestId = RELEASES.find((r) => r.version !== null)?.id;

function anchorOf(id: string): string {
  return `release-${id.replace(/\./g, "-")}`;
}

function prefersReducedMotion(): boolean {
  try {
    return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  } catch {
    return true;
  }
}

export function Changelog() {
  const { t, lang } = useI18n();
  const running = useLive((s) => s.snapshot?.version);
  const installed = releaseOfVersion(running);
  const [filter, setFilter] = useState<KindFilter>("all");
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState<Set<string>>(() => new Set([RELEASES[0]?.id ?? ""]));
  const [active, setActive] = useState<string | null>(null);
  const list = useRef<HTMLDivElement>(null);

  const timeline = useMemo(() => buildTimeline(RELEASES, filter, query, t), [filter, query, t]);
  const searching = query.trim() !== "" || filter !== "all";

  useEffect(() => {
    if (installed) setOpen((prev) => (prev.has(installed) ? prev : new Set(prev).add(installed)));
  }, [installed]);

  useEffect(() => {
    const hash = decodeURIComponent(window.location.hash.slice(1));
    const target = RELEASES.find((r) => anchorOf(r.id) === hash || r.id === hash);
    if (!target) return;
    setOpen((prev) => new Set(prev).add(target.id));
    requestAnimationFrame(() => document.getElementById(anchorOf(target.id))?.scrollIntoView());
  }, []);

  useEffect(() => {
    const root = list.current;
    if (!root || timeline.releases === 0 || typeof IntersectionObserver === "undefined") return;
    const seen = new Map<string, number>();
    const observer = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          const id = (e.target as HTMLElement).dataset.release ?? "";
          if (e.isIntersecting) seen.set(id, e.boundingClientRect.top);
          else seen.delete(id);
        }
        const first = [...seen.entries()].sort((a, b) => a[1] - b[1])[0];
        if (first) setActive(first[0]);
      },
      { rootMargin: "-80px 0px -55% 0px" },
    );
    for (const el of root.querySelectorAll("[data-release]")) observer.observe(el);
    return () => observer.disconnect();
  }, [timeline]);

  const toggle = (id: string) =>
    setOpen((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const jump = (id: string) => {
    setOpen((prev) => new Set(prev).add(id));
    window.history.replaceState(null, "", `#${anchorOf(id)}`);
    requestAnimationFrame(() =>
      document
        .getElementById(anchorOf(id))
        ?.scrollIntoView({ behavior: prefersReducedMotion() ? "auto" : "smooth", block: "start" }),
    );
  };

  const allIds = timeline.months.flatMap((m) => m.releases.map((r) => r.release.id));
  const allOpen = allIds.length > 0 && allIds.every((id) => open.has(id));

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

  const monthLabel = (key: string) => (key === NEXT_MONTH ? t("cl_in_progress") : formatMonth(lang, key));

  return (
    <div className="doc-page cl-page">
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

      <div className="cl-toolbar">
        <label className="cl-search">
          <Search size={16} aria-hidden />
          <span className="sr-only">{t("cl_search")}</span>
          <input
            type="search"
            value={query}
            placeholder={t("cl_search_placeholder")}
            onChange={(e) => setQuery(e.target.value)}
          />
        </label>
        <fieldset className="cl-filter" aria-label={t("cl_filter")}>
          {(["all", ...KINDS] as KindFilter[]).map((k) => {
            const Icon = k === "all" ? null : KIND[k].icon;
            return (
              <button
                key={k}
                type="button"
                className="chip"
                aria-pressed={filter === k}
                style={k === "all" ? undefined : ({ "--kind": KIND[k].tone } as CSSProperties)}
                onClick={() => setFilter(k)}
              >
                {Icon ? <Icon size={14} aria-hidden /> : null}
                {k === "all" ? t("cl_all") : t(KIND[k].label)}
              </button>
            );
          })}
        </fieldset>
      </div>

      <div className="cl-summary">
        <p role="status" className="cl-count">
          {searching
            ? t("cl_results", {
                changes: formatNumber(lang, timeline.changes),
                releases: formatNumber(lang, timeline.releases),
              })
            : null}
        </p>
        {timeline.releases > 0 && !searching ? (
          <button
            type="button"
            className="btn btn-ghost btn-sm"
            onClick={() => setOpen(allOpen ? new Set() : new Set(allIds))}
          >
            {t(allOpen ? "cl_collapse_all" : "cl_expand_all")}
          </button>
        ) : null}
      </div>

      <div className="cl-layout">
        {timeline.releases > 0 ? (
          <nav className="cl-map" aria-label={t("cl_jump")}>
            <span className="cl-map-title">{t("cl_map")}</span>
            {timeline.months.map((m) => (
              <div key={m.key} className="cl-map-month">
                <span className="cl-map-label">{monthLabel(m.key)}</span>
                {m.releases.map(({ release }) => (
                  <a
                    key={release.id}
                    href={`#${anchorOf(release.id)}`}
                    className="num"
                    aria-current={active === release.id ? "location" : undefined}
                    data-installed={installed === release.id || undefined}
                    onClick={(e) => {
                      e.preventDefault();
                      jump(release.id);
                    }}
                  >
                    {shortVersion(release, t)}
                  </a>
                ))}
              </div>
            ))}
          </nav>
        ) : null}

        <div className="cl-timeline" ref={list}>
          {timeline.releases === 0 ? (
            <div className="card cl-empty">
              <p>{t("cl_no_results")}</p>
              <button
                type="button"
                className="btn btn-edge btn-sm"
                onClick={() => {
                  setQuery("");
                  setFilter("all");
                }}
              >
                <X size={14} aria-hidden />
                {t("cl_clear")}
              </button>
            </div>
          ) : null}
          {timeline.months.map((m) => (
            <section key={m.key} className="cl-month" aria-labelledby={`month-${m.key}`}>
              <h2 id={`month-${m.key}`} className="cl-month-head">
                <span>{monthLabel(m.key)}</span>
              </h2>
              <ol className="cl-releases">
                {m.releases.map(({ release, groups, total }) => {
                  const expanded = searching || open.has(release.id);
                  const mine = installed === release.id;
                  const panel = `${anchorOf(release.id)}-panel`;
                  return (
                    <li
                      key={release.id}
                      id={anchorOf(release.id)}
                      data-release={release.id}
                      className="cl-release"
                      data-current={mine || undefined}
                      data-next={release.version === null || undefined}
                    >
                      <span className="cl-node" aria-hidden />
                      <article className="card cl-card" data-open={expanded || undefined}>
                        <h3 className="cl-card-head">
                          <button
                            type="button"
                            aria-expanded={expanded}
                            aria-controls={panel}
                            onClick={() => toggle(release.id)}
                            disabled={searching}
                          >
                            <span className="cl-card-top">
                              <span className="cl-version num">{versionLabel(release, t)}</span>
                              {release.date ? (
                                <time dateTime={release.date}>
                                  {formatDate(lang, release.date, DATE_OPTS)}
                                </time>
                              ) : null}
                              {release.id === latestId ? (
                                <span className="tag cl-latest">{t("cl_latest")}</span>
                              ) : null}
                              {mine ? <span className="tag cl-current">{t("cl_installed")}</span> : null}
                            </span>
                            <span className="cl-card-title">{t(release.title)}</span>
                            <span className="cl-badges">
                              {groups.map((g) => {
                                const Icon = KIND[g.kind].icon;
                                return (
                                  <span
                                    key={g.kind}
                                    className="cl-badge num"
                                    style={{ "--kind": KIND[g.kind].tone } as CSSProperties}
                                    title={t(KIND[g.kind].label)}
                                  >
                                    <Icon size={12} aria-hidden />
                                    <span className="sr-only">{t(KIND[g.kind].label)}</span>
                                    {formatNumber(lang, g.items.length)}
                                  </span>
                                );
                              })}
                              <span className="sr-only">
                                {t("cl_change_count", { n: formatNumber(lang, total) })}
                              </span>
                            </span>
                            <ChevronDown size={18} className="cl-chevron" aria-hidden />
                          </button>
                        </h3>
                        <div id={panel} className="cl-panel" hidden={!expanded}>
                          {groups.map((g) => {
                            const Icon = KIND[g.kind].icon;
                            return (
                              <section
                                key={g.kind}
                                className="cl-group"
                                style={{ "--kind": KIND[g.kind].tone } as CSSProperties}
                                aria-labelledby={`${panel}-${g.kind}`}
                              >
                                <h4 id={`${panel}-${g.kind}`} className="cl-group-head">
                                  <span className="cl-group-icon">
                                    <Icon size={14} aria-hidden />
                                  </span>
                                  {t(KIND[g.kind].label)}
                                  <span className="cl-group-count num">
                                    {formatNumber(lang, g.items.length)}
                                  </span>
                                </h4>
                                <ul>
                                  {g.items.map((key) => (
                                    <li key={key}>{t(key)}</li>
                                  ))}
                                </ul>
                              </section>
                            );
                          })}
                        </div>
                      </article>
                    </li>
                  );
                })}
              </ol>
            </section>
          ))}
        </div>
      </div>
    </div>
  );
}
