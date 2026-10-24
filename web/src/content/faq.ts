import { BellRing, Compass, FileText, type LucideIcon, Radar, ShieldCheck } from "lucide-react";
import type { MessageKey } from "../lib/i18n";

export type FaqAccent = "blue" | "violet" | "amber" | "green" | "coral";

export interface FaqItem {
  id: string;
  q: MessageKey;
  a: MessageKey;
}

export interface FaqSection {
  id: string;
  accent: FaqAccent;
  icon: LucideIcon;
  title: MessageKey;
  items: FaqItem[];
}

export const FAQ: FaqSection[] = [
  {
    id: "start",
    accent: "blue",
    icon: Compass,
    title: "faq.section.start",
    items: [
      { id: "what", q: "faq.what.q", a: "faq.what.a" },
      { id: "open", q: "faq.open.q", a: "faq.open.a" },
      { id: "platforms", q: "faq.platforms.q", a: "faq.platforms.a" },
      { id: "free", q: "faq.free.q", a: "faq.free.a" },
    ],
  },
  {
    id: "monitoring",
    accent: "violet",
    icon: Radar,
    title: "faq.section.monitoring",
    items: [
      { id: "disconnected", q: "faq.disconnected.q", a: "faq.disconnected.a" },
      { id: "unstable", q: "faq.unstable.q", a: "faq.unstable.a" },
      { id: "layers", q: "faq.layers.q", a: "faq.layers.a" },
      { id: "router", q: "faq.router.q", a: "faq.router.a" },
      { id: "targets", q: "faq.targets.q", a: "faq.targets.a" },
      { id: "data", q: "faq.data.q", a: "faq.data.a" },
    ],
  },
  {
    id: "alerts",
    accent: "amber",
    icon: BellRing,
    title: "faq.section.alerts",
    items: [
      { id: "where", q: "faq.where.q", a: "faq.where.a" },
      { id: "telegram", q: "faq.telegram.q", a: "faq.telegram.a" },
      { id: "sound", q: "faq.sound.q", a: "faq.sound.a" },
      { id: "silent", q: "faq.silent.q", a: "faq.silent.a" },
    ],
  },
  {
    id: "reports",
    accent: "green",
    icon: FileText,
    title: "faq.section.reports",
    items: [
      { id: "proof", q: "faq.proof.q", a: "faq.proof.a" },
      { id: "retention", q: "faq.retention.q", a: "faq.retention.a" },
      { id: "speedserver", q: "faq.speedserver.q", a: "faq.speedserver.a" },
      { id: "schedule", q: "faq.schedule.q", a: "faq.schedule.a" },
    ],
  },
  {
    id: "privacy",
    accent: "coral",
    icon: ShieldCheck,
    title: "faq.section.privacy",
    items: [
      { id: "send", q: "faq.send.q", a: "faq.send.a" },
      { id: "lan", q: "faq.lan.q", a: "faq.lan.a" },
      { id: "secrets", q: "faq.secrets.q", a: "faq.secrets.a" },
      { id: "updates", q: "faq.updates.q", a: "faq.updates.a" },
      { id: "bug", q: "faq.bug.q", a: "faq.bug.a" },
    ],
  },
];
