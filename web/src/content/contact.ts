import { GitBranch, type LucideIcon, Mail, MessageCircle, MessagesSquare } from "lucide-react";
import type { Text } from "./text";
import { REPO_URL } from "./text";

export type ContactAccent = "green" | "violet" | "blue" | "coral";

export interface ContactChannel {
  id: string;
  icon: LucideIcon;
  accent: ContactAccent;
  name: Text;
  handle: string;
  href: string;
  best: Text;
}

export const CONTACT_EMAIL = "sabrydawood@outlook.com";
export const WHATSAPP_NUMBER = "+20 106 901 7890";

export const CHANNELS: ContactChannel[] = [
  {
    id: "whatsapp",
    icon: MessageCircle,
    accent: "green",
    name: { en: "WhatsApp", ar: "واتساب" },
    handle: WHATSAPP_NUMBER,
    href: "https://wa.me/201069017890",
    best: {
      en: "Quick questions and help getting set up.",
      ar: "الأسئلة السريعة والمساعدة في الإعداد.",
    },
  },
  {
    id: "discord",
    icon: MessagesSquare,
    accent: "violet",
    name: { en: "Discord", ar: "ديسكورد" },
    handle: "discord.gg/7Vhx6fw4qj",
    href: "https://discord.gg/7Vhx6fw4qj",
    best: {
      en: "Chat with other users, share ideas and follow new releases.",
      ar: "الدردشة مع المستخدمين، ومشاركة الأفكار، ومتابعة الإصدارات الجديدة.",
    },
  },
  {
    id: "email",
    icon: Mail,
    accent: "blue",
    name: { en: "Email", ar: "البريد الإلكتروني" },
    handle: CONTACT_EMAIL,
    href: `mailto:${CONTACT_EMAIL}`,
    best: {
      en: "Business, privacy questions and anything you'd rather not post publicly.",
      ar: "التعاون، وأسئلة الخصوصية، وأي شيء لا تريد نشره علناً.",
    },
  },
  {
    id: "github",
    icon: GitBranch,
    accent: "coral",
    name: { en: "GitHub", ar: "GitHub" },
    handle: "FutureSolutionDev/internet-monitor",
    href: `${REPO_URL}/issues`,
    best: {
      en: "Bug reports and feature requests, in public, so others can follow them.",
      ar: "الإبلاغ عن الأخطاء وطلب الميزات بشكل علني، ليتابعها الآخرون.",
    },
  },
];
