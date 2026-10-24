import type { Lang } from "../lib/i18n";

export interface Text {
  en: string;
  ar: string;
}

export function pick(text: Text, lang: Lang): string {
  return text[lang];
}

export const REPO_URL = "https://github.com/FutureSolutionDev/internet-monitor";
