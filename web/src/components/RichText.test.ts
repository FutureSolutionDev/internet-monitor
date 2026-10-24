import { describe, expect, test } from "vitest";
import ar from "../../locales/ar.json";
import en from "../../locales/en.json";
import { FAQ } from "../content/faq";
import { parseRichText, plainText } from "./RichText";

const ANCHORS = new Set([
  "monitoring",
  "targets",
  "alerts",
  "notifications",
  "sound",
  "speed",
  "app",
  "n-sound",
  "n-quiet",
  "n-min-outage",
  "n-reminder",
]);

describe("rich text", () => {
  test("splits text and in-app links", () => {
    expect(parseRichText("Open [Settings → Alerts](settings#alerts) now.")).toEqual([
      { text: "Open " },
      { text: "Settings → Alerts", route: "settings", anchor: "alerts" },
      { text: " now." },
    ]);
  });

  test("unknown routes stay plain text", () => {
    expect(parseRichText("[x](nowhere#a)")).toEqual([{ text: "x" }]);
    expect(plainText("a [b](history) c")).toBe("a b c");
  });

  test("every FAQ link points to a real page and a known spot", () => {
    for (const messages of [en.messages, ar.messages] as Record<string, string>[]) {
      for (const section of FAQ) {
        for (const item of section.items) {
          const raw = messages[item.a] ?? "";
          const links = [...raw.matchAll(/\[[^\]]+\]\(([^)]*)\)/g)].map((m) => m[1] ?? "");
          const parsed = parseRichText(raw).filter((p) => p.route);
          expect(parsed.length, `${item.a}: ${links.join(", ")}`).toBe(links.length);
          for (const p of parsed) if (p.anchor) expect(ANCHORS.has(p.anchor), p.anchor).toBe(true);
        }
      }
    }
  });
});
