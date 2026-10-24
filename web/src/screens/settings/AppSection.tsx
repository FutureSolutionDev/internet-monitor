import { RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";
import { TextField } from "../../components/Field";
import { LanguageField } from "../../components/LanguagePicker";
import { toast } from "../../components/Toast";
import { Button, Card, CardHead, Segmented, Switch } from "../../components/ui";
import { api } from "../../lib/api";
import { useI18n } from "../../lib/i18n";
import { useLive } from "../../lib/live";
import { type Theme, useTheme } from "../../lib/theme";
import { changeThemeWithTransition } from "../../lib/themeTransition";
import type { StartupState } from "../../lib/types";
import type { SectionProps } from "./Sections";

function StartupRow() {
  const { t } = useI18n();
  const [state, setState] = useState<StartupState | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api.startup().then(setState, () => setState({ supported: false, enabled: false }));
  }, []);

  const toggle = async (next: boolean) => {
    setBusy(true);
    try {
      const r = await api.setStartup(next);
      setState((s) => ({ supported: s?.supported ?? true, enabled: r.enabled }));
      toast(t(r.enabled ? "startup_on" : "startup_off"), "ok");
    } catch {
      toast(t("save_failed"), "error");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="switch-row full">
      <div className="field">
        <span className="field-label">{t("f_startup")}</span>
        <span className="field-hint" id="startup-hint">
          {state && !state.supported ? t("f_startup_unsupported") : t("f_startup_hint")}
        </span>
      </div>
      <Switch
        checked={state?.enabled ?? false}
        disabled={!state?.supported || busy}
        onChange={toggle}
        label={t("f_startup")}
        describedBy="startup-hint"
      />
    </div>
  );
}

function UpdatesRow() {
  const { t } = useI18n();
  const version = useLive((s) => s.snapshot?.version ?? "");
  const [checking, setChecking] = useState(false);

  const check = async () => {
    setChecking(true);
    try {
      const info = await api.checkUpdate();
      if (info.has_update) toast(t("update_available", { v: info.latest_version }), "info");
      else toast(t("update_latest"), "ok");
    } catch {
      toast(t("update_check_failed"), "error");
    } finally {
      setChecking(false);
    }
  };

  return (
    <div className="switch-row full">
      <div className="field">
        <span className="field-label">{t("updates")}</span>
        <span className="field-hint num">{t("version_current", { v: version })}</span>
      </div>
      <Button variant="edge" size="sm" icon={RefreshCw} busy={checking} onClick={check}>
        {checking ? t("update_checking") : t("update_check")}
      </Button>
    </div>
  );
}

export function AppSection({ draft, set, err }: SectionProps) {
  const { t } = useI18n();
  const theme = useTheme();

  return (
    <Card id="app" className="settings-section" labelledBy="app-title">
      <CardHead id="app-title" title={t("sec_app")} />
      <div className="form-grid">
        <div className="field">
          <span className="field-label">{t("f_language")}</span>
          <LanguageField />
        </div>
        <div className="field">
          <span className="field-label">{t("f_theme")}</span>
          <Segmented<Theme>
            label={t("f_theme")}
            value={theme}
            onChange={(next, e) => changeThemeWithTransition(next, e)}
            options={[
              { value: "dark", label: t("theme_dark") },
              { value: "light", label: t("theme_light") },
            ]}
          />
        </div>
        <StartupRow />
        <TextField
          label={t("f_port")}
          hint={t("f_port_hint")}
          error={err("dashboard_port")}
          value={draft.dashboard_port}
          onChange={(v) => set("dashboard_port", v)}
          inputMode="numeric"
          narrow
        />
        <TextField
          label={t("f_logdir")}
          hint={t("f_logdir_hint")}
          error={err("log_dir")}
          value={draft.log_dir}
          onChange={(v) => set("log_dir", v)}
          ltr
        />
        <hr className="divider full" />
        <UpdatesRow />
      </div>
    </Card>
  );
}
