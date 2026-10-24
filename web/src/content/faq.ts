import { BellRing, Compass, FileText, type LucideIcon, Radar, ShieldCheck } from "lucide-react";
import type { Text } from "./text";

export type FaqAccent = "blue" | "violet" | "amber" | "green" | "coral";

export interface FaqItem {
  id: string;
  q: Text;
  a: Text;
}

export interface FaqSection {
  id: string;
  accent: FaqAccent;
  icon: LucideIcon;
  title: Text;
  items: FaqItem[];
}

export const FAQ: FaqSection[] = [
  {
    id: "start",
    accent: "blue",
    icon: Compass,
    title: { en: "Getting started", ar: "البداية" },
    items: [
      {
        id: "what",
        q: { en: "What does Internet Monitor do?", ar: "ماذا يفعل مراقب الإنترنت؟" },
        a: {
          en: "It checks your connection every few seconds in three ways: it connects to DNS servers, loads a small web page, and looks up website names. Every outage is saved with its exact start time, how long it lasted, and which part failed. You can then turn that into a report for your internet provider.",
          ar: "يفحص اتصالك كل بضع ثوانٍ بثلاث طرق: يتصل بخوادم DNS، ويفتح صفحة ويب صغيرة، ويبحث عن أسماء المواقع. كل انقطاع يُحفظ بوقت بدايته بالضبط، ومدته، والجزء الذي تعطّل. ثم يمكنك تحويل ذلك لتقرير ترسله لمزوّد الخدمة.",
        },
      },
      {
        id: "open",
        q: { en: "How do I open the dashboard?", ar: "كيف أفتح لوحة التحكم؟" },
        a: {
          en: "Click the tray icon and choose “Open Dashboard”, or use the app window if you installed the window version. Typing the address into the browser yourself shows a locked page, because the dashboard needs a key that changes every time the app starts.",
          ar: "اضغط على أيقونة التطبيق في شريط المهام واختر «Open Dashboard»، أو استخدم نافذة التطبيق إن كنت قد ثبّتّ نسخة النافذة. كتابة العنوان في المتصفح بنفسك تفتح صفحة مقفلة، لأن لوحة التحكم تحتاج مفتاحاً يتغيّر مع كل تشغيل للتطبيق.",
        },
      },
      {
        id: "platforms",
        q: { en: "Does it work on macOS and Linux?", ar: "هل يعمل على macOS وLinux؟" },
        a: {
          en: "Yes. There are builds for Windows, macOS and Linux. The “Start with Windows” option only exists on Windows.",
          ar: "نعم. توجد نسخ لـ Windows وmacOS وLinux. خيار «التشغيل مع Windows» موجود على Windows فقط.",
        },
      },
      {
        id: "free",
        q: { en: "Is it free?", ar: "هل هو مجاني؟" },
        a: {
          en: "Yes. It's open source under the MIT license, free for personal and commercial use. There's no account and nothing to sign up for.",
          ar: "نعم. مفتوح المصدر برخصة MIT، ومجاني للاستخدام الشخصي والتجاري. لا يوجد حساب ولا تسجيل.",
        },
      },
    ],
  },
  {
    id: "monitoring",
    accent: "violet",
    icon: Radar,
    title: { en: "Monitoring and outages", ar: "المراقبة والانقطاعات" },
    items: [
      {
        id: "disconnected",
        q: { en: "When does it say I'm disconnected?", ar: "متى يقول إن الاتصال مقطوع؟" },
        a: {
          en: "When checks fail 3 times in a row. With the default 5-second interval, that's about 15 seconds. This keeps a single lost check from counting as an outage. You can change both numbers in Settings → Monitoring.",
          ar: "عندما يفشل الفحص 3 مرات متتالية. مع الفاصل الافتراضي 5 ثوانٍ، هذا حوالي 15 ثانية. بهذا لا يُحسب فحص واحد ضائع كانقطاع. يمكنك تغيير الرقمين من الإعدادات ← المراقبة.",
        },
      },
      {
        id: "unstable",
        q: { en: "What does “Unstable” mean?", ar: "ماذا تعني «غير مستقر»؟" },
        a: {
          en: "You're online, but the connection is poor: more than 20% of packets are lost, or responses take longer than 500 ms. Both limits can be changed in Settings → Monitoring.",
          ar: "أنت متصل، لكن الاتصال ضعيف: أكثر من 20% من الحزم تضيع، أو الاستجابة تأخذ أكثر من 500 مللي ثانية. الحدّان قابلان للتغيير من الإعدادات ← المراقبة.",
        },
      },
      {
        id: "layers",
        q: {
          en: "What's the difference between a DNS problem and a full outage?",
          ar: "ما الفرق بين مشكلة DNS والانقطاع الكامل؟",
        },
        a: {
          en: "A full outage means nothing gets through. A DNS problem means the line works but website names can't be looked up, so pages don't open. A website problem means names resolve but pages don't load. The History screen and the report show which one happened, which is useful when your provider says “the line was fine”.",
          ar: "الانقطاع الكامل يعني أن لا شيء يمر. مشكلة DNS تعني أن الخط شغّال لكن أسماء المواقع لا تُترجم، فالصفحات لا تفتح. مشكلة المواقع تعني أن الأسماء تُترجم لكن الصفحات لا تُحمَّل. شاشة السجل والتقرير يوضّحان أيّها حدث، وهذا مفيد عندما يقول المزوّد «الخط كان سليماً».",
        },
      },
      {
        id: "router",
        q: { en: "Is the problem my router or my provider?", ar: "هل المشكلة في الراوتر أم عند المزوّد؟" },
        a: {
          en: "The app also checks your router (the local gateway). If the router answers but the internet doesn't, the diagnosis points to your provider. If the router doesn't answer either, the problem is in your home network: Wi-Fi, cables, or the router itself.",
          ar: "التطبيق يفحص الراوتر أيضاً (البوابة المحلية). إذا كان الراوتر يستجيب والإنترنت لا، فالتشخيص يشير للمزوّد. وإذا لم يستجب الراوتر نفسه، فالمشكلة في شبكة المنزل: الـ Wi-Fi أو الكابلات أو الراوتر.",
        },
      },
      {
        id: "targets",
        q: { en: "Can I change what it checks?", ar: "هل أقدر أغيّر ما يفحصه؟" },
        a: {
          en: "Yes. Settings → Check targets has three lists: DNS servers (8.8.8.8 and 1.1.1.1 by default), web pages, and names to look up. Each list holds up to 10 entries, and you can test them before saving.",
          ar: "نعم. الإعدادات ← أهداف الفحص فيها ثلاث قوائم: خوادم DNS (افتراضياً 8.8.8.8 و1.1.1.1)، وصفحات الويب، والأسماء التي يبحث عنها. كل قائمة تقبل حتى 10 عناصر، ويمكنك اختبارها قبل الحفظ.",
        },
      },
      {
        id: "data",
        q: { en: "How much internet data does it use?", ar: "كم يستهلك من باقة الإنترنت؟" },
        a: {
          en: "Regular checks are tiny: a connection, an empty web page and a name lookup. Speed tests are different. A test moves as much data as your line can in its time limit, so on a fast connection one test can use hundreds of megabytes. If your data is limited, don't schedule automatic speed tests.",
          ar: "الفحوصات العادية صغيرة جداً: اتصال، وصفحة ويب فارغة، وبحث عن اسم. قياس السرعة مختلف: ينقل أكبر قدر يستطيعه خطك خلال المهلة، فعلى اتصال سريع قد يستهلك القياس الواحد مئات الميجابايت. إذا كانت باقتك محدودة، لا تفعّل القياس التلقائي المجدول.",
        },
      },
    ],
  },
  {
    id: "alerts",
    accent: "amber",
    icon: BellRing,
    title: { en: "Alerts and sound", ar: "التنبيهات والصوت" },
    items: [
      {
        id: "where",
        q: { en: "Where do alerts show up?", ar: "أين تظهر التنبيهات؟" },
        a: {
          en: "As a system notification with a sound on this computer. You can also send them to a Discord or Slack channel with a webhook, or to Telegram. Set these up in Settings → Alerts and use the test button to make sure they arrive.",
          ar: "كإشعار نظام مع صوت على هذا الجهاز. ويمكنك أيضاً إرسالها إلى قناة Discord أو Slack عن طريق webhook، أو إلى Telegram. اضبطها من الإعدادات ← التنبيهات واستخدم زر التجربة لتتأكد أنها تصل.",
        },
      },
      {
        id: "telegram",
        q: { en: "How do I set up Telegram alerts?", ar: "كيف أفعّل تنبيهات Telegram؟" },
        a: {
          en: "Create a bot with @BotFather in Telegram and copy its token. Send your bot any message, then find your chat ID. Paste both into Settings → Alerts and press the test button.",
          ar: "أنشئ بوت عن طريق @BotFather في Telegram وانسخ التوكن. أرسل أي رسالة للبوت، ثم اعرف رقم المحادثة (chat ID). الصق الاثنين في الإعدادات ← التنبيهات واضغط زر التجربة.",
        },
      },
      {
        id: "sound",
        q: { en: "Why isn't my custom sound playing?", ar: "لماذا لا يعمل صوتي المخصص؟" },
        a: {
          en: "Open Settings → Sound. It shows whether your own sound is in use and how long it is. Files must be MP3 or WAV, up to 30 seconds and 10 MB. “Play like a real alert” plays it exactly as an alert would. If you hear nothing, check that the computer isn't muted.",
          ar: "افتح الإعدادات ← الصوت. ستجد إن كان صوتك هو المستخدم ومدته. الملف يجب أن يكون MP3 أو WAV، بحد أقصى 30 ثانية و10 ميجابايت. زر «شغّله كتنبيه حقيقي» يشغّله بنفس طريقة التنبيه. وإذا لم تسمع شيئاً، تأكد أن صوت الجهاز غير مكتوم.",
        },
      },
      {
        id: "silent",
        q: { en: "I didn't get a notification on Windows. Why?", ar: "لم يصلني إشعار على Windows، لماذا؟" },
        a: {
          en: "Windows hides notifications while Do not disturb or Focus is on, and it lets you turn them off per app. Check Settings → System → Notifications in Windows. The outage is still saved in History either way.",
          ar: "Windows يخفي الإشعارات أثناء وضع «عدم الإزعاج» أو «التركيز»، ويسمح بإيقافها لكل تطبيق. راجع الإعدادات ← النظام ← الإشعارات في Windows. وفي كل الأحوال، الانقطاع يُحفظ في السجل.",
        },
      },
    ],
  },
  {
    id: "reports",
    accent: "green",
    icon: FileText,
    title: { en: "Reports and speed tests", ar: "التقارير وقياس السرعة" },
    items: [
      {
        id: "proof",
        q: {
          en: "How do I send proof to my internet provider?",
          ar: "كيف أرسل إثباتاً لمزوّد الخدمة؟",
        },
        a: {
          en: "In History, press “Monthly report” and choose “Print or save as PDF”. It lists every outage with its time, duration and cause. For raw data, export CSV from History for one day, 7 days or 30 days.",
          ar: "من شاشة السجل اضغط «التقرير الشهري» واختر «اطبع أو احفظ PDF». فيه كل انقطاع بوقته ومدته وسببه. وللبيانات الخام، صدّر CSV من السجل ليوم أو 7 أيام أو 30 يوماً.",
        },
      },
      {
        id: "retention",
        q: { en: "How long is my history kept?", ar: "إلى متى يُحفظ السجل؟" },
        a: {
          en: "90 days. Older log files are deleted automatically. The logs live in the “logs” folder next to the app. Choose “Open Logs Folder” from the tray menu to find them.",
          ar: "90 يوماً. ملفات السجل الأقدم تُحذف تلقائياً. السجلات موجودة في مجلد «logs» بجانب التطبيق، واختر «Open Logs Folder» من قائمة شريط المهام للوصول إليها.",
        },
      },
      {
        id: "speedserver",
        q: { en: "Which server does the speed test use?", ar: "ما الخادم الذي يستخدمه قياس السرعة؟" },
        a: {
          en: "Cloudflare's speed test servers by default. You can point it at other servers in Settings → Speed test.",
          ar: "خوادم Cloudflare لقياس السرعة افتراضياً. ويمكنك تغييرها لخوادم أخرى من الإعدادات ← قياس السرعة.",
        },
      },
      {
        id: "schedule",
        q: { en: "Can speed tests run on their own?", ar: "هل يمكن أن يعمل قياس السرعة تلقائياً؟" },
        a: {
          en: "Yes. Set an interval in Settings → Speed test. You can also set a minimum speed. Results are sent to your Discord, Slack or Telegram alerts, and a result below that speed is marked as too slow.",
          ar: "نعم. حدد فاصلاً زمنياً من الإعدادات ← قياس السرعة. ويمكنك أيضاً تحديد سرعة دنيا. النتائج تُرسل لتنبيهات Discord أو Slack أو Telegram، والنتيجة الأقل من هذه السرعة تُعلَّم كبطيئة.",
        },
      },
    ],
  },
  {
    id: "privacy",
    accent: "coral",
    icon: ShieldCheck,
    title: { en: "Privacy and updates", ar: "الخصوصية والتحديثات" },
    items: [
      {
        id: "send",
        q: { en: "Does it send my data anywhere?", ar: "هل يرسل بياناتي لأي مكان؟" },
        a: {
          en: "No. There's no account, no tracking and no analytics. The only connections it makes are the checks themselves, speed tests, update checks on GitHub, and the Discord, Slack or Telegram alerts you set up. The Privacy page lists each one.",
          ar: "لا. لا يوجد حساب ولا تتبّع ولا تحليلات. الاتصالات الوحيدة هي الفحوصات نفسها، وقياس السرعة، والبحث عن تحديثات على GitHub، وتنبيهات Discord أو Slack أو Telegram التي تضبطها أنت. صفحة الخصوصية تذكر كل واحد منها.",
        },
      },
      {
        id: "lan",
        q: {
          en: "Can someone else on my network open the dashboard?",
          ar: "هل يقدر شخص آخر على شبكتي أن يفتح لوحة التحكم؟",
        },
        a: {
          en: "No. The dashboard only accepts connections from this computer, and even then it needs the key the app creates each time it starts.",
          ar: "لا. لوحة التحكم تقبل الاتصال من هذا الجهاز فقط، وحتى من هذا الجهاز تحتاج المفتاح الذي ينشئه التطبيق مع كل تشغيل.",
        },
      },
      {
        id: "secrets",
        q: { en: "Where are my settings stored?", ar: "أين تُحفظ إعداداتي؟" },
        a: {
          en: "In config.json next to the app. Your webhook URL and Telegram token are saved there as plain text, so don't share that file or that folder. The dashboard never sends them back to the browser.",
          ar: "في ملف config.json بجانب التطبيق. رابط الـ webhook وتوكن Telegram محفوظان فيه كنص عادي، فلا تشارك هذا الملف أو المجلد. ولوحة التحكم لا ترسلهما للمتصفح أبداً.",
        },
      },
      {
        id: "updates",
        q: { en: "How do updates work?", ar: "كيف تعمل التحديثات؟" },
        a: {
          en: "The app checks GitHub for a new version every 6 hours, or right away when you press “Check for updates”. Before installing, it checks the download against the release's checksums. Test (pre-release) versions are never offered automatically.",
          ar: "التطبيق يبحث عن إصدار جديد على GitHub كل 6 ساعات، أو فوراً عند الضغط على «ابحث عن تحديثات». وقبل التثبيت يطابق الملف المنزّل مع البصمات المنشورة مع الإصدار. والإصدارات التجريبية لا تُعرض تلقائياً أبداً.",
        },
      },
      {
        id: "bug",
        q: { en: "I found a bug. Where do I report it?", ar: "وجدت خطأ، أين أبلغ عنه؟" },
        a: {
          en: "Open an issue on the project's GitHub page. Say what you expected, what happened, and your app version (shown at the bottom of the sidebar). Don't paste your config.json, since it may contain your webhook or token. You can also reach us on WhatsApp, Discord or email from the Contact page.",
          ar: "افتح تذكرة (issue) على صفحة المشروع في GitHub. اذكر ما توقعته، وما حدث، ورقم إصدار التطبيق (يظهر أسفل القائمة الجانبية). ولا تلصق ملف config.json لأنه قد يحتوي على الـ webhook أو التوكن. ويمكنك أيضاً التواصل معنا على واتساب أو ديسكورد أو البريد من صفحة «تواصل معنا».",
        },
      },
    ],
  },
];
