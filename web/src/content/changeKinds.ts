import { type LucideIcon, ShieldCheck, Sparkles, TrendingUp, Wrench } from "lucide-react";
import type { MessageKey } from "../lib/i18n";
import type { ChangeKind } from "./changelog";

export const CHANGE_KIND: Record<ChangeKind, { label: MessageKey; icon: LucideIcon; tone: string }> = {
  new: { label: "kind_new", icon: Sparkles, tone: "var(--accent-blue)" },
  improved: { label: "kind_improved", icon: TrendingUp, tone: "var(--accent-violet)" },
  fixed: { label: "kind_fixed", icon: Wrench, tone: "var(--accent-green)" },
  security: { label: "kind_security", icon: ShieldCheck, tone: "var(--accent-amber)" },
};
