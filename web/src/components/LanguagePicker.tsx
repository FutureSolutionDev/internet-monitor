import { Languages } from "lucide-react";
import { api } from "../lib/api";
import { type Lang, LOCALES, setLangOverride, useI18n } from "../lib/i18n";
import { SmartSelect } from "./SmartSelect";
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

export function LanguageSelect({ iconOnly = false }: { iconOnly?: boolean }) {
  const { t, lang } = useI18n();
  const change = useChangeLanguage();
  return (
    <SmartSelect<Lang>
      variant="pill"
      iconOnly={iconOnly}
      label={t("f_language")}
      leading={<Languages size={16} aria-hidden />}
      value={lang}
      onChange={(next) => void change(next)}
      options={LOCALES.map((l) => ({ value: l.code, label: l.name, lang: l.code, keywords: l.code }))}
    />
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
