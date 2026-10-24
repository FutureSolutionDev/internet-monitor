import type { MessageKey } from "../lib/i18n";
import { type ChangeKind, KINDS, type Release } from "./changelog";

export type KindFilter = ChangeKind | "all";

export interface KindGroup {
  kind: ChangeKind;
  items: MessageKey[];
}

export interface TimelineRelease {
  release: Release;
  groups: KindGroup[];
  total: number;
}

export interface TimelineMonth {
  key: string;
  releases: TimelineRelease[];
}

export interface Timeline {
  months: TimelineMonth[];
  releases: number;
  changes: number;
}

export const NEXT_MONTH = "next";

function normalize(text: string): string {
  return text.toLocaleLowerCase().normalize("NFKD").replace(/[ً-ٰٟ]/g, "");
}

export function buildTimeline(
  releases: Release[],
  filter: KindFilter,
  query: string,
  textOf: (key: MessageKey) => string,
): Timeline {
  const needle = normalize(query.trim());
  const months: TimelineMonth[] = [];
  let releaseCount = 0;
  let changeCount = 0;

  for (const release of releases) {
    const versionHit =
      needle !== "" &&
      [release.version, release.from, release.id, textOf(release.title)].some(
        (v) => v && normalize(v).includes(needle),
      );
    const groups: KindGroup[] = [];
    for (const kind of KINDS) {
      if (filter !== "all" && filter !== kind) continue;
      const items = release.items
        .filter((i) => i.kind === kind)
        .filter((i) => needle === "" || versionHit || normalize(textOf(i.text)).includes(needle))
        .map((i) => i.text);
      if (items.length > 0) groups.push({ kind, items });
    }
    const total = groups.reduce((n, g) => n + g.items.length, 0);
    if (total === 0) continue;

    const key = release.date ? release.date.slice(0, 7) : NEXT_MONTH;
    let month = months[months.length - 1];
    if (!month || month.key !== key) {
      month = { key, releases: [] };
      months.push(month);
    }
    month.releases.push({ release, groups, total });
    releaseCount += 1;
    changeCount += total;
  }

  return { months, releases: releaseCount, changes: changeCount };
}
