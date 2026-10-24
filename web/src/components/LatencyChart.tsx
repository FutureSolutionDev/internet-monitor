import { type PointerEvent, useEffect, useMemo, useRef, useState } from "react";
import { formatNumber, formatTime } from "../lib/format";
import { isRtl, type Lang, type Translate } from "../lib/i18n";

interface Props {
  values: number[];
  lastCheck: number;
  intervalSec: number;
  lang: Lang;
  t: Translate;
  height?: number;
}

const PAD = { top: 12, bottom: 24, start: 40, end: 8 };

function niceMax(v: number): number {
  if (v <= 50) return 50;
  const pow = 10 ** Math.floor(Math.log10(v));
  for (const m of [1, 2, 2.5, 5, 10]) {
    if (m * pow >= v) return m * pow;
  }
  return 10 * pow;
}

export function LatencyChart({ values, lastCheck, intervalSec, lang, t, height = 200 }: Props) {
  const wrap = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(600);
  const [active, setActive] = useState<number | null>(null);
  const rtl = isRtl(lang);

  useEffect(() => {
    const el = wrap.current;
    if (!el || typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver((entries) => {
      const w = entries[0]?.contentRect.width;
      if (w) setWidth(Math.round(w));
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  const geo = useMemo(() => {
    const ok = values.filter((v) => v > 0);
    const max = niceMax(Math.max(...ok, 1) * 1.1);
    const plotW = Math.max(10, width - PAD.start - PAD.end);
    const plotH = height - PAD.top - PAD.bottom;
    const n = values.length;
    const xAt = (i: number) => {
      const p = n <= 1 ? plotW : (i / (n - 1)) * plotW;
      return rtl ? PAD.end + (plotW - p) : PAD.start + p;
    };
    const yAt = (v: number) => PAD.top + plotH - (Math.min(v, max) / max) * plotH;
    const segments: string[] = [];
    let current = "";
    values.forEach((v, i) => {
      if (v > 0) {
        current += `${current ? "L" : "M"}${xAt(i).toFixed(1)},${yAt(v).toFixed(1)}`;
      } else if (current) {
        segments.push(current);
        current = "";
      }
    });
    if (current) segments.push(current);
    const areas = segments.map((d) => {
      const pts = d.slice(1).split(/[LM]/);
      const first = pts[0]?.split(",")[0];
      const last = pts[pts.length - 1]?.split(",")[0];
      return `${d}L${last},${PAD.top + plotH}L${first},${PAD.top + plotH}Z`;
    });
    const ticks = [0, max / 2, max];
    return {
      max,
      plotW,
      plotH,
      xAt,
      yAt,
      segments,
      areas,
      ticks,
      avg: ok.length ? ok.reduce((a, b) => a + b, 0) / ok.length : 0,
      min: ok.length ? Math.min(...ok) : 0,
      peak: ok.length ? Math.max(...ok) : 0,
    };
  }, [values, width, height, rtl]);

  const n = values.length;
  const timeOf = (i: number) => lastCheck - (n - 1 - i) * intervalSec * 1000;

  const pick = (e: PointerEvent<SVGSVGElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    let p = rtl ? PAD.end + geo.plotW - x : x - PAD.start;
    p = Math.max(0, Math.min(geo.plotW, p));
    setActive(n <= 1 ? 0 : Math.round((p / geo.plotW) * (n - 1)));
  };

  const ms = t("ms_unit");
  const summary = `${t("chart_latency")}: ${formatNumber(lang, geo.min)}–${formatNumber(lang, geo.peak)} ${ms}, ${t("stat_avg", { value: `${formatNumber(lang, geo.avg)} ${ms}` })}`;
  const activeValue = active !== null ? values[active] : undefined;
  const tipX = active !== null ? geo.xAt(active) : 0;
  const tipLeft = Math.max(0, Math.min(width - 150, tipX - 75));

  return (
    <div className="chart" ref={wrap}>
      <svg
        height={height}
        viewBox={`0 0 ${width} ${height}`}
        role="img"
        aria-label={summary}
        onPointerMove={pick}
        onPointerLeave={() => setActive(null)}
      >
        {geo.ticks.map((tick) => (
          <g key={tick}>
            <line
              className="chart-grid"
              x1={rtl ? PAD.end : PAD.start}
              x2={rtl ? PAD.end + geo.plotW : PAD.start + geo.plotW}
              y1={geo.yAt(tick)}
              y2={geo.yAt(tick)}
            />
            <text
              className="chart-axis num"
              x={rtl ? width - 4 : PAD.start - 8}
              y={geo.yAt(tick) + 4}
              textAnchor="end"
            >
              {formatNumber(lang, tick)}
            </text>
          </g>
        ))}
        {geo.areas.map((d) => (
          <path key={`a${d}`} className="chart-area" d={d} />
        ))}
        {geo.segments.map((d) => (
          <path key={`l${d}`} className="chart-line" d={d} />
        ))}
        {values.map((v, i) =>
          v <= 0 ? (
            <rect
              // biome-ignore lint/suspicious/noArrayIndexKey: one marker per check slot
              key={i}
              className="chart-fail"
              x={geo.xAt(i) - 2}
              y={PAD.top + geo.plotH - 10}
              width={4}
              height={10}
              rx={2}
            />
          ) : null,
        )}
        {active !== null ? (
          <>
            <line className="chart-cursor" x1={tipX} x2={tipX} y1={PAD.top} y2={PAD.top + geo.plotH} />
            {activeValue && activeValue > 0 ? (
              <circle className="chart-dot" cx={tipX} cy={geo.yAt(activeValue)} r={5} />
            ) : null}
          </>
        ) : null}
        <text
          className="chart-axis"
          x={rtl ? PAD.end + geo.plotW : PAD.start}
          y={height - 4}
          textAnchor={rtl ? "end" : "start"}
        >
          {n ? formatTime(lang, timeOf(0), false) : ""}
        </text>
        <text
          className="chart-axis"
          x={rtl ? PAD.end : PAD.start + geo.plotW}
          y={height - 4}
          textAnchor={rtl ? "start" : "end"}
        >
          {n ? formatTime(lang, timeOf(n - 1), false) : ""}
        </text>
      </svg>
      {active !== null && activeValue !== undefined ? (
        <div className="tooltip" style={{ left: tipLeft, top: 4 }} aria-live="polite">
          <span className="muted">{formatTime(lang, timeOf(active))}</span>
          <strong className="num">
            {activeValue > 0 ? `${formatNumber(lang, activeValue)} ${ms}` : t("chart_failed")}
          </strong>
        </div>
      ) : null}
    </div>
  );
}
