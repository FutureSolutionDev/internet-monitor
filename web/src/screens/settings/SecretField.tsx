import { KeyRound, Undo2 } from "lucide-react";
import type { ReactNode } from "react";
import { Field } from "../../components/Field";
import { Button } from "../../components/ui";
import type { Translate } from "../../lib/i18n";
import type { SecretDraft } from "../../lib/settingsForm";

interface Props {
  label: string;
  hint?: ReactNode;
  error?: string;
  saved: boolean;
  value: SecretDraft;
  onChange: (v: SecretDraft) => void;
  placeholder: string;
  masked?: boolean;
  t: Translate;
}

export function SecretField({ label, hint, error, saved, value, onChange, placeholder, masked, t }: Props) {
  return (
    <Field label={label} hint={hint} error={error} className="full">
      {({ inputId, describedBy, invalid }) => {
        if (saved && value.mode === "keep") {
          return (
            <div className="row">
              <span className="pill" id={inputId}>
                <KeyRound size={14} aria-hidden />
                {t("secret_saved")}
              </span>
              <Button size="sm" variant="edge" onClick={() => onChange({ mode: "set", value: "" })}>
                {t("secret_replace")}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => onChange({ mode: "clear" })}>
                {t("secret_remove")}
              </Button>
            </div>
          );
        }
        if (value.mode === "clear") {
          return (
            <div className="row">
              <span className="pill" id={inputId} style={{ color: "var(--accent-coral)" }}>
                {t("secret_will_remove")}
              </span>
              <Button size="sm" variant="ghost" icon={Undo2} onClick={() => onChange({ mode: "keep" })}>
                {t("secret_keep")}
              </Button>
            </div>
          );
        }
        return (
          <div className="row" style={{ flexWrap: "nowrap" }}>
            <input
              id={inputId}
              className="input ltr"
              dir="ltr"
              type={masked ? "password" : "text"}
              value={value.mode === "set" ? value.value : ""}
              placeholder={placeholder}
              spellCheck={false}
              autoComplete="off"
              aria-invalid={invalid || undefined}
              aria-describedby={describedBy}
              onChange={(e) =>
                onChange(e.target.value || saved ? { mode: "set", value: e.target.value } : { mode: "keep" })
              }
            />
            {saved ? (
              <Button size="sm" variant="ghost" icon={Undo2} onClick={() => onChange({ mode: "keep" })}>
                {t("secret_keep")}
              </Button>
            ) : null}
          </div>
        );
      }}
    </Field>
  );
}
