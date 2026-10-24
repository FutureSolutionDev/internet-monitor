import { Fragment, type ReactNode } from "react";
import { hrefOf, isRoute } from "../lib/router";
import { goToTarget } from "../lib/spotlight";

const LINK = /\[([^\]]+)\]\(([a-z]+)(?:#([a-z0-9-]+))?\)/g;

export interface TextPart {
  text: string;
  route?: string;
  anchor?: string;
}

export function parseRichText(source: string): TextPart[] {
  const parts: TextPart[] = [];
  let last = 0;
  for (const m of source.matchAll(LINK)) {
    const at = m.index ?? 0;
    if (at > last) parts.push({ text: source.slice(last, at) });
    const [, text = "", route = "", anchor] = m;
    parts.push(isRoute(route) ? { text, route, anchor } : { text });
    last = at + m[0].length;
  }
  if (last < source.length) parts.push({ text: source.slice(last) });
  return parts;
}

export function plainText(source: string): string {
  return parseRichText(source)
    .map((p) => p.text)
    .join("");
}

export function RichText({ text }: { text: string }) {
  const out: ReactNode[] = parseRichText(text).map((p, i) => {
    const key = `${i}-${p.text}`;
    if (!p.route || !isRoute(p.route)) return <Fragment key={key}>{p.text}</Fragment>;
    const route = p.route;
    const href = hrefOf(route) + (p.anchor ? `#${p.anchor}` : "");
    return (
      <a
        key={key}
        className="text-link"
        href={href}
        onClick={(e) => {
          if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
          e.preventDefault();
          goToTarget(route, p.anchor);
        }}
      >
        {p.text}
      </a>
    );
  });
  return <>{out}</>;
}
