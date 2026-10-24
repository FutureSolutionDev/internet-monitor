import type { MessageKey } from "../lib/i18n";

export interface LegalSection {
  id: string;
  title: MessageKey;
  body: MessageKey[];
  list?: MessageKey[];
  quote?: MessageKey;
}

export interface LegalDoc {
  updated: string;
  intro: MessageKey;
  sections: LegalSection[];
}

export const PRIVACY: LegalDoc = {
  updated: "2026-10-24",
  intro: "privacy.intro",
  sections: [
    {
      id: "short",
      title: "privacy.short.title",
      body: ["privacy.short.p1"],
    },
    {
      id: "stored",
      title: "privacy.stored.title",
      body: ["privacy.stored.p1"],
      list: [
        "privacy.stored.li1",
        "privacy.stored.li2",
        "privacy.stored.li3",
        "privacy.stored.li4",
        "privacy.stored.li5",
      ],
    },
    {
      id: "network",
      title: "privacy.network.title",
      body: ["privacy.network.p1"],
      list: ["privacy.network.li1", "privacy.network.li2", "privacy.network.li3", "privacy.network.li4"],
    },
    {
      id: "dashboard",
      title: "privacy.dashboard.title",
      body: ["privacy.dashboard.p1"],
    },
    {
      id: "control",
      title: "privacy.control.title",
      body: ["privacy.control.p1"],
    },
    {
      id: "changes",
      title: "privacy.changes.title",
      body: ["privacy.changes.p1"],
    },
    {
      id: "contact",
      title: "privacy.contact.title",
      body: ["privacy.contact.p1"],
    },
  ],
};

export const TERMS: LegalDoc = {
  updated: "2026-10-24",
  intro: "terms.intro",
  sections: [
    {
      id: "use",
      title: "terms.use.title",
      body: ["terms.use.p1", "terms.use.p2", "terms.use.p3"],
    },
    {
      id: "warranty",
      title: "terms.warranty.title",
      body: ["terms.warranty.p1"],
      quote: "terms.warranty.quote",
    },
    {
      id: "measurements",
      title: "terms.measurements.title",
      body: ["terms.measurements.p1"],
    },
    {
      id: "fair",
      title: "terms.fair.title",
      body: ["terms.fair.p1"],
    },
    {
      id: "changes",
      title: "terms.changes.title",
      body: ["terms.changes.p1"],
    },
    {
      id: "contact",
      title: "terms.contact.title",
      body: ["terms.contact.p1"],
    },
  ],
};
