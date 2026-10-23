import { CircleAlert, CircleCheck, CircleX, Plus, Trash2 } from "lucide-react";
import { type ReactNode, useId, useState } from "react";
import { Button } from "./ui";

interface FieldProps {
  label: string;
  hint?: ReactNode;
  error?: string;
  className?: string;
  children: (ids: { inputId: string; describedBy: string | undefined; invalid: boolean }) => ReactNode;
}

export function Field({ label, hint, error, className, children }: FieldProps) {
  const inputId = useId();
  const hintId = useId();
  const errorId = useId();
  const describedBy = [hint ? hintId : "", error ? errorId : ""].filter(Boolean).join(" ") || undefined;
  return (
    <div className={["field", className].filter(Boolean).join(" ")}>
      <label className="field-label" htmlFor={inputId}>
        {label}
      </label>
      {children({ inputId, describedBy, invalid: !!error })}
      {error ? (
        <span className="field-error" id={errorId}>
          <CircleAlert size={13} aria-hidden />
          {error}
        </span>
      ) : null}
      {hint ? (
        <span className="field-hint" id={hintId}>
          {hint}
        </span>
      ) : null}
    </div>
  );
}

interface TextFieldProps {
  label: string;
  hint?: ReactNode;
  error?: string;
  value: string;
  onChange: (v: string) => void;
  suffix?: string;
  inputMode?: "numeric" | "decimal" | "text" | "url";
  ltr?: boolean;
  narrow?: boolean;
  placeholder?: string;
  className?: string;
  disabled?: boolean;
  type?: "text" | "password" | "url";
  autoComplete?: string;
}

export function TextField({
  label,
  hint,
  error,
  value,
  onChange,
  suffix,
  inputMode,
  ltr,
  narrow,
  placeholder,
  className,
  disabled,
  type = "text",
  autoComplete = "off",
}: TextFieldProps) {
  return (
    <Field label={label} hint={hint} error={error} className={className}>
      {({ inputId, describedBy, invalid }) => (
        <div className={["input-wrap", narrow && "input-narrow"].filter(Boolean).join(" ")}>
          <input
            id={inputId}
            className={["input", suffix && "has-suffix", ltr && "ltr"].filter(Boolean).join(" ")}
            dir={ltr ? "ltr" : undefined}
            type={type}
            value={value}
            inputMode={inputMode}
            placeholder={placeholder}
            disabled={disabled}
            spellCheck={false}
            autoComplete={autoComplete}
            aria-invalid={invalid || undefined}
            aria-describedby={describedBy}
            onChange={(e) => onChange(e.target.value)}
          />
          {suffix ? (
            <span className="input-suffix" aria-hidden>
              {suffix}
            </span>
          ) : null}
        </div>
      )}
    </Field>
  );
}

export interface ItemResult {
  ok: boolean;
  text: string;
}

interface ListEditorProps {
  label: string;
  hint?: ReactNode;
  error?: string;
  items: string[];
  onChange: (items: string[]) => void;
  max: number;
  placeholder: string;
  addLabel: string;
  removeLabel: (item: string) => string;
  results?: Record<string, ItemResult>;
  className?: string;
}

export function ListEditor({
  label,
  hint,
  error,
  items,
  onChange,
  max,
  placeholder,
  addLabel,
  removeLabel,
  results,
  className,
}: ListEditorProps) {
  const [draft, setDraft] = useState("");
  const add = () => {
    const v = draft.trim();
    if (!v || items.length >= max) return;
    onChange([...items, v]);
    setDraft("");
  };
  return (
    <Field label={label} hint={hint} error={error} className={className}>
      {({ inputId, describedBy, invalid }) => (
        <div className="list-editor">
          {items.map((item, i) => {
            const r = results?.[item.trim()];
            return (
              // biome-ignore lint/suspicious/noArrayIndexKey: rows are positional and may hold duplicates while editing
              <div className="list-editor-row" key={i}>
                <input
                  className="input ltr"
                  dir="ltr"
                  value={item}
                  spellCheck={false}
                  autoComplete="off"
                  aria-label={`${label} ${i + 1}`}
                  aria-invalid={invalid || undefined}
                  onChange={(e) => onChange(items.map((v, j) => (j === i ? e.target.value : v)))}
                />
                {r ? (
                  <span
                    className="result num"
                    style={{ color: r.ok ? "var(--accent-green)" : "var(--accent-coral)" }}
                  >
                    {r.ok ? (
                      <CircleCheck size={14} aria-hidden style={{ verticalAlign: "-2px" }} />
                    ) : (
                      <CircleX size={14} aria-hidden style={{ verticalAlign: "-2px" }} />
                    )}{" "}
                    {r.text}
                  </span>
                ) : null}
                <Button
                  variant="ghost"
                  size="sm"
                  iconOnly
                  icon={Trash2}
                  aria-label={removeLabel(item || String(i + 1))}
                  onClick={() => onChange(items.filter((_, j) => j !== i))}
                />
              </div>
            );
          })}
          {items.length < max ? (
            <div className="list-editor-row">
              <input
                id={inputId}
                className="input ltr"
                dir="ltr"
                value={draft}
                placeholder={placeholder}
                spellCheck={false}
                autoComplete="off"
                aria-describedby={describedBy}
                onChange={(e) => setDraft(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    add();
                  }
                }}
                onBlur={add}
              />
              <Button variant="edge" size="sm" icon={Plus} onClick={add} disabled={!draft.trim()}>
                {addLabel}
              </Button>
            </div>
          ) : null}
        </div>
      )}
    </Field>
  );
}
