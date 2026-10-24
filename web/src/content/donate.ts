import { CreditCard, Landmark, type LucideIcon, Smartphone } from "lucide-react";
import type { MessageKey } from "../lib/i18n";

export interface DonateConfig {
  paypal: string;
  vodafoneCash: string;
  instapay: string;
  instapayLink: string;
}

export const DONATE: DonateConfig = {
  paypal: "",
  vodafoneCash: "",
  instapay: "",
  instapayLink: "",
};

export type DonateAccent = "blue" | "coral" | "violet";

export interface DonateMethod {
  id: "paypal" | "vodafone-cash" | "instapay";
  icon: LucideIcon;
  accent: DonateAccent;
  name: MessageKey;
  how: MessageKey;
  handle: string;
  href?: string;
}

const PAYPAL_HOSTS = new Set(["paypal.me", "www.paypal.me", "paypal.com", "www.paypal.com"]);
const INSTAPAY_HOSTS = new Set(["ipn.eg"]);

function httpsOn(raw: string, hosts: Set<string>): boolean {
  try {
    const u = new URL(raw);
    return u.protocol === "https:" && !u.username && !u.port && hosts.has(u.hostname.toLowerCase());
  } catch {
    return false;
  }
}

export const isPaypalLink = (raw: string) => httpsOn(raw, PAYPAL_HOSTS);
export const isInstapayLink = (raw: string) => httpsOn(raw, INSTAPAY_HOSTS);
export const isVodafoneCash = (raw: string) => /^010\d{8}$/.test(raw);
export const isInstapayAddress = (raw: string) => /^[a-z0-9._-]+@instapay$/i.test(raw);

export function donateMethods(cfg: DonateConfig = DONATE): DonateMethod[] {
  const out: DonateMethod[] = [];
  const paypal = cfg.paypal.trim();
  if (isPaypalLink(paypal)) {
    out.push({
      id: "paypal",
      icon: CreditCard,
      accent: "blue",
      name: "support.paypal.name",
      how: "support.paypal.how",
      handle: new URL(paypal).host + new URL(paypal).pathname.replace(/\/$/, ""),
      href: paypal,
    });
  }
  const phone = cfg.vodafoneCash.replace(/[\s-]/g, "");
  if (isVodafoneCash(phone)) {
    out.push({
      id: "vodafone-cash",
      icon: Smartphone,
      accent: "coral",
      name: "support.vodafone.name",
      how: "support.vodafone.how",
      handle: phone,
    });
  }
  const address = cfg.instapay.trim();
  const link = cfg.instapayLink.trim();
  if (isInstapayAddress(address)) {
    out.push({
      id: "instapay",
      icon: Landmark,
      accent: "violet",
      name: "support.instapay.name",
      how: "support.instapay.how",
      handle: address,
      href: isInstapayLink(link) ? link : undefined,
    });
  }
  return out;
}
