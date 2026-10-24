import type { MessageKey } from "../lib/i18n";

export type ChangeKind = "new" | "improved" | "fixed" | "security";

export interface ChangeItem {
  kind: ChangeKind;
  text: MessageKey;
}

export interface Release {
  id: string;
  version: string | null;
  from?: string;
  date: string | null;
  title: MessageKey;
  items: ChangeItem[];
}

export const KINDS: ChangeKind[] = ["new", "improved", "fixed", "security"];

export const RELEASES: Release[] = [
  {
    id: "v2.0.0",
    version: "2.0.0",
    date: "2026-10-24",
    title: "changelog.v2.0.0.title",
    items: [
      { kind: "new", text: "changelog.v2.0.0.items.1" },
      { kind: "new", text: "changelog.v2.0.0.items.2" },
      { kind: "new", text: "changelog.v2.0.0.items.3" },
      { kind: "new", text: "changelog.v2.0.0.items.4" },
      { kind: "new", text: "changelog.v2.0.0.items.5" },
      { kind: "new", text: "changelog.v2.0.0.items.6" },
      { kind: "new", text: "changelog.v2.0.0.items.7" },
      { kind: "new", text: "changelog.v2.0.0.items.17" },
      { kind: "new", text: "changelog.v2.0.0.items.18" },
      { kind: "new", text: "changelog.v2.0.0.items.25" },
      { kind: "new", text: "changelog.v2.0.0.items.28" },
      { kind: "improved", text: "changelog.v2.0.0.items.8" },
      { kind: "improved", text: "changelog.v2.0.0.items.9" },
      { kind: "improved", text: "changelog.v2.0.0.items.10" },
      { kind: "improved", text: "changelog.v2.0.0.items.19" },
      { kind: "improved", text: "changelog.v2.0.0.items.20" },
      { kind: "improved", text: "changelog.v2.0.0.items.21" },
      { kind: "improved", text: "changelog.v2.0.0.items.22" },
      { kind: "improved", text: "changelog.v2.0.0.items.23" },
      { kind: "improved", text: "changelog.v2.0.0.items.24" },
      { kind: "improved", text: "changelog.v2.0.0.items.26" },
      { kind: "improved", text: "changelog.v2.0.0.items.27" },
      { kind: "improved", text: "changelog.v2.0.0.items.29" },
      { kind: "fixed", text: "changelog.v2.0.0.items.11" },
      { kind: "fixed", text: "changelog.v2.0.0.items.12" },
      { kind: "security", text: "changelog.v2.0.0.items.13" },
      { kind: "security", text: "changelog.v2.0.0.items.14" },
      { kind: "security", text: "changelog.v2.0.0.items.15" },
      { kind: "security", text: "changelog.v2.0.0.items.16" },
    ],
  },
  {
    id: "v0.10.1",
    version: "0.10.1",
    date: "2026-06-01",
    title: "changelog.v0.10.1.title",
    items: [{ kind: "improved", text: "changelog.v0.10.1.items.1" }],
  },
  {
    id: "v0.10.0",
    version: "0.10.0",
    date: "2026-06-01",
    title: "changelog.v0.10.0.title",
    items: [
      { kind: "new", text: "changelog.v0.10.0.items.1" },
      { kind: "improved", text: "changelog.v0.10.0.items.2" },
      { kind: "fixed", text: "changelog.v0.10.0.items.3" },
    ],
  },
  {
    id: "v0.9.4",
    version: "0.9.4",
    from: "0.9.0",
    date: "2026-06-01",
    title: "changelog.v0.9.4.title",
    items: [
      { kind: "new", text: "changelog.v0.9.4.items.1" },
      { kind: "new", text: "changelog.v0.9.4.items.2" },
      { kind: "new", text: "changelog.v0.9.4.items.3" },
      { kind: "new", text: "changelog.v0.9.4.items.4" },
      { kind: "new", text: "changelog.v0.9.4.items.5" },
      { kind: "new", text: "changelog.v0.9.4.items.6" },
      { kind: "improved", text: "changelog.v0.9.4.items.7" },
      { kind: "improved", text: "changelog.v0.9.4.items.8" },
      { kind: "fixed", text: "changelog.v0.9.4.items.9" },
      { kind: "fixed", text: "changelog.v0.9.4.items.10" },
      { kind: "security", text: "changelog.v0.9.4.items.11" },
      { kind: "security", text: "changelog.v0.9.4.items.12" },
    ],
  },
  {
    id: "v0.8.1",
    version: "0.8.1",
    from: "0.8.0",
    date: "2026-05-12",
    title: "changelog.v0.8.1.title",
    items: [
      { kind: "improved", text: "changelog.v0.8.1.items.1" },
      { kind: "improved", text: "changelog.v0.8.1.items.2" },
    ],
  },
  {
    id: "v0.7.0",
    version: "0.7.0",
    date: "2026-05-12",
    title: "changelog.v0.7.0.title",
    items: [{ kind: "new", text: "changelog.v0.7.0.items.1" }],
  },
  {
    id: "v0.6.0",
    version: "0.6.0",
    date: "2026-05-12",
    title: "changelog.v0.6.0.title",
    items: [
      { kind: "new", text: "changelog.v0.6.0.items.1" },
      { kind: "fixed", text: "changelog.v0.6.0.items.2" },
    ],
  },
  {
    id: "v0.5.0",
    version: "0.5.0",
    date: "2026-05-12",
    title: "changelog.v0.5.0.title",
    items: [
      { kind: "new", text: "changelog.v0.5.0.items.1" },
      { kind: "new", text: "changelog.v0.5.0.items.2" },
      { kind: "improved", text: "changelog.v0.5.0.items.3" },
      { kind: "improved", text: "changelog.v0.5.0.items.4" },
      { kind: "fixed", text: "changelog.v0.5.0.items.5" },
    ],
  },
  {
    id: "v0.4.3",
    version: "0.4.3",
    from: "0.4.0",
    date: "2026-05-12",
    title: "changelog.v0.4.3.title",
    items: [
      { kind: "new", text: "changelog.v0.4.3.items.1" },
      { kind: "fixed", text: "changelog.v0.4.3.items.2" },
      { kind: "security", text: "changelog.v0.4.3.items.3" },
    ],
  },
  {
    id: "v0.3.2",
    version: "0.3.2",
    from: "0.3.0",
    date: "2026-05-11",
    title: "changelog.v0.3.2.title",
    items: [
      { kind: "new", text: "changelog.v0.3.2.items.1" },
      { kind: "improved", text: "changelog.v0.3.2.items.2" },
      { kind: "fixed", text: "changelog.v0.3.2.items.3" },
    ],
  },
  {
    id: "v0.2.0",
    version: "0.2.0",
    date: "2026-05-11",
    title: "changelog.v0.2.0.title",
    items: [
      { kind: "new", text: "changelog.v0.2.0.items.1" },
      { kind: "new", text: "changelog.v0.2.0.items.2" },
      { kind: "new", text: "changelog.v0.2.0.items.3" },
      { kind: "improved", text: "changelog.v0.2.0.items.4" },
    ],
  },
  {
    id: "v0.1.0",
    version: "0.1.0",
    date: "2026-05-11",
    title: "changelog.v0.1.0.title",
    items: [
      { kind: "new", text: "changelog.v0.1.0.items.1" },
      { kind: "new", text: "changelog.v0.1.0.items.2" },
    ],
  },
  {
    id: "v0.0.1",
    version: "0.0.1",
    date: "2026-05-11",
    title: "changelog.v0.0.1.title",
    items: [
      { kind: "new", text: "changelog.v0.0.1.items.1" },
      { kind: "new", text: "changelog.v0.0.1.items.2" },
      { kind: "new", text: "changelog.v0.0.1.items.3" },
    ],
  },
];

function parts(v: string): number[] {
  const core = v.replace(/^v/, "").split("-")[0] ?? "";
  return core.split(".").map((n) => Number.parseInt(n, 10) || 0);
}

export function compareVersions(a: string, b: string): number {
  const pa = parts(a);
  const pb = parts(b);
  for (let i = 0; i < 3; i++) {
    const d = (pa[i] ?? 0) - (pb[i] ?? 0);
    if (d !== 0) return d;
  }
  return 0;
}

export function isReleasedVersion(v: string | undefined | null): v is string {
  return !!v && /^v?\d+\.\d+\.\d+/.test(v);
}

export function releaseOfVersion(running: string | undefined | null): string | null {
  if (!isReleasedVersion(running)) return null;
  for (const r of RELEASES) {
    if (r.version === null) continue;
    const low = r.from ?? r.version;
    if (compareVersions(running, low) >= 0 && compareVersions(running, r.version) <= 0) return r.id;
  }
  const newest = RELEASES.find((r) => r.version !== null)?.version;
  const next = RELEASES[0];
  if (newest && next?.version === null && compareVersions(running, newest) > 0) return next.id;
  return null;
}
