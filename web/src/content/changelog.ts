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
    title: "changelog.v2.title",
    items: [
      { kind: "new", text: "changelog.v2.1" },
      { kind: "new", text: "changelog.v2.2" },
      { kind: "new", text: "changelog.v2.3" },
      { kind: "new", text: "changelog.v2.4" },
      { kind: "new", text: "changelog.v2.5" },
      { kind: "new", text: "changelog.v2.6" },
      { kind: "new", text: "changelog.v2.7" },
      { kind: "new", text: "changelog.v2.17" },
      { kind: "new", text: "changelog.v2.18" },
      { kind: "improved", text: "changelog.v2.8" },
      { kind: "improved", text: "changelog.v2.9" },
      { kind: "improved", text: "changelog.v2.10" },
      { kind: "improved", text: "changelog.v2.19" },
      { kind: "improved", text: "changelog.v2.20" },
      { kind: "improved", text: "changelog.v2.21" },
      { kind: "improved", text: "changelog.v2.22" },
      { kind: "improved", text: "changelog.v2.23" },
      { kind: "improved", text: "changelog.v2.24" },
      { kind: "fixed", text: "changelog.v2.11" },
      { kind: "fixed", text: "changelog.v2.12" },
      { kind: "security", text: "changelog.v2.13" },
      { kind: "security", text: "changelog.v2.14" },
      { kind: "security", text: "changelog.v2.15" },
      { kind: "security", text: "changelog.v2.16" },
    ],
  },
  {
    id: "v0.10.1",
    version: "0.10.1",
    date: "2026-06-01",
    title: "changelog.v0_10_1.title",
    items: [{ kind: "improved", text: "changelog.v0_10_1.1" }],
  },
  {
    id: "v0.10.0",
    version: "0.10.0",
    date: "2026-06-01",
    title: "changelog.v0_10_0.title",
    items: [
      { kind: "new", text: "changelog.v0_10_0.1" },
      { kind: "improved", text: "changelog.v0_10_0.2" },
      { kind: "fixed", text: "changelog.v0_10_0.3" },
    ],
  },
  {
    id: "v0.9.4",
    version: "0.9.4",
    from: "0.9.0",
    date: "2026-06-01",
    title: "changelog.v0_9_4.title",
    items: [
      { kind: "new", text: "changelog.v0_9_4.1" },
      { kind: "new", text: "changelog.v0_9_4.2" },
      { kind: "new", text: "changelog.v0_9_4.3" },
      { kind: "new", text: "changelog.v0_9_4.4" },
      { kind: "new", text: "changelog.v0_9_4.5" },
      { kind: "new", text: "changelog.v0_9_4.6" },
      { kind: "improved", text: "changelog.v0_9_4.7" },
      { kind: "improved", text: "changelog.v0_9_4.8" },
      { kind: "fixed", text: "changelog.v0_9_4.9" },
      { kind: "fixed", text: "changelog.v0_9_4.10" },
      { kind: "security", text: "changelog.v0_9_4.11" },
      { kind: "security", text: "changelog.v0_9_4.12" },
    ],
  },
  {
    id: "v0.8.1",
    version: "0.8.1",
    from: "0.8.0",
    date: "2026-05-12",
    title: "changelog.v0_8_1.title",
    items: [
      { kind: "improved", text: "changelog.v0_8_1.1" },
      { kind: "improved", text: "changelog.v0_8_1.2" },
    ],
  },
  {
    id: "v0.7.0",
    version: "0.7.0",
    date: "2026-05-12",
    title: "changelog.v0_7_0.title",
    items: [{ kind: "new", text: "changelog.v0_7_0.1" }],
  },
  {
    id: "v0.6.0",
    version: "0.6.0",
    date: "2026-05-12",
    title: "changelog.v0_6_0.title",
    items: [
      { kind: "new", text: "changelog.v0_6_0.1" },
      { kind: "fixed", text: "changelog.v0_6_0.2" },
    ],
  },
  {
    id: "v0.5.0",
    version: "0.5.0",
    date: "2026-05-12",
    title: "changelog.v0_5_0.title",
    items: [
      { kind: "new", text: "changelog.v0_5_0.1" },
      { kind: "new", text: "changelog.v0_5_0.2" },
      { kind: "improved", text: "changelog.v0_5_0.3" },
      { kind: "improved", text: "changelog.v0_5_0.4" },
      { kind: "fixed", text: "changelog.v0_5_0.5" },
    ],
  },
  {
    id: "v0.4.3",
    version: "0.4.3",
    from: "0.4.0",
    date: "2026-05-12",
    title: "changelog.v0_4_3.title",
    items: [
      { kind: "new", text: "changelog.v0_4_3.1" },
      { kind: "fixed", text: "changelog.v0_4_3.2" },
      { kind: "security", text: "changelog.v0_4_3.3" },
    ],
  },
  {
    id: "v0.3.2",
    version: "0.3.2",
    from: "0.3.0",
    date: "2026-05-11",
    title: "changelog.v0_3_2.title",
    items: [
      { kind: "new", text: "changelog.v0_3_2.1" },
      { kind: "improved", text: "changelog.v0_3_2.2" },
      { kind: "fixed", text: "changelog.v0_3_2.3" },
    ],
  },
  {
    id: "v0.2.0",
    version: "0.2.0",
    date: "2026-05-11",
    title: "changelog.v0_2_0.title",
    items: [
      { kind: "new", text: "changelog.v0_2_0.1" },
      { kind: "new", text: "changelog.v0_2_0.2" },
      { kind: "new", text: "changelog.v0_2_0.3" },
      { kind: "improved", text: "changelog.v0_2_0.4" },
    ],
  },
  {
    id: "v0.1.0",
    version: "0.1.0",
    date: "2026-05-11",
    title: "changelog.v0_1_0.title",
    items: [
      { kind: "new", text: "changelog.v0_1_0.1" },
      { kind: "new", text: "changelog.v0_1_0.2" },
    ],
  },
  {
    id: "v0.0.1",
    version: "0.0.1",
    date: "2026-05-11",
    title: "changelog.v0_0_1.title",
    items: [
      { kind: "new", text: "changelog.v0_0_1.1" },
      { kind: "new", text: "changelog.v0_0_1.2" },
      { kind: "new", text: "changelog.v0_0_1.3" },
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
