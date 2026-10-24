import { useCallback, useEffect, useMemo, useSyncExternalStore } from "react";
import about from "../../locales/en/about.json";
import changelog from "../../locales/en/changelog.json";
import common from "../../locales/en/common.json";
import faq from "../../locales/en/faq.json";
import history from "../../locales/en/history.json";
import legal from "../../locales/en/legal.json";
import notifications from "../../locales/en/notifications.json";
import overview from "../../locales/en/overview.json";
import report from "../../locales/en/report.json";
import server from "../../locales/en/server.json";
import settings from "../../locales/en/settings.json";
import speed from "../../locales/en/speed.json";
import { useLive } from "./live";

export const MESSAGE_FILES = {
  about,
  changelog,
  common,
  faq,
  history,
  legal,
  notifications,
  overview,
  report,
  server,
  settings,
  speed,
};

type Leaves<T, P extends string = ""> = {
  [K in keyof T & string]: T[K] extends string ? `${P}${K}` : Leaves<T[K], `${P}${K}.`>;
}[keyof T & string];

type Files = typeof MESSAGE_FILES;
export type MessageKey = { [F in keyof Files]: Leaves<Files[F]> }[keyof Files];
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

export interface MessageTree {
  [key: string]: string | MessageTree;
}

export const DEFAULT_LANG = "en";

export function flattenMessages(
  tree: MessageTree,
  where: string,
  out: Record<string, string> = {},
  prefix = "",
): Record<string, string> {
  for (const [name, value] of Object.entries(tree)) {
    const key = prefix + name;
    if (typeof value !== "string") {
      flattenMessages(value, where, out, `${key}.`);
      continue;
    }
    if (key in out) throw new Error(`${where}: "${key}" is defined more than once`);
    out[key] = value;
  }
  return out;
}

export function loadLocales(files: Record<string, unknown>): Locale[] {
  const byLang = new Map<Lang, { meta?: LocaleMeta; messages: Record<string, string> }>();
  for (const [path, data] of Object.entries(files)) {
    const match = /\/([^/]+)\/([^/]+)\.json$/.exec(path);
    if (!match) continue;
    const [, code = "", file = ""] = match;
    const entry = byLang.get(code) ?? { messages: {} };
    byLang.set(code, entry);
    if (file === "meta") entry.meta = data as LocaleMeta;
    else flattenMessages(data as MessageTree, code, entry.messages);
  }
  const list: Locale[] = [];
  for (const [code, { meta, messages }] of byLang) {
    if (!meta) throw new Error(`${code}: meta.json is missing`);
    list.push({ code, ...meta, messages: messages as Locale["messages"] });
  }
  return list.sort((a, b) =>
    a.code === DEFAULT_LANG ? -1 : b.code === DEFAULT_LANG ? 1 : a.code.localeCompare(b.code),
  );
}

export const LOCALE_FILES = import.meta.glob("../../locales/*/*.json", { eager: true, import: "default" });

export const LOCALES: Locale[] = loadLocales(LOCALE_FILES);

const byCode = new Map(LOCALES.map((l) => [l.code, l]));
const fallback = (byCode.get(DEFAULT_LANG)?.messages ?? {}) as Record<MessageKey, string>;

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
