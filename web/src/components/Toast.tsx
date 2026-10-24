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

const timers = new Map<number, ReturnType<typeof setTimeout>>();
let paused = false;
const AUTO_MS = 5000;

function arm(item: ToastItem) {
  if (item.tone === "error" || paused || timers.has(item.id)) return;
  timers.set(
    item.id,
    setTimeout(() => dismissToast(item.id), AUTO_MS),
  );
}

export function dismissToast(id: number) {
  const timer = timers.get(id);
  if (timer) clearTimeout(timer);
  timers.delete(id);
  items = items.filter((t) => t.id !== id);
  if (items.length === 0) paused = false;
  emit();
}

function setPaused(next: boolean) {
  paused = next;
  if (next) {
    for (const timer of timers.values()) clearTimeout(timer);
    timers.clear();
  } else {
    for (const item of items) arm(item);
  }
}

export function toast(text: string, tone: ToastTone = "info") {
  const id = nextId++;
  const dropped = items.slice(0, -3);
  items = [...items.slice(-3), { id, tone, text }];
  for (const d of dropped) {
    const timer = timers.get(d.id);
    if (timer) clearTimeout(timer);
    timers.delete(d.id);
  }
  emit();
  const item = items[items.length - 1];
  if (item) arm(item);
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
    // biome-ignore lint/a11y/noStaticElementInteractions: hover and focus only pause auto-dismiss; no action is triggered
    <div
      className="toasts"
      aria-live="polite"
      aria-relevant="additions"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
      onFocus={() => setPaused(true)}
      onBlur={() => setPaused(false)}
    >
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
