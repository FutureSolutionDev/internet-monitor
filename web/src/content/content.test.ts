import { describe, expect, test } from "vitest";
import { LOCALES, type MessageKey, translate } from "../lib/i18n";
import { ABOUT_STORY, PRINCIPLES } from "./about";
import { compareVersions, RELEASES, releaseOfVersion } from "./changelog";
import { CHANNELS } from "./contact";
import { FAQ } from "./faq";
import { PRIVACY, TERMS } from "./legal";

function expectTranslated(key: MessageKey) {
  for (const l of LOCALES) {
    const text = l.messages[key];
    expect(text?.trim().length ?? 0, `${l.code}: ${key}`).toBeGreaterThan(0);
  }
}

describe("changelog", () => {
  test("every release and item is translated", () => {
    for (const r of RELEASES) {
      expectTranslated(r.title);
      expect(r.items.length, r.id).toBeGreaterThan(0);
      for (const i of r.items) expectTranslated(i.text);
    }
  });

  test("releases are newest first with valid dates and unique ids", () => {
    const ids = new Set(RELEASES.map((r) => r.id));
    expect(ids.size).toBe(RELEASES.length);
    for (let i = 0; i < RELEASES.length; i++) {
      const r = RELEASES[i];
      if (!r) continue;
      if (r.date === null) expect(i, `${r.id} has no date`).toBe(0);
      else expect(Number.isNaN(Date.parse(r.date))).toBe(false);
      const next = RELEASES[i + 1];
      if (!next?.date || !r.date) continue;
      expect(r.date >= next.date, `${r.id} before ${next.id}`).toBe(true);
      if (r.version && next.version) expect(compareVersions(r.version, next.version)).toBeGreaterThan(0);
    }
  });

  test("item keys are unique inside a release", () => {
    for (const r of RELEASES) {
      const keys = r.items.map((i) => i.text);
      expect(new Set(keys).size, r.id).toBe(keys.length);
    }
  });

  test("finds the release that holds the running version", () => {
    expect(releaseOfVersion("v0.9.2")).toBe("v0.9.4");
    expect(releaseOfVersion("0.10.0")).toBe("v0.10.0");
    expect(releaseOfVersion("v0.11.0")).toBe("v2");
    expect(releaseOfVersion("dev")).toBeNull();
    expect(releaseOfVersion(undefined)).toBeNull();
  });
});

describe("faq", () => {
  test("questions and answers are translated with unique ids", () => {
    const ids = new Set<string>();
    for (const s of FAQ) {
      expectTranslated(s.title);
      for (const item of s.items) {
        expect(ids.has(item.id), item.id).toBe(false);
        ids.add(item.id);
        expectTranslated(item.q);
        expectTranslated(item.a);
      }
    }
    expect(ids.size).toBeGreaterThan(10);
  });

  test("uses at most five accents", () => {
    expect(new Set(FAQ.map((s) => s.accent)).size).toBeLessThanOrEqual(5);
  });
});

describe("legal pages", () => {
  test.each([
    ["privacy", PRIVACY],
    ["terms", TERMS],
  ] as const)("%s is complete", (_, doc) => {
    expect(Number.isNaN(Date.parse(doc.updated))).toBe(false);
    expectTranslated(doc.intro);
    const ids = doc.sections.map((s) => s.id);
    expect(new Set(ids).size).toBe(ids.length);
    expect(ids).toContain("contact");
    for (const s of doc.sections) {
      expectTranslated(s.title);
      const parts = [...s.body, ...(s.list ?? []), ...(s.quote ? [s.quote] : [])];
      for (const p of parts) expectTranslated(p);
      expect(new Set(parts).size, s.id).toBe(parts.length);
    }
  });
});

describe("about and contact", () => {
  test("about copy is translated", () => {
    for (const p of ABOUT_STORY) expectTranslated(p);
    for (const p of PRINCIPLES) {
      expectTranslated(p.title);
      expectTranslated(p.body);
    }
  });

  test("every contact channel has a safe link and translated copy", () => {
    const ids = CHANNELS.map((c) => c.id);
    expect(new Set(ids).size).toBe(ids.length);
    for (const c of CHANNELS) {
      expect(c.href, c.id).toMatch(/^(https:\/\/|mailto:)/);
      expect(c.handle.trim().length, c.id).toBeGreaterThan(0);
      expectTranslated(c.name);
      expectTranslated(c.best);
    }
  });

  test("translate falls back to English for a missing key", () => {
    expect(translate("xx", "app_name")).toBe("Internet Monitor");
  });
});
