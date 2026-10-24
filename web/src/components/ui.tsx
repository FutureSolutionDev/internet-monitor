import type { LucideIcon } from "lucide-react";
import { LoaderCircle } from "lucide-react";
import {
  type ButtonHTMLAttributes,
  type CSSProperties,
  type MouseEvent,
  type PointerEvent,
  type ReactNode,
  type Ref,
  useRef,
  useState,
} from "react";
import { ChartTooltip, type TipRow } from "./ChartTooltip";

type Variant = "default" | "edge" | "aurora" | "solid" | "ghost" | "danger";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: "sm" | "md" | "lg";
  icon?: LucideIcon;
  busy?: boolean;
  iconOnly?: boolean;
  ref?: Ref<HTMLButtonElement>;
}

export function Button({
  variant = "default",
  size = "md",
  icon: Icon,
  busy,
  iconOnly,
  className,
  children,
  disabled,
  type = "button",
  ...rest
}: ButtonProps) {
  const classes = [
    "btn",
    variant !== "default" && `btn-${variant}`,
    size !== "md" && `btn-${size}`,
    iconOnly && "btn-icon",
    className,
  ]
    .filter(Boolean)
    .join(" ");
  const iconSize = size === "sm" ? 15 : 17;
  return (
    <button
      type={type}
      className={classes}
      disabled={disabled || busy}
      aria-busy={busy || undefined}
      {...rest}
    >
      {busy ? (
        <LoaderCircle size={iconSize} className="spin" aria-hidden />
      ) : Icon ? (
        <Icon size={iconSize} aria-hidden />
      ) : null}
      {children}
    </button>
  );
}

interface CardProps {
  children: ReactNode;
  className?: string;
  flare?: string;
  id?: string;
  as?: "section" | "div" | "article";
  labelledBy?: string;
}

export function Card({ children, className, flare, id, as: Tag = "section", labelledBy }: CardProps) {
  const style = flare ? ({ "--flare": flare } as CSSProperties) : undefined;
  return (
    <Tag
      id={id}
      className={["card", flare && "card-flare", className].filter(Boolean).join(" ")}
      style={style}
      aria-labelledby={labelledBy}
    >
      {children}
    </Tag>
  );
}

export function CardHead({
  title,
  sub,
  id,
  children,
}: {
  title: ReactNode;
  sub?: ReactNode;
  id?: string;
  children?: ReactNode;
}) {
  return (
    <div className="card-head">
      <div>
        <h2 id={id}>{title}</h2>
        {sub ? <p>{sub}</p> : null}
      </div>
      {children ? <div className="row">{children}</div> : null}
    </div>
  );
}

export function Banner({
  tone = "info",
  icon: Icon,
  children,
  actions,
}: {
  tone?: "info" | "warn" | "danger" | "ok";
  icon: LucideIcon;
  children: ReactNode;
  actions?: ReactNode;
}) {
  return (
    <div className={`banner ${tone}`} role={tone === "danger" || tone === "warn" ? "alert" : "status"}>
      <Icon size={18} aria-hidden />
      <div className="banner-body">{children}</div>
      {actions ? <div className="banner-actions">{actions}</div> : null}
    </div>
  );
}

export function Empty({
  icon: Icon,
  children,
  action,
}: {
  icon: LucideIcon;
  children: ReactNode;
  action?: ReactNode;
}) {
  return (
    <div className="empty">
      <Icon size={28} aria-hidden />
      <p>{children}</p>
      {action}
    </div>
  );
}

export function Skeleton({ height = 16, width = "100%" }: { height?: number; width?: number | string }) {
  return <div className="skeleton" style={{ height, width }} aria-hidden />;
}

export function Switch({
  checked,
  onChange,
  label,
  disabled,
  id,
  describedBy,
}: {
  checked: boolean;
  onChange: (next: boolean) => void;
  label: string;
  disabled?: boolean;
  id?: string;
  describedBy?: string;
}) {
  return (
    <button
      id={id}
      type="button"
      role="switch"
      className="switch"
      aria-checked={checked}
      aria-label={label}
      aria-describedby={describedBy}
      disabled={disabled}
      onClick={() => onChange(!checked)}
    />
  );
}

export function Segmented<T extends string>({
  value,
  options,
  onChange,
  label,
}: {
  value: T;
  options: { value: T; label: string }[];
  onChange: (v: T, e: MouseEvent<HTMLButtonElement>) => void;
  label: string;
}) {
  return (
    <fieldset className="segmented" aria-label={label}>
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          aria-pressed={o.value === value}
          onClick={(e) => onChange(o.value, e)}
        >
          {o.label}
        </button>
      ))}
    </fieldset>
  );
}

export function Meter({
  label,
  value,
  display,
  tone,
  tipTitle,
  tipRows,
}: {
  label: string;
  value: number | null;
  display: string;
  tone: string;
  tipTitle?: string;
  tipRows?: TipRow[];
}) {
  const pct = value === null ? 0 : Math.max(0, Math.min(100, value));
  const box = useRef<HTMLDivElement>(null);
  const track = useRef<HTMLDivElement>(null);
  const [tip, setTip] = useState<{ x: number; y: number; w: number } | null>(null);
  const move = (e: PointerEvent<HTMLDivElement>) => {
    const b = box.current?.getBoundingClientRect();
    const tr = track.current?.getBoundingClientRect();
    if (!b || !tr) return;
    setTip({ x: e.clientX - b.left, y: tr.top - b.top, w: b.width });
  };
  return (
    <div className="meter" ref={box} onPointerMove={move} onPointerLeave={() => setTip(null)}>
      {tip ? (
        <ChartTooltip
          x={tip.x}
          y={tip.y}
          bounds={tip.w}
          title={tipTitle ?? label}
          rows={tipRows ?? [{ label: label, value: display, tone }]}
        />
      ) : null}
      <div className="meter-top">
        <span className="secondary">{label}</span>
        <span className="num">{display}</span>
      </div>
      {/* biome-ignore lint/a11y/useSemanticElements: styled track needs a div; native meter cannot be themed consistently */}
      <div
        ref={track}
        className="meter-track"
        role="meter"
        aria-label={label}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={value === null ? undefined : Number(pct.toFixed(2))}
        aria-valuetext={display}
      >
        <div className="meter-fill" style={{ width: `${pct}%`, "--tone": tone } as CSSProperties} />
      </div>
    </div>
  );
}

export function Ltr({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <bdi className={["ltr", className].filter(Boolean).join(" ")} dir="ltr">
      {children}
    </bdi>
  );
}
