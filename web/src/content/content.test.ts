import { describe, expect, test } from "vitest";
import { ABOUT_STORY, PRINCIPLES } from "./about";
import { compareVersions, RELEASES, releaseOfVersion } from "./changelog";
import { CHANNELS } from "./contact";
import { FAQ } from "./faq";
import { PRIVACY, TERMS } from "./legal";
import type { Text } from "./text";

function filled(text: Text) {
  return text.en.trim().length > 0 && text.ar.trim().length > 0;
}

describe("changelog", () => {
  test("every release and item is written in both languages", () => {
    for (const r of RELEASES) {
      expect(filled(r.title), r.id).toBe(true);
      expect(r.items.length, r.id).toBeGreaterThan(0);
      for (const i of r.items) expect(filled(i.text), `${r.id}: ${i.text.en}`).toBe(true);
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
      const keys = r.items.map((i) => i.text.en);
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
  test("questions and answers exist in both languages with unique ids", () => {
    const ids = new Set<string>();
    for (const s of FAQ) {
      expect(filled(s.title), s.id).toBe(true);
      for (const item of s.items) {
        expect(ids.has(item.id), item.id).toBe(false);
        ids.add(item.id);
        expect(filled(item.q) && filled(item.a), item.id).toBe(true);
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
  ] as const)("%s is complete in both languages", (_, doc) => {
    expect(Number.isNaN(Date.parse(doc.updated))).toBe(false);
    expect(filled(doc.intro)).toBe(true);
    const ids = doc.sections.map((s) => s.id);
    expect(new Set(ids).size).toBe(ids.length);
    expect(ids).toContain("contact");
    for (const s of doc.sections) {
      expect(filled(s.title), s.id).toBe(true);
      const parts = [...s.body, ...(s.list ?? [])];
      for (const p of [...parts, ...(s.quote ? [s.quote] : [])]) {
        expect(filled(p), s.id).toBe(true);
      }
      expect(new Set(parts.map((p) => p.en)).size, s.id).toBe(parts.length);
    }
  });
});

describe("about and contact", () => {
  test("about copy exists in both languages", () => {
    for (const p of ABOUT_STORY) expect(filled(p)).toBe(true);
    for (const p of PRINCIPLES) expect(filled(p.title) && filled(p.body), p.id).toBe(true);
  });

  test("every contact channel has a safe link and copy in both languages", () => {
    const ids = CHANNELS.map((c) => c.id);
    expect(new Set(ids).size).toBe(ids.length);
    for (const c of CHANNELS) {
      expect(c.href, c.id).toMatch(/^(https:\/\/|mailto:)/);
      expect(c.handle.trim().length, c.id).toBeGreaterThan(0);
      expect(filled(c.name) && filled(c.best), c.id).toBe(true);
    }
  });
});
