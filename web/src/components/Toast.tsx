import { CircleAlert, CircleCheck, Info, X } from "lucide-react";
import { useSyncExternalStore } from "react";

export type ToastTone = "ok" | "error" | "info";
interface ToastItem {
  id: number;
  tone: ToastTone;
  text: string;
}

let items: ToastItem[] = [];
let nextId = 1;
const listeners = new Set<() => void>();
const emit = () => {
  for (const fn of listeners) fn();
};

export function dismissToast(id: number) {
  items = items.filter((t) => t.id !== id);
  emit();
}

export function toast(text: string, tone: ToastTone = "info", ms = tone === "error" ? 7000 : 4000) {
  const id = nextId++;
  items = [...items.slice(-3), { id, tone, text }];
  emit();
  setTimeout(() => dismissToast(id), ms);
}

function subscribe(fn: () => void) {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}

export function Toaster({ closeLabel }: { closeLabel: string }) {
  const list = useSyncExternalStore(subscribe, () => items);
  return (
    <div className="toasts" aria-live="polite" aria-relevant="additions">
      {list.map((t) => {
        const Icon = t.tone === "ok" ? CircleCheck : t.tone === "error" ? CircleAlert : Info;
        return (
          <div key={t.id} className={`toast ${t.tone}`} role={t.tone === "error" ? "alert" : "status"}>
            <Icon size={18} aria-hidden />
            <span style={{ flex: 1 }}>{t.text}</span>
            <button
              type="button"
              className="btn btn-ghost btn-sm btn-icon"
              aria-label={closeLabel}
              onClick={() => dismissToast(t.id)}
            >
              <X size={14} aria-hidden />
            </button>
          </div>
        );
      })}
    </div>
  );
}
