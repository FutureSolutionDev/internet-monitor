import type { CSSProperties } from "react";

const STOPS = [0, 5, 10, 25, 50, 100, 250, 500, 1000];
const SWEEP = 240;
const R = 120;
const CX = 150;
const CY = 150;

export function gaugeFraction(mbps: number): number {
  if (!Number.isFinite(mbps) || mbps <= 0) return 0;
  const last = STOPS.length - 1;
  for (let i = 1; i <= last; i++) {
    const hi = STOPS[i] as number;
    if (mbps <= hi) {
      const lo = STOPS[i - 1] as number;
      return (i - 1 + (mbps - lo) / (hi - lo)) / last;
    }
  }
  return 1;
}

function point(angleDeg: number, radius: number) {
  const a = ((angleDeg - 90) * Math.PI) / 180;
  return { x: CX + radius * Math.cos(a), y: CY + radius * Math.sin(a) };
}

function arc(from: number, to: number) {
  const s = point(from, R);
  const e = point(to, R);
  const large = to - from > 180 ? 1 : 0;
  return `M${s.x.toFixed(2)},${s.y.toFixed(2)} A${R},${R} 0 ${large} 1 ${e.x.toFixed(2)},${e.y.toFixed(2)}`;
}

interface Props {
  value: number;
  display: string;
  unit: string;
  label: string;
  tone: string;
}

export function Gauge({ value, display, unit, label, tone }: Props) {
  const start = -SWEEP / 2;
  const path = arc(start, start + SWEEP);
  const length = (SWEEP / 360) * 2 * Math.PI * R;
  const frac = gaugeFraction(value);
  return (
    <div className="gauge" style={{ "--tone": tone } as CSSProperties}>
      <svg viewBox="0 0 300 260" role="img" aria-label={`${label}: ${display} ${unit}`}>
        <path className="gauge-track" d={path} strokeWidth={14} />
        <path
          className="gauge-fill"
          d={path}
          strokeWidth={14}
          strokeDasharray={length}
          strokeDashoffset={length * (1 - frac)}
          opacity={frac > 0 ? 1 : 0}
        />
        {STOPS.map((s, i) => {
          const p = point(start + (SWEEP * i) / (STOPS.length - 1), R - 28);
          return (
            <text key={s} x={p.x} y={p.y + 4} textAnchor="middle" className="chart-axis num">
              {s}
            </text>
          );
        })}
      </svg>
      <div className="gauge-center" aria-hidden>
        <span className="gauge-value">{display}</span>
        <span className="gauge-unit">{unit}</span>
      </div>
    </div>
  );
}
