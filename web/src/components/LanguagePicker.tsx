import { Languages } from "lucide-react";
import { api } from "../lib/api";
import { type Lang, LOCALES, setLangOverride, useI18n } from "../lib/i18n";
import { toast } from "./Toast";
import { Segmented } from "./ui";

export function useChangeLanguage() {
  const { t, lang } = useI18n();
  return async (next: Lang) => {
    if (next === lang) return;
    setLangOverride(next);
    try {
      await api.setLanguage(next);
    } catch {
      setLangOverride(null);
      toast(t("save_failed"), "error");
    }
  };
}

export function LanguageSelect({ compact = false }: { compact?: boolean }) {
  const { t, lang } = useI18n();
  const change = useChangeLanguage();
  return (
    <label className={compact ? "lang-select lang-select-compact" : "lang-select"}>
      <Languages size={16} aria-hidden />
      <span className="sr-only">{t("f_language")}</span>
      <select value={lang} onChange={(e) => void change(e.target.value)}>
        {LOCALES.map((l) => (
          <option key={l.code} value={l.code} lang={l.code}>
            {l.name}
          </option>
        ))}
      </select>
    </label>
  );
}

export function LanguageField() {
  const { t, lang } = useI18n();
  const change = useChangeLanguage();
  if (LOCALES.length > 3) return <LanguageSelect />;
  return (
    <Segmented<Lang>
      label={t("f_language")}
      value={lang}
      onChange={(next) => void change(next)}
      options={LOCALES.map((l) => ({ value: l.code, label: l.name }))}
    />
  );
}
