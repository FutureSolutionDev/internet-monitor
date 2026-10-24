import { describe, expect, test } from "vitest";
import { type MessageKey, translate } from "../lib/i18n";
import { RELEASES } from "./changelog";
import { buildTimeline, NEXT_MONTH } from "./changelogView";

const en = (key: MessageKey) => translate("en", key);
const ar = (key: MessageKey) => translate("ar", key);

describe("changelog timeline", () => {
  test("groups releases by month, newest first", () => {
    const t = buildTimeline(RELEASES, "all", "", en);
    expect(t.months.map((m) => m.key)).toEqual(["2026-10", "2026-06", "2026-05"]);
    const unreleased = buildTimeline(
      [{ ...(RELEASES[0] as (typeof RELEASES)[number]), id: "next", version: null, date: null }, ...RELEASES],
      "all",
      "",
      en,
    );
    expect(unreleased.months[0]?.key).toBe(NEXT_MONTH);
    expect(t.releases).toBe(RELEASES.length);
    expect(t.changes).toBe(RELEASES.reduce((n, r) => n + r.items.length, 0));
  });

  test("orders kinds as new, improved, fixed, security inside a release", () => {
    const v2 = buildTimeline(RELEASES, "all", "", en).months[0]?.releases[0];
    expect(v2?.groups.map((g) => g.kind)).toEqual(["new", "improved", "fixed", "security"]);
  });

  test("the kind filter drops releases without that kind", () => {
    const t = buildTimeline(RELEASES, "security", "", en);
    for (const m of t.months)
      for (const r of m.releases) expect(r.groups.map((g) => g.kind)).toEqual(["security"]);
    expect(t.releases).toBeLessThan(RELEASES.length);
  });

  test("search matches item text in the current language, ignoring case", () => {
    const t = buildTimeline(RELEASES, "all", "TELEGRAM", en);
    expect(t.changes).toBeGreaterThan(0);
    for (const m of t.months)
      for (const r of m.releases)
        for (const g of r.groups) for (const i of g.items) expect(en(i).toLowerCase()).toContain("telegram");
  });

  test("search ignores Arabic diacritics", () => {
    const plain = buildTimeline(RELEASES, "all", "التحديث", ar).changes;
    expect(plain).toBeGreaterThan(0);
    expect(buildTimeline(RELEASES, "all", "التَّحْدِيث", ar).changes).toBe(plain);
  });

  test("searching a version shows that whole release", () => {
    const t = buildTimeline(RELEASES, "all", "0.7.0", en);
    expect(t.releases).toBe(1);
    expect(t.months[0]?.releases[0]?.release.id).toBe("v0.7.0");
  });

  test("no match returns an empty timeline", () => {
    const t = buildTimeline(RELEASES, "all", "zzzz-no-such-text", en);
    expect(t).toEqual({ months: [], releases: 0, changes: 0 });
  });
});
