import { useCallback, useEffect, useMemo, useSyncExternalStore } from "react";
import en from "../../locales/en.json";
import { useLive } from "./live";

export type MessageKey = keyof typeof en.messages;
export type Lang = string;
export type Vars = Record<string, string | number>;
export type Translate = (key: MessageKey, vars?: Vars) => string;

export interface LocaleMeta {
  name: string;
  dir: "ltr" | "rtl";
  intl: string;
}

export interface Locale extends LocaleMeta {
  code: Lang;
  messages: Partial<Record<MessageKey, string>>;
}

interface LocaleFile {
  meta: LocaleMeta;
  messages: Record<string, string>;
}

export const DEFAULT_LANG = "en";

function loadLocales(files: Record<string, LocaleFile>): Locale[] {
  const list = Object.entries(files).map(([path, file]) => ({
    code: path.replace(/^.*\/([^/]+)\.json$/, "$1"),
    ...file.meta,
    messages: file.messages as Locale["messages"],
  }));
  return list.sort((a, b) =>
    a.code === DEFAULT_LANG ? -1 : b.code === DEFAULT_LANG ? 1 : a.code.localeCompare(b.code),
  );
}

export const LOCALES: Locale[] = loadLocales(
  import.meta.glob<LocaleFile>("../../locales/*.json", { eager: true, import: "default" }),
);

const byCode = new Map(LOCALES.map((l) => [l.code, l]));
const fallback = en.messages as Record<MessageKey, string>;

export function hasLang(code: string | null | undefined): code is Lang {
  return !!code && byCode.has(code);
}

export function localeOf(lang: Lang): string {
  return byCode.get(lang)?.intl ?? lang;
}

export function dirOf(lang: Lang): "ltr" | "rtl" {
  return byCode.get(lang)?.dir ?? "ltr";
}

export function isRtl(lang: Lang): boolean {
  return dirOf(lang) === "rtl";
}

export function translate(lang: Lang, key: MessageKey, vars?: Vars): string {
  const template = byCode.get(lang)?.messages[key] ?? fallback[key] ?? key;
  if (!vars) return template;
  return template.replace(/\{(\w+)\}/g, (m, name: string) => (name in vars ? String(vars[name]) : m));
}

const LANG_KEY = "im.lang";
const DIR_KEY = "im.dir";

function storedLang(): Lang {
  try {
    const stored = localStorage.getItem(LANG_KEY);
    return hasLang(stored) ? stored : DEFAULT_LANG;
  } catch {
    return DEFAULT_LANG;
  }
}

let fallbackLang: Lang = typeof window === "undefined" ? DEFAULT_LANG : storedLang();
let override: Lang | null = null;
const overrideListeners = new Set<() => void>();

export function setLangOverride(lang: Lang | null) {
  override = lang;
  for (const fn of overrideListeners) fn();
}

export function applyDocumentLang(lang: Lang) {
  const root = document.documentElement;
  root.lang = lang;
  root.dir = dirOf(lang);
  fallbackLang = lang;
  try {
    localStorage.setItem(LANG_KEY, lang);
    localStorage.setItem(DIR_KEY, dirOf(lang));
  } catch {}
}

function subscribeOverride(fn: () => void) {
  overrideListeners.add(fn);
  return () => {
    overrideListeners.delete(fn);
  };
}

export function useLang(): Lang {
  const server = useLive((s) => s.snapshot?.language);
  const pending = useSyncExternalStore(subscribeOverride, () => override);
  useEffect(() => {
    if (pending && server === pending) setLangOverride(null);
  }, [pending, server]);
  const lang = pending ?? server ?? fallbackLang;
  return hasLang(lang) ? lang : DEFAULT_LANG;
}

export function useI18n() {
  const lang = useLang();
  useEffect(() => applyDocumentLang(lang), [lang]);
  const t = useCallback<Translate>((key, vars) => translate(lang, key, vars), [lang]);
  return useMemo(() => ({ lang, t, locale: localeOf(lang), rtl: isRtl(lang) }), [lang, t]);
}
