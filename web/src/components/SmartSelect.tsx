import { Check, ChevronDown, Search } from "lucide-react";
import {
  type CSSProperties,
  type KeyboardEvent,
  type ReactNode,
  useCallback,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";
import { useI18n } from "../lib/i18n";

export interface SmartOption<T extends string = string> {
  value: T;
  label: string;
  detail?: string;
  icon?: ReactNode;
  lang?: string;
  keywords?: string;
}

interface Props<T extends string> {
  options: SmartOption<T>[];
  value: T | null;
  onChange: (value: T) => void;
  label: string;
  id?: string;
  placeholder?: string;
  searchable?: boolean;
  disabled?: boolean;
  variant?: "field" | "pill";
  leading?: ReactNode;
  iconOnly?: boolean;
  className?: string;
  describedBy?: string;
}

interface PanelPos {
  top?: number;
  bottom?: number;
  left?: number;
  right?: number;
  width: number;
  maxHeight: number;
}

const SEARCH_THRESHOLD = 7;
const GAP = 6;
const MARGIN = 8;
const DESIRED_HEIGHT = 320;
const MIN_WIDTH = 200;
const MIN_ROOM = 160;

function normalize(text: string): string {
  return text.toLocaleLowerCase().normalize("NFKD").replace(/[ً-ٰٟ]/g, "");
}

export function SmartSelect<T extends string>({
  options,
  value,
  onChange,
  label,
  id,
  placeholder,
  searchable,
  disabled = false,
  variant = "field",
  leading,
  iconOnly = false,
  className,
  describedBy,
}: Props<T>) {
  const { t, rtl } = useI18n();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [active, setActive] = useState(0);
  const [pos, setPos] = useState<PanelPos | null>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLDivElement>(null);
  const search = useRef<HTMLInputElement>(null);
  const list = useRef<HTMLDivElement>(null);
  const typed = useRef({ text: "", at: 0 });
  const baseId = useId();
  const listId = `${baseId}-list`;
  const optionId = (i: number) => `${baseId}-opt-${i}`;
  const withSearch = searchable ?? options.length > SEARCH_THRESHOLD;

  const current = options.find((o) => o.value === value);
  const filtered = useMemo(() => {
    const q = normalize(query.trim());
    if (!q) return options;
    return options.filter((o) => normalize(`${o.label} ${o.detail ?? ""} ${o.keywords ?? ""}`).includes(q));
  }, [options, query]);

  const place = useCallback(() => {
    const el = trigger.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    const vw = document.documentElement.clientWidth || window.innerWidth;
    const vh = document.documentElement.clientHeight || window.innerHeight;
    if (r.bottom < 0 || r.top > vh) {
      setOpen(false);
      setQuery("");
      return;
    }
    const width = Math.max(0, Math.min(Math.max(r.width, MIN_WIDTH), vw - MARGIN * 2));
    const below = vh - r.bottom - GAP - MARGIN;
    const above = r.top - GAP - MARGIN;
    const next: PanelPos = { width, maxHeight: 0 };
    if (below >= Math.min(DESIRED_HEIGHT, MIN_ROOM) || below >= above) {
      next.top = Math.max(MARGIN, r.bottom + GAP);
      next.maxHeight = below;
    } else {
      next.bottom = Math.max(MARGIN, vh - r.top + GAP);
      next.maxHeight = above;
    }
    if (next.maxHeight < MIN_ROOM) {
      next.top = MARGIN;
      next.bottom = undefined;
      next.maxHeight = vh - MARGIN * 2;
    }
    next.maxHeight = Math.max(0, Math.min(DESIRED_HEIGHT, next.maxHeight));
    if (rtl) next.right = Math.max(MARGIN, Math.min(vw - r.right, vw - MARGIN - width));
    else next.left = Math.max(MARGIN, Math.min(r.left, vw - MARGIN - width));
    setPos(next);
  }, [rtl]);

  const close = useCallback((refocus: boolean) => {
    setOpen(false);
    setQuery("");
    if (refocus) trigger.current?.focus();
  }, []);

  useLayoutEffect(() => {
    if (!open) {
      setPos(null);
      return;
    }
    place();
    const onMove = () => place();
    window.addEventListener("resize", onMove);
    window.addEventListener("scroll", onMove, true);
    return () => {
      window.removeEventListener("resize", onMove);
      window.removeEventListener("scroll", onMove, true);
    };
  }, [open, place]);

  useEffect(() => {
    if (!open) return;
    const onDown = (e: PointerEvent) => {
      const target = e.target as Node;
      if (trigger.current?.contains(target) || panel.current?.contains(target)) return;
      close(false);
    };
    document.addEventListener("pointerdown", onDown);
    return () => document.removeEventListener("pointerdown", onDown);
  }, [open, close]);

  useEffect(() => {
    if (!open || !pos) return;
    const coarse = typeof window.matchMedia === "function" && window.matchMedia("(pointer: coarse)").matches;
    (withSearch && !coarse ? search.current : list.current)?.focus({ preventScroll: true });
  }, [open, pos, withSearch]);

  useEffect(() => {
    if (!open) return;
    list.current?.querySelector<HTMLElement>(`[data-idx="${active}"]`)?.scrollIntoView({ block: "nearest" });
  }, [active, open]);

  const openList = () => {
    if (disabled) return;
    const selected = options.findIndex((o) => o.value === value);
    setActive(Math.max(selected, 0));
    setOpen(true);
  };

  const commit = (option: SmartOption<T> | undefined) => {
    if (!option) return;
    close(true);
    if (option.value !== value) onChange(option.value);
  };

  const typeahead = (key: string) => {
    const now = Date.now();
    typed.current.text = now - typed.current.at > 700 ? key : typed.current.text + key;
    typed.current.at = now;
    const needle = normalize(typed.current.text);
    const hit = filtered.findIndex((o) => normalize(o.label).startsWith(needle));
    if (hit >= 0) setActive(hit);
  };

  const onListKey = (e: KeyboardEvent) => {
    const last = filtered.length - 1;
    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        setActive((a) => Math.min(a + 1, last));
        break;
      case "ArrowUp":
        e.preventDefault();
        setActive((a) => Math.max(a - 1, 0));
        break;
      case "Home":
        if (withSearch) return;
        e.preventDefault();
        setActive(0);
        break;
      case "End":
        if (withSearch) return;
        e.preventDefault();
        setActive(last);
        break;
      case "PageDown":
        e.preventDefault();
        setActive((a) => Math.min(a + 8, last));
        break;
      case "PageUp":
        e.preventDefault();
        setActive((a) => Math.max(a - 8, 0));
        break;
      case "Enter":
      case " ":
        if (e.key === " " && withSearch) return;
        e.preventDefault();
        commit(filtered[active]);
        break;
      case "Escape":
        e.preventDefault();
        e.stopPropagation();
        close(true);
        break;
      case "Tab":
        close(true);
        break;
      default:
        if (!withSearch && e.key.length === 1 && !e.ctrlKey && !e.metaKey && !e.altKey) typeahead(e.key);
    }
  };

  const onTriggerKey = (e: KeyboardEvent) => {
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      e.preventDefault();
      openList();
    }
  };

  const activeId = filtered[active] ? optionId(active) : undefined;
  const style: CSSProperties | undefined = pos
    ? {
        top: pos.top,
        bottom: pos.bottom,
        left: pos.left,
        right: pos.right,
        width: pos.width,
        maxHeight: pos.maxHeight,
      }
    : undefined;

  return (
    <div className={["ss", `ss-${variant}`, iconOnly && "ss-icon-only", className].filter(Boolean).join(" ")}>
      <button
        ref={trigger}
        type="button"
        id={id}
        className="ss-trigger"
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={open ? listId : undefined}
        aria-describedby={describedBy}
        title={iconOnly ? `${label}: ${current?.label ?? ""}` : undefined}
        onClick={() => (open ? close(true) : openList())}
        onKeyDown={onTriggerKey}
      >
        <span className="sr-only">{label}: </span>
        {leading ? <span className="ss-leading">{leading}</span> : null}
        {current?.icon ? <span className="ss-icon">{current.icon}</span> : null}
        <span
          className={iconOnly ? "sr-only" : current ? "ss-value" : "ss-value ss-placeholder"}
          lang={current?.lang}
        >
          {current?.label ?? placeholder ?? t("select_placeholder")}
        </span>
        {iconOnly ? null : <ChevronDown size={16} className="ss-chevron" aria-hidden />}
      </button>

      {open && pos
        ? createPortal(
            <div
              ref={panel}
              className="ss-panel"
              style={style}
              data-up={pos.bottom !== undefined || undefined}
            >
              {withSearch ? (
                <div className="ss-search">
                  <Search size={15} aria-hidden />
                  <input
                    ref={search}
                    type="text"
                    role="combobox"
                    aria-expanded="true"
                    aria-controls={listId}
                    aria-activedescendant={activeId}
                    aria-autocomplete="list"
                    aria-label={t("select_search")}
                    placeholder={t("select_search")}
                    value={query}
                    onChange={(e) => {
                      setQuery(e.target.value);
                      setActive(0);
                    }}
                    onKeyDown={onListKey}
                  />
                </div>
              ) : null}
              <div
                ref={list}
                id={listId}
                role="listbox"
                aria-label={label}
                tabIndex={withSearch ? -1 : 0}
                aria-activedescendant={withSearch ? undefined : activeId}
                className="ss-list"
                onKeyDown={withSearch ? undefined : onListKey}
              >
                {filtered.length === 0 ? (
                  <div className="ss-empty">{t("select_empty")}</div>
                ) : (
                  filtered.map((o, i) => {
                    const selected = o.value === value;
                    return (
                      // biome-ignore lint/a11y/useKeyWithClickEvents: keyboard is handled by the listbox through aria-activedescendant
                      <div
                        key={o.value}
                        id={optionId(i)}
                        data-idx={i}
                        role="option"
                        tabIndex={-1}
                        aria-selected={selected}
                        data-active={i === active || undefined}
                        className="ss-option"
                        lang={o.lang}
                        onPointerMove={() => setActive(i)}
                        onPointerDown={(e) => e.preventDefault()}
                        onClick={() => commit(o)}
                      >
                        <span className="ss-check" aria-hidden>
                          {selected ? <Check size={15} /> : null}
                        </span>
                        {o.icon ? <span className="ss-icon">{o.icon}</span> : null}
                        <span className="ss-text">
                          <span className="ss-label">{o.label}</span>
                          {o.detail ? <span className="ss-detail">{o.detail}</span> : null}
                        </span>
                      </div>
                    );
                  })
                )}
              </div>
            </div>,
            document.body,
          )
        : null}
    </div>
  );
}
