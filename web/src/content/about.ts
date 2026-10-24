import { HardDrive, Languages, type LucideIcon, Scale, ShieldCheck } from "lucide-react";
import type { Text } from "./text";

export interface Principle {
  id: string;
  icon: LucideIcon;
  title: Text;
  body: Text;
}

export const PUBLISHER = "FutureSolutionDev";
export const DEVELOPER = "Sabry Dawood";
export const PUBLISHER_URL = "https://github.com/FutureSolutionDev";

export const ABOUT_STORY: Text[] = [
  {
    en: "Internet Monitor started with a familiar problem: the internet keeps dropping, and your provider says the line is fine. Without a record, it's your word against theirs.",
    ar: "بدأ مراقب الإنترنت من مشكلة مألوفة: الإنترنت ينقطع بين حين وآخر، والمزوّد يقول إن الخط سليم. ومن دون سجل، يصبح كلامك في مواجهة كلامهم.",
  },
  {
    en: "So the app watches your connection quietly in the background, records every outage with its exact time, length and cause, and turns that into a report you can send. It also tells you whether the problem is in your home network or at your provider, so you know who to call.",
    ar: "لذلك يراقب التطبيق اتصالك بهدوء في الخلفية، ويسجّل كل انقطاع بوقته ومدته وسببه بالضبط، ويحوّل ذلك إلى تقرير يمكنك إرساله. ويخبرك أيضاً إن كانت المشكلة في شبكة المنزل أم عند المزوّد، فتعرف بمن تتصل.",
  },
];

export const PRINCIPLES: Principle[] = [
  {
    id: "local",
    icon: HardDrive,
    title: { en: "Your data stays with you", ar: "بياناتك تبقى معك" },
    body: {
      en: "Everything is stored on your computer. No account, no cloud, no tracking.",
      ar: "كل شيء محفوظ على جهازك. لا حساب، ولا سحابة، ولا تتبّع.",
    },
  },
  {
    id: "open",
    icon: Scale,
    title: { en: "Free and open source", ar: "مجاني ومفتوح المصدر" },
    body: {
      en: "MIT licensed, free for personal and commercial use. Anyone can read the code.",
      ar: "برخصة MIT، مجاني للاستخدام الشخصي والتجاري، والكود متاح لأي شخص يقرأه.",
    },
  },
  {
    id: "proof",
    icon: ShieldCheck,
    title: { en: "Evidence, not guesses", ar: "دليل، لا تخمين" },
    body: {
      en: "Exact times, durations and the layer that failed, ready for a PDF or CSV.",
      ar: "أوقات ومدد دقيقة والجزء الذي تعطّل، جاهزة كـ PDF أو CSV.",
    },
  },
  {
    id: "bilingual",
    icon: Languages,
    title: { en: "Arabic and English", ar: "عربي وإنجليزي" },
    body: {
      en: "Every screen, alert and report works in both languages, right to left included.",
      ar: "كل شاشة وتنبيه وتقرير يعمل باللغتين، مع الاتجاه من اليمين لليسار.",
    },
  },
];
