import { type CSSProperties, useLayoutEffect, useRef, useState } from "react";

export interface TipRow {
  label: string;
  value: string;
  tone?: string;
}

interface Props {
  id?: string;
  x: number;
  y: number;
  bounds: number;
  title: string;
  rows: TipRow[];
}

const OFFSET = 12;

const VIEWPORT_MARGIN = 8;

export function placeTip(x: number, y: number, bounds: number, w: number, h: number, originTop = 0) {
  const left = Math.max(0, Math.min(bounds - w, x - w / 2));
  const above = y - h - OFFSET;
  const flipped = originTop + above < VIEWPORT_MARGIN;
  return { left, top: flipped ? y + OFFSET : above, flipped };
}

export function ChartTooltip({ id, x, y, bounds, title, rows }: Props) {
  const ref = useRef<HTMLDivElement>(null);
  const [size, setSize] = useState<{ w: number; h: number; top: number } | null>(null);

  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;
    const w = el.offsetWidth;
    const h = el.offsetHeight;
    const top = Math.round(el.offsetParent?.getBoundingClientRect().top ?? 0);
    setSize((s) => (s && s.w === w && s.h === h && s.top === top ? s : { w, h, top }));
  });

  const pos = size ? placeTip(x, y, bounds, size.w, size.h, size.top) : null;
  const style: CSSProperties = pos
    ? { left: pos.left, top: pos.top }
    : { left: 0, top: 0, visibility: "hidden" };

  return (
    <div
      ref={ref}
      id={id}
      className="chart-tip"
      role="tooltip"
      style={style}
      data-below={pos?.flipped || undefined}
    >
      <div className="chart-tip-title">{title}</div>
      {rows.map((r) => (
        <div key={r.label} className="chart-tip-row">
          {r.tone ? <span className="dot" style={{ "--tone": r.tone } as CSSProperties} aria-hidden /> : null}
          <span className="chart-tip-label">{r.label}</span>
          <span className="chart-tip-value num">{r.value}</span>
        </div>
      ))}
    </div>
  );
}
