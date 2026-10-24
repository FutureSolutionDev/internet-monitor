import { GitBranch, type LucideIcon, Mail, MessageCircle, MessagesSquare } from "lucide-react";
import type { MessageKey } from "../lib/i18n";
import { REPO_URL } from "./text";

export type ContactAccent = "green" | "violet" | "blue" | "coral";

export interface ContactChannel {
  id: string;
  icon: LucideIcon;
  accent: ContactAccent;
  name: MessageKey;
  handle: string;
  href: string;
  best: MessageKey;
}

export const CONTACT_EMAIL = "sabrydawood@outlook.com";
export const WHATSAPP_NUMBER = "+20 106 901 7890";

export const CHANNELS: ContactChannel[] = [
  {
    id: "whatsapp",
    icon: MessageCircle,
    accent: "green",
    name: "contact.whatsapp.name",
    handle: WHATSAPP_NUMBER,
    href: "https://wa.me/201069017890",
    best: "contact.whatsapp.best",
  },
  {
    id: "discord",
    icon: MessagesSquare,
    accent: "violet",
    name: "contact.discord.name",
    handle: "discord.gg/7Vhx6fw4qj",
    href: "https://discord.gg/7Vhx6fw4qj",
    best: "contact.discord.best",
  },
  {
    id: "email",
    icon: Mail,
    accent: "blue",
    name: "contact.email.name",
    handle: CONTACT_EMAIL,
    href: `mailto:${CONTACT_EMAIL}`,
    best: "contact.email.best",
  },
  {
    id: "github",
    icon: GitBranch,
    accent: "coral",
    name: "contact.github.name",
    handle: "FutureSolutionDev/internet-monitor",
    href: `${REPO_URL}/issues`,
    best: "contact.github.best",
  },
];
