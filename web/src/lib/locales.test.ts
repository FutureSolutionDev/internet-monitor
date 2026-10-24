import { describe, expect, test } from "vitest";
import { DEFAULT_LANG, LOCALES, translate } from "./i18n";

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

  test("unknown placeholders are left untouched", () => {
    expect(translate(DEFAULT_LANG, "version_current", {})).toContain("{v}");
  });
});
