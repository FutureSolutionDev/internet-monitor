import { useCallback, useEffect, useMemo, useSyncExternalStore } from "react";
import { useLive } from "./live";
import { ar, en, type MessageKey } from "./strings";

export type Lang = "ar" | "en";
export type Vars = Record<string, string | number>;
export type Translate = (key: MessageKey, vars?: Vars) => string;

const LANG_KEY = "im.lang";

function storedLang(): Lang {
  try {
    return localStorage.getItem(LANG_KEY) === "ar" ? "ar" : "en";
  } catch {
    return "en";
  }
}

let fallbackLang: Lang = typeof window === "undefined" ? "en" : storedLang();
let override: Lang | null = null;
const overrideListeners = new Set<() => void>();

export function setLangOverride(lang: Lang | null) {
  override = lang;
  for (const fn of overrideListeners) fn();
}

export function translate(lang: Lang, key: MessageKey, vars?: Vars): string {
  const dict = lang === "ar" ? ar : en;
  const template = dict[key] ?? en[key] ?? key;
  if (!vars) return template;
  return template.replace(/\{(\w+)\}/g, (m, name: string) => (name in vars ? String(vars[name]) : m));
}

export function localeOf(lang: Lang): string {
  return lang === "ar" ? "ar-EG-u-nu-latn" : "en-GB";
}

export function applyDocumentLang(lang: Lang) {
  const root = document.documentElement;
  root.lang = lang;
  root.dir = lang === "ar" ? "rtl" : "ltr";
  fallbackLang = lang;
  try {
    localStorage.setItem(LANG_KEY, lang);
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
  return pending ?? server ?? fallbackLang;
}

export function useI18n() {
  const lang = useLang();
  useEffect(() => applyDocumentLang(lang), [lang]);
  const t = useCallback<Translate>((key, vars) => translate(lang, key, vars), [lang]);
  return useMemo(() => ({ lang, t, locale: localeOf(lang) }), [lang, t]);
}
