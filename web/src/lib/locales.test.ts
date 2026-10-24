import { describe, expect, test } from "vitest";
import {
  DEFAULT_LANG,
  flattenMessages,
  LOCALE_FILES,
  LOCALES,
  loadLocales,
  MESSAGE_FILES,
  type MessageTree,
  translate,
} from "./i18n";

const base = LOCALES.find((l) => l.code === DEFAULT_LANG);
const baseKeys = Object.keys(base?.messages ?? {}) as (keyof NonNullable<typeof base>["messages"])[];

function placeholders(text: string | undefined): string[] {
  return [...(text ?? "").matchAll(/\{(\w+)\}/g)].map((m) => m[1] ?? "").sort();
}

describe("locale files", () => {
  test("English is the first locale and the reference", () => {
    expect(LOCALES[0]?.code).toBe(DEFAULT_LANG);
    expect(baseKeys.length).toBeGreaterThan(100);
  });

  test.each(LOCALES.map((l) => [l.code, l] as const))("%s has complete metadata", (code, l) => {
    expect(code).toMatch(/^[a-z]{2,3}(-[A-Za-z0-9]+)*$/);
    expect(l.name.trim()).not.toBe("");
    expect(["ltr", "rtl"]).toContain(l.dir);
    expect(() => new Intl.NumberFormat(l.intl)).not.toThrow();
  });

  test.each(LOCALES.map((l) => [l.code, l] as const))("%s has every key and no extras", (_, l) => {
    const keys = Object.keys(l.messages);
    const extra = keys.filter((k) => !(k in (base?.messages ?? {})));
    const missing = baseKeys.filter((k) => !l.messages[k]?.trim());
    expect(extra).toEqual([]);
    expect(missing).toEqual([]);
  });

  test.each(LOCALES.map((l) => [l.code, l] as const))("%s keeps the English placeholders", (_, l) => {
    const wrong = baseKeys.filter(
      (k) =>
        l.messages[k] !== undefined &&
        placeholders(l.messages[k]).join() !== placeholders(base?.messages[k]).join(),
    );
    expect(wrong).toEqual([]);
  });

  test("every language has the same page files with the same keys", () => {
    const byLang = new Map<string, Map<string, string[]>>();
    for (const [path, data] of Object.entries(LOCALE_FILES)) {
      const [, code = "", file = ""] = /\/([^/]+)\/([^/]+)\.json$/.exec(path) ?? [];
      if (file === "meta") continue;
      const files = byLang.get(code) ?? new Map<string, string[]>();
      byLang.set(code, files);
      files.set(file, Object.keys(flattenMessages(data as MessageTree, path)).sort());
    }
    const reference = byLang.get(DEFAULT_LANG);
    expect([...(reference?.keys() ?? [])].sort()).toEqual(Object.keys(MESSAGE_FILES).sort());
    for (const [code, files] of byLang) {
      expect([...files.keys()].sort(), code).toEqual([...(reference?.keys() ?? [])].sort());
      for (const [file, keys] of files) expect(keys, `${code}/${file}.json`).toEqual(reference?.get(file));
    }
  });

  test("a key defined in two files is rejected", () => {
    const meta = { name: "X", dir: "ltr", intl: "en" };
    expect(() =>
      loadLocales({
        "/locales/en/meta.json": meta,
        "/locales/en/a.json": { faq: { q: "1" } },
        "/locales/en/b.json": { "faq.q": "2" },
      }),
    ).toThrow(/faq\.q/);
    expect(() => loadLocales({ "/locales/en/a.json": { k: "v" } })).toThrow(/meta\.json/);
  });

  test.each([
    ["a number", 1],
    ["null", null],
    ["an array", ["a"]],
  ])("a leaf that is %s is rejected", (_, value) => {
    expect(() => flattenMessages({ k: value } as unknown as MessageTree, "x.json")).toThrow(/x\.json: "k"/);
  });

  test("unknown placeholders are left untouched", () => {
    expect(translate(DEFAULT_LANG, "version_current", {})).toContain("{v}");
  });
});
