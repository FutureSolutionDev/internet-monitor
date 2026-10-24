import { HardDrive, Languages, type LucideIcon, Scale, ShieldCheck } from "lucide-react";
import type { MessageKey } from "../lib/i18n";

export interface Principle {
  id: string;
  icon: LucideIcon;
  title: MessageKey;
  body: MessageKey;
}

export const PUBLISHER = "FutureSolutionDev";
export const DEVELOPER = "Sabry Dawood";
export const PUBLISHER_URL = "https://github.com/FutureSolutionDev";

export const ABOUT_STORY: MessageKey[] = ["about.story.1", "about.story.2"];

export const PRINCIPLES: Principle[] = [
  {
    id: "local",
    icon: HardDrive,
    title: "about.principle.local.title",
    body: "about.principle.local.body",
  },
  {
    id: "open",
    icon: Scale,
    title: "about.principle.open.title",
    body: "about.principle.open.body",
  },
  {
    id: "proof",
    icon: ShieldCheck,
    title: "about.principle.proof.title",
    body: "about.principle.proof.body",
  },
  {
    id: "bilingual",
    icon: Languages,
    title: "about.principle.bilingual.title",
    body: "about.principle.bilingual.body",
  },
];
