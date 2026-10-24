import type { Text } from "./text";

export type ChangeKind = "new" | "improved" | "fixed" | "security";

export interface ChangeItem {
  kind: ChangeKind;
  text: Text;
}

export interface Release {
  id: string;
  version: string | null;
  from?: string;
  date: string | null;
  title: Text;
  items: ChangeItem[];
}

export const KINDS: ChangeKind[] = ["new", "improved", "fixed", "security"];

export const RELEASES: Release[] = [
  {
    id: "v2",
    version: null,
    date: null,
    title: { en: "Version 2: rebuilt from the ground up", ar: "الإصدار 2: إعادة بناء كاملة" },
    items: [
      {
        kind: "new",
        text: {
          en: "A new dashboard with live status, a plain-language diagnosis, a latency chart, jitter, packet loss, and availability for today, 7 days and 30 days.",
          ar: "لوحة تحكم جديدة فيها الحالة المباشرة، وتشخيص بلغة بسيطة، ورسم لزمن الاستجابة، والتذبذب، وفقد الحزم، ونسبة التوفر لليوم و7 أيام و30 يوماً.",
        },
      },
      {
        kind: "new",
        text: {
          en: "A History screen with a log for each day, filters, and CSV export for one day, 7 days or 30 days.",
          ar: "شاشة السجل: سجل لكل يوم، مع فلاتر، وتصدير CSV ليوم واحد أو 7 أيام أو 30 يوماً.",
        },
      },
      {
        kind: "new",
        text: {
          en: "A monthly report page with “Print or save as PDF”, ready to send to your internet provider.",
          ar: "صفحة تقرير شهري فيها «اطبع أو احفظ PDF»، جاهزة للإرسال لمزوّد الخدمة.",
        },
      },
      {
        kind: "new",
        text: {
          en: "A Speed test screen with a live gauge, today's runs, scheduled tests and an alert threshold.",
          ar: "شاشة قياس السرعة: مؤشر مباشر، ونتائج اليوم، واختبارات مجدولة، وحد للتنبيه.",
        },
      },
      {
        kind: "new",
        text: {
          en: "Dark and light themes with a circular reveal when you switch, plus full Arabic and English with a right-to-left layout.",
          ar: "ثيم داكن وفاتح مع انتقال دائري عند التبديل، ودعم كامل للعربية والإنجليزية مع اتجاه من اليمين لليسار.",
        },
      },
      {
        kind: "new",
        text: {
          en: "In-app pages: What's new, FAQ, Privacy and Terms of use.",
          ar: "صفحات داخل التطبيق: ما الجديد، والأسئلة الشائعة، والخصوصية، وشروط الاستخدام.",
        },
      },
      {
        kind: "new",
        text: { en: "A new, gentler default alert sound.", ar: "صوت تنبيه افتراضي جديد وأهدأ." },
      },
      {
        kind: "improved",
        text: {
          en: "Settings shows whether your custom sound is active and how long it is, with “Try the sound” and “Play like a real alert”.",
          ar: "الإعدادات توضّح إن كان صوتك المخصص هو المستخدم ومدته، مع زرّي «جرّب الصوت» و«شغّله كتنبيه حقيقي».",
        },
      },
      {
        kind: "improved",
        text: {
          en: "Settings save safely even when two windows change them at once, and warn you before leaving with unsaved changes.",
          ar: "حفظ الإعدادات آمن حتى لو عدّلتها نافذتان في نفس الوقت، وتنبيه قبل المغادرة إذا وُجدت تعديلات غير محفوظة.",
        },
      },
      {
        kind: "improved",
        text: {
          en: "The layout works in phone-size windows.",
          ar: "التصميم يعمل في النوافذ الصغيرة بحجم الهاتف.",
        },
      },
      {
        kind: "fixed",
        text: {
          en: "Uploaded custom alert sounds silently fell back to the default. Uploads are now checked and converted, and the sound you upload is the one that plays.",
          ar: "الأصوات المخصصة المرفوعة كانت ترجع للصوت الافتراضي بدون أي رسالة. الآن يُفحص الملف ويُحوَّل، والصوت الذي ترفعه هو الذي يعمل.",
        },
      },
      {
        kind: "fixed",
        text: {
          en: "The auto-updater no longer picks the window (GUI) build for a tray install.",
          ar: "التحديث التلقائي لم يعد يختار نسخة النافذة (GUI) لتثبيت يعمل من شريط المهام.",
        },
      },
      {
        kind: "security",
        text: {
          en: "The dashboard is reachable only from this computer and needs a key that changes every launch. Open it from the tray menu or the app window.",
          ar: "لوحة التحكم متاحة من هذا الجهاز فقط وتحتاج مفتاحاً يتغيّر مع كل تشغيل. افتحها من قائمة شريط المهام أو نافذة التطبيق.",
        },
      },
      {
        kind: "security",
        text: {
          en: "Your webhook URL and Telegram token are never sent back to the browser. Settings shows them as “Saved — hidden”.",
          ar: "رابط الـ webhook وتوكن Telegram لا يُرسلان للمتصفح أبداً، وتظهر في الإعدادات كـ«محفوظ — مخفي».",
        },
      },
      {
        kind: "security",
        text: {
          en: "Webhook URLs must use https and point to Discord or Slack.",
          ar: "روابط الـ webhook يجب أن تكون https وتشير إلى Discord أو Slack.",
        },
      },
      {
        kind: "security",
        text: {
          en: "Updates download only from GitHub, and the Go networking and crypto libraries were updated.",
          ar: "التحديثات تُنزَّل من GitHub فقط، وتم تحديث مكتبات الشبكة والتشفير في Go.",
        },
      },
    ],
  },
  {
    id: "v0.10.1",
    version: "0.10.1",
    date: "2026-06-01",
    title: { en: "Documentation refresh", ar: "تحديث التوثيق" },
    items: [
      {
        kind: "improved",
        text: {
          en: "The README and screenshots now cover the latest features.",
          ar: "ملف README والصور أصبحت تغطي أحدث الميزات.",
        },
      },
    ],
  },
  {
    id: "v0.10.0",
    version: "0.10.0",
    date: "2026-06-01",
    title: { en: "Check for updates on demand", ar: "البحث عن تحديثات عند الطلب" },
    items: [
      {
        kind: "new",
        text: {
          en: "A “Check for updates” button in the dashboard, so you don't have to wait for the background check.",
          ar: "زر «ابحث عن تحديثات» في لوحة التحكم، بدل انتظار الفحص التلقائي في الخلفية.",
        },
      },
      {
        kind: "improved",
        text: {
          en: "A new version is published only when every platform builds successfully.",
          ar: "الإصدار الجديد يُنشر فقط عندما تنجح النسخ لكل الأنظمة.",
        },
      },
      {
        kind: "fixed",
        text: {
          en: "Release builds that stopped before starting, and build checks that failed on Linux.",
          ar: "عمليات بناء الإصدارات التي كانت تتوقف قبل أن تبدأ، وفحوصات البناء التي كانت تفشل على Linux.",
        },
      },
    ],
  },
  {
    id: "v0.9.4",
    version: "0.9.4",
    from: "0.9.0",
    date: "2026-06-01",
    title: { en: "Outage reports you can send to your provider", ar: "تقارير انقطاع ترسلها لمزوّد الخدمة" },
    items: [
      {
        kind: "new",
        text: {
          en: "A monthly outage report (PDF) with a month picker, charts, a log of events in order, and the cause of each outage. It's neutral by default, with an optional branding footer.",
          ar: "تقرير انقطاع شهري (PDF) فيه اختيار الشهر، ورسوم بيانية، وسجل الأحداث بالترتيب، وسبب كل انقطاع. محايد افتراضياً مع تذييل اختياري باسم التطبيق.",
        },
      },
      {
        kind: "new",
        text: {
          en: "Upload speed test with live progress, next to download and ping, on a circular gauge.",
          ar: "قياس سرعة الرفع مع تقدّم مباشر، بجانب التنزيل والـ ping، على مؤشر دائري.",
        },
      },
      {
        kind: "new",
        text: { en: "Scheduled automatic speed tests.", ar: "قياسات سرعة تلقائية مجدولة." },
      },
      {
        kind: "new",
        text: {
          en: "Telegram notifications, and system notifications in your chosen language.",
          ar: "إشعارات Telegram، وإشعارات النظام بلغتك المختارة.",
        },
      },
      {
        kind: "new",
        text: {
          en: "Tells DNS, website and full outages apart, says whether the problem is your local network or your provider, and adds optional ICMP ping, jitter and a Prometheus /metrics endpoint.",
          ar: "يفرّق بين انقطاع DNS والمواقع والانقطاع الكامل، ويحدد إن كانت المشكلة في شبكتك المحلية أو عند المزوّد، مع ping ICMP اختياري، والتذبذب، ونقطة Prometheus على /metrics.",
        },
      },
      {
        kind: "new",
        text: {
          en: "Availability widget, copy summary, and CSV export for a date range.",
          ar: "بطاقة نسبة التوفر، ونسخ الملخص، وتصدير CSV لفترة زمنية.",
        },
      },
      {
        kind: "improved",
        text: {
          en: "Settings changes apply right away, without a restart.",
          ar: "تعديلات الإعدادات تُطبَّق فوراً بدون إعادة تشغيل.",
        },
      },
      {
        kind: "improved",
        text: {
          en: "One sound player for every alert. The previous sound always stops first, so chimes never overlap, and uploaded MP3s are converted so system notifications can play them.",
          ar: "مشغّل صوت واحد لكل التنبيهات. الصوت السابق يتوقف أولاً فلا تتداخل الأصوات، وملفات MP3 المرفوعة تُحوَّل لتعمل مع إشعارات النظام.",
        },
      },
      {
        kind: "fixed",
        text: {
          en: "Outages that cross into a new month are counted correctly in both months.",
          ar: "الانقطاعات الممتدة لشهر جديد تُحسب بشكل صحيح في الشهرين.",
        },
      },
      {
        kind: "fixed",
        text: {
          en: "More accurate latency, captive-portal (login page) detection, uptime, and fewer false alarms at startup. Speed tests no longer race.",
          ar: "زمن استجابة أدق، واكتشاف صفحات تسجيل الدخول للشبكات العامة، وحساب أدق لمدة التشغيل، وإنذارات خاطئة أقل عند البدء. وقياس السرعة لم يعد يتعارض مع نفسه.",
        },
      },
      {
        kind: "security",
        text: {
          en: "Downloaded updates are checked against the release's SHA256SUMS before they're installed.",
          ar: "التحديثات المنزّلة تُطابَق مع ملف SHA256SUMS الخاص بالإصدار قبل تثبيتها.",
        },
      },
      {
        kind: "security",
        text: {
          en: "The local dashboard API is protected against cross-site requests and oversized requests.",
          ar: "واجهة لوحة التحكم المحلية محمية من الطلبات القادمة من مواقع أخرى ومن الطلبات الضخمة.",
        },
      },
    ],
  },
  {
    id: "v0.8.1",
    version: "0.8.1",
    from: "0.8.0",
    date: "2026-05-12",
    title: { en: "Native Windows notifications", ar: "إشعارات Windows أصلية" },
    items: [
      {
        kind: "improved",
        text: {
          en: "Notifications use the native Windows balloon instead of PowerShell, and the app makes sure Windows lets it show them.",
          ar: "الإشعارات تستخدم فقاعة Windows الأصلية بدل PowerShell، والتطبيق يتأكد أن Windows يسمح بعرضها.",
        },
      },
      {
        kind: "improved",
        text: {
          en: "The window and tray icons change with the connection status.",
          ar: "أيقونة النافذة وشريط المهام تتغيّر حسب حالة الاتصال.",
        },
      },
    ],
  },
  {
    id: "v0.7.0",
    version: "0.7.0",
    date: "2026-05-12",
    title: { en: "Your own alert sound", ar: "صوت تنبيه من اختيارك" },
    items: [
      {
        kind: "new",
        text: { en: "Upload a custom notification sound.", ar: "رفع صوت إشعار مخصص." },
      },
    ],
  },
  {
    id: "v0.6.0",
    version: "0.6.0",
    date: "2026-05-12",
    title: { en: "A new logo", ar: "شعار جديد" },
    items: [
      {
        kind: "new",
        text: {
          en: "A new app logo: a globe with a pulse line.",
          ar: "شعار جديد للتطبيق: كرة أرضية مع خط نبض.",
        },
      },
      {
        kind: "fixed",
        text: {
          en: "Notifications that fired on top of each other, sounds that replayed without stopping, and empty notification text.",
          ar: "إشعارات كانت تظهر فوق بعضها، وأصوات تُعاد بدون أن تتوقف، ونص إشعار فارغ.",
        },
      },
    ],
  },
  {
    id: "v0.5.0",
    version: "0.5.0",
    date: "2026-05-12",
    title: { en: "Speed test", ar: "قياس السرعة" },
    items: [
      {
        kind: "new",
        text: {
          en: "A Speed test tab with your own download targets and timeout.",
          ar: "تبويب قياس السرعة مع أهداف تنزيل ومهلة من اختيارك.",
        },
      },
      {
        kind: "new",
        text: { en: "A table of the last 20 checks.", ar: "جدول بآخر 20 فحصاً." },
      },
      {
        kind: "improved",
        text: {
          en: "Settings split into four tabs: monitoring, targets, notifications and speed test.",
          ar: "الإعدادات مقسّمة لأربعة تبويبات: المراقبة، والأهداف، والإشعارات، وقياس السرعة.",
        },
      },
      {
        kind: "improved",
        text: {
          en: "A rewritten monitoring engine, and a full README in English and Arabic.",
          ar: "محرك مراقبة أُعيدت كتابته، وملف README كامل بالعربية والإنجليزية.",
        },
      },
      {
        kind: "fixed",
        text: {
          en: "The speed test progress bar follows the timeout you set.",
          ar: "شريط تقدّم قياس السرعة يتبع المهلة التي حددتها.",
        },
      },
    ],
  },
  {
    id: "v0.4.3",
    version: "0.4.3",
    from: "0.4.0",
    date: "2026-05-12",
    title: { en: "Live update notices", ar: "إشعار التحديثات مباشرة" },
    items: [
      {
        kind: "new",
        text: {
          en: "The dashboard shows when a new version is available, live.",
          ar: "لوحة التحكم توضّح مباشرة عند توفر إصدار جديد.",
        },
      },
      {
        kind: "fixed",
        text: { en: "The close button on the update banner.", ar: "زر الإغلاق في شريط التحديث." },
      },
      {
        kind: "security",
        text: { en: "Updated the Go crypto library.", ar: "تحديث مكتبة التشفير في Go." },
      },
    ],
  },
  {
    id: "v0.3.2",
    version: "0.3.2",
    from: "0.3.0",
    date: "2026-05-11",
    title: { en: "Start with Windows", ar: "التشغيل مع Windows" },
    items: [
      {
        kind: "new",
        text: { en: "An option to start the app with Windows.", ar: "خيار لتشغيل التطبيق مع بدء Windows." },
      },
      {
        kind: "improved",
        text: {
          en: "Versions are numbered automatically, so the updater always finds the newest one.",
          ar: "ترقيم الإصدارات أصبح تلقائياً، فيجد التحديث التلقائي أحدث إصدار دائماً.",
        },
      },
      {
        kind: "fixed",
        text: {
          en: "The app could freeze when you quit it from the tray on Windows.",
          ar: "التطبيق كان قد يتجمد عند إغلاقه من شريط المهام على Windows.",
        },
      },
    ],
  },
  {
    id: "v0.2.0",
    version: "0.2.0",
    date: "2026-05-11",
    title: { en: "Automatic updates", ar: "التحديث التلقائي" },
    items: [
      {
        kind: "new",
        text: { en: "Automatic updates.", ar: "تحديث تلقائي للتطبيق." },
      },
      {
        kind: "new",
        text: {
          en: "Several website and DNS targets instead of one.",
          ar: "أكثر من هدف للمواقع والـ DNS بدل هدف واحد.",
        },
      },
      {
        kind: "new",
        text: {
          en: "A button to test your webhook, and only one copy of the app runs at a time.",
          ar: "زر لتجربة الـ webhook، ونسخة واحدة فقط من التطبيق تعمل في نفس الوقت.",
        },
      },
      {
        kind: "improved",
        text: {
          en: "Errors and outage reasons are shown in Arabic or English.",
          ar: "الأخطاء وأسباب الانقطاع تظهر بالعربية أو الإنجليزية.",
        },
      },
    ],
  },
  {
    id: "v0.1.0",
    version: "0.1.0",
    date: "2026-05-11",
    title: { en: "Tray icon and alerts", ar: "أيقونة شريط المهام والتنبيهات" },
    items: [
      {
        kind: "new",
        text: {
          en: "A tray icon that changes color with the connection status.",
          ar: "أيقونة في شريط المهام يتغيّر لونها حسب حالة الاتصال.",
        },
      },
      {
        kind: "new",
        text: { en: "Sound alerts and desktop notifications.", ar: "تنبيهات صوتية وإشعارات على سطح المكتب." },
      },
    ],
  },
  {
    id: "v0.0.1",
    version: "0.0.1",
    date: "2026-05-11",
    title: { en: "First release", ar: "الإصدار الأول" },
    items: [
      {
        kind: "new",
        text: {
          en: "A web dashboard with live status, settings and a log viewer.",
          ar: "لوحة تحكم ويب فيها الحالة المباشرة، والإعدادات، وعرض السجلات.",
        },
      },
      {
        kind: "new",
        text: {
          en: "Discord and Slack webhook alerts, and a default config created on first run.",
          ar: "تنبيهات webhook على Discord وSlack، وملف إعدادات افتراضي يُنشأ عند أول تشغيل.",
        },
      },
      {
        kind: "new",
        text: {
          en: "Two builds: one that lives in the tray and one with its own window.",
          ar: "نسختان: واحدة تعمل من شريط المهام، وأخرى بنافذة خاصة.",
        },
      },
    ],
  },
];

function parts(v: string): number[] {
  const core = v.replace(/^v/, "").split("-")[0] ?? "";
  return core.split(".").map((n) => Number.parseInt(n, 10) || 0);
}

export function compareVersions(a: string, b: string): number {
  const pa = parts(a);
  const pb = parts(b);
  for (let i = 0; i < 3; i++) {
    const d = (pa[i] ?? 0) - (pb[i] ?? 0);
    if (d !== 0) return d;
  }
  return 0;
}

export function isReleasedVersion(v: string | undefined | null): v is string {
  return !!v && /^v?\d+\.\d+\.\d+/.test(v);
}

export function releaseOfVersion(running: string | undefined | null): string | null {
  if (!isReleasedVersion(running)) return null;
  for (const r of RELEASES) {
    if (r.version === null) continue;
    const low = r.from ?? r.version;
    if (compareVersions(running, low) >= 0 && compareVersions(running, r.version) <= 0) return r.id;
  }
  const newest = RELEASES.find((r) => r.version !== null)?.version;
  const next = RELEASES[0];
  if (newest && next?.version === null && compareVersions(running, newest) > 0) return next.id;
  return null;
}
