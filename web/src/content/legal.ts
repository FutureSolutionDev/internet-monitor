import type { Text } from "./text";

export interface LegalSection {
  id: string;
  title: Text;
  body: Text[];
  list?: Text[];
  quote?: Text;
}

export interface LegalDoc {
  updated: string;
  intro: Text;
  sections: LegalSection[];
}

export const PRIVACY: LegalDoc = {
  updated: "2026-10-24",
  intro: {
    en: "Internet Monitor runs on your computer and keeps your data there. This page lists everything it stores and every place it connects to.",
    ar: "مراقب الإنترنت يعمل على جهازك ويحتفظ ببياناتك عليه. هذه الصفحة تذكر كل ما يحفظه وكل مكان يتصل به.",
  },
  sections: [
    {
      id: "short",
      title: { en: "The short version", ar: "باختصار" },
      body: [
        {
          en: "There's no account, no analytics, no ads and no tracking. The people who make Internet Monitor can't see your history, your settings or your alerts.",
          ar: "لا يوجد حساب، ولا تحليلات، ولا إعلانات، ولا تتبّع. صُنّاع مراقب الإنترنت لا يرون سجلك ولا إعداداتك ولا تنبيهاتك.",
        },
      ],
    },
    {
      id: "stored",
      title: { en: "What's stored on your computer", ar: "ما يُحفظ على جهازك" },
      body: [
        {
          en: "Everything is kept in the app's folder, except two small preferences kept by your browser.",
          ar: "كل شيء محفوظ في مجلد التطبيق، ما عدا تفضيلين صغيرين يحفظهما المتصفح.",
        },
      ],
      list: [
        {
          en: "config.json: your settings. This includes your webhook URL, Telegram token and chat ID, saved as plain text. Don't share this file.",
          ar: "config.json: إعداداتك، ومنها رابط الـ webhook وتوكن Telegram ورقم المحادثة، محفوظة كنص عادي. لا تشارك هذا الملف.",
        },
        {
          en: "The logs folder: every check, outage and speed test result, plus app.log, a diagnostic log of what the app did. Files older than 90 days are deleted automatically.",
          ar: "مجلد logs: كل فحص وانقطاع ونتيجة قياس سرعة، بالإضافة إلى app.log وهو سجل تشخيصي لما قام به التطبيق. الملفات الأقدم من 90 يوماً تُحذف تلقائياً.",
        },
        {
          en: "notification.wav: your custom alert sound, if you uploaded one.",
          ar: "notification.wav: صوت التنبيه المخصص، إن كنت قد رفعت صوتاً.",
        },
        {
          en: ".session: the dashboard key for the current launch. It's replaced every time the app starts.",
          ar: ".session: مفتاح لوحة التحكم للتشغيل الحالي، ويُستبدل مع كل تشغيل للتطبيق.",
        },
        {
          en: "Your browser remembers your theme and language, so the dashboard opens the way you left it.",
          ar: "المتصفح يتذكر الثيم واللغة، لتفتح لوحة التحكم كما تركتها.",
        },
      ],
    },
    {
      id: "network",
      title: { en: "What leaves your computer", ar: "ما يخرج من جهازك" },
      body: [
        {
          en: "The app only connects to the places below. Each of them sees your IP address, as with any internet connection, and its own privacy policy applies.",
          ar: "التطبيق يتصل فقط بالأماكن التالية. كل منها يرى عنوان IP الخاص بك كأي اتصال بالإنترنت، وتنطبق عليه سياسة الخصوصية الخاصة به.",
        },
      ],
      list: [
        {
          en: "Connection checks go to the targets in Settings → Check targets. By default: Google and Cloudflare DNS servers (8.8.8.8 and 1.1.1.1), connectivitycheck.gstatic.com, and name lookups for www.google.com and www.cloudflare.com. It also checks your own router.",
          ar: "فحوصات الاتصال تذهب للأهداف الموجودة في الإعدادات ← أهداف الفحص. افتراضياً: خوادم DNS من Google وCloudflare ‏(8.8.8.8 و1.1.1.1)، وconnectivitycheck.gstatic.com، والبحث عن www.google.com وwww.cloudflare.com. ويفحص أيضاً الراوتر الخاص بك.",
        },
        {
          en: "Speed tests go to speed.cloudflare.com by default, and only when you start one or schedule them. The test data contains nothing that identifies your device.",
          ar: "قياس السرعة يذهب إلى speed.cloudflare.com افتراضياً، وفقط عندما تبدأه أو تجدوله. بيانات القياس لا تحتوي على أي شيء يعرّف جهازك.",
        },
        {
          en: "Update checks go to GitHub (api.github.com) every 6 hours, or when you press “Check for updates”. Installing an update downloads it from GitHub.",
          ar: "البحث عن التحديثات يذهب إلى GitHub ‏(api.github.com) كل 6 ساعات، أو عند الضغط على «ابحث عن تحديثات». وتثبيت التحديث يُنزّله من GitHub.",
        },
        {
          en: "Alerts go to Discord, Slack or Telegram only if you set them up. They contain the status, times, durations and causes of outages, or a speed test result. The app doesn't add your computer's name or IP address.",
          ar: "التنبيهات تذهب إلى Discord أو Slack أو Telegram فقط إذا فعّلتها. تحتوي على الحالة، والأوقات، والمدد، وأسباب الانقطاع، أو نتيجة قياس السرعة. والتطبيق لا يضيف اسم جهازك ولا عنوان IP.",
        },
      ],
    },
    {
      id: "dashboard",
      title: { en: "The dashboard", ar: "لوحة التحكم" },
      body: [
        {
          en: "The dashboard is served by the app on this computer (127.0.0.1). Other devices on your network can't open it. It uses one cookie, which holds the key for the current launch. It loads no outside scripts, fonts or images.",
          ar: "لوحة التحكم يقدّمها التطبيق على هذا الجهاز (127.0.0.1)، ولا تستطيع الأجهزة الأخرى على شبكتك فتحها. تستخدم ملف تعريف ارتباط (cookie) واحداً يحمل مفتاح التشغيل الحالي، ولا تحمّل أي سكربتات أو خطوط أو صور من الخارج.",
        },
      ],
    },
    {
      id: "control",
      title: { en: "You're in control", ar: "التحكم بيدك" },
      body: [
        {
          en: "You can delete the logs folder at any time to erase your history. Deleting config.json resets your settings. To remove the app completely, turn off “Start with Windows” in Settings → App, quit the app, and delete its folder.",
          ar: "يمكنك حذف مجلد logs في أي وقت لمسح سجلك، وحذف config.json يعيد الإعدادات للوضع الافتراضي. ولإزالة التطبيق بالكامل: أوقف «التشغيل مع Windows» من الإعدادات ← التطبيق، ثم أغلق التطبيق واحذف مجلده.",
        },
      ],
    },
    {
      id: "changes",
      title: { en: "Changes to this page", ar: "تغييرات هذه الصفحة" },
      body: [
        {
          en: "If a new version changes what the app stores or where it connects, this page is updated and the change is listed in What's new. The date at the top shows the last update.",
          ar: "إذا غيّر إصدار جديد ما يحفظه التطبيق أو الأماكن التي يتصل بها، تُحدَّث هذه الصفحة ويُذكر التغيير في «ما الجديد». التاريخ في الأعلى يوضّح آخر تحديث.",
        },
      ],
    },
    {
      id: "contact",
      title: { en: "Questions", ar: "أسئلة" },
      body: [
        {
          en: "Reach us through any channel on the Contact page. Don't send your config.json or logs unless we ask, since they may contain your webhook, token or history.",
          ar: "تواصل معنا عبر أي وسيلة في صفحة «تواصل معنا». ولا ترسل ملف config.json أو السجلات إلا إذا طلبناها، لأنها قد تحتوي على الـ webhook أو التوكن أو سجلك.",
        },
      ],
    },
  ],
};

export const TERMS: LegalDoc = {
  updated: "2026-10-24",
  intro: {
    en: "Internet Monitor is free, open-source software under the MIT license. This page explains what that means in plain language. If anything here differs from the license, the license applies.",
    ar: "مراقب الإنترنت برنامج مجاني ومفتوح المصدر برخصة MIT. هذه الصفحة تشرح معنى ذلك بلغة بسيطة، وإذا اختلف شيء هنا عن نص الرخصة، فالرخصة هي المعتمدة.",
  },
  sections: [
    {
      id: "use",
      title: { en: "What you may do", ar: "ما يحق لك فعله" },
      body: [
        {
          en: "You may use Internet Monitor for free, at home or at work. You may also copy, change, share, or sell copies of it. When you share it or a changed version, keep the copyright and license notice with it.",
          ar: "يحق لك استخدام مراقب الإنترنت مجاناً، في البيت أو العمل. ويحق لك نسخه وتعديله ومشاركته وبيع نسخ منه. وعند مشاركته أو مشاركة نسخة معدّلة، أبقِ معه إشعار حقوق النشر والرخصة.",
        },
      ],
    },
    {
      id: "warranty",
      title: { en: "No warranty", ar: "بلا ضمان" },
      body: [
        {
          en: "The app is provided as is. The license says:",
          ar: "التطبيق مقدَّم كما هو. وهذه ترجمة لنص الرخصة، يليها النص الإنجليزي الأصلي وهو المعتمد:",
        },
      ],
      quote: {
        en: "THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.",
        ar: "البرنامج مقدَّم «كما هو»، بدون أي ضمان صريح أو ضمني، بما في ذلك على سبيل المثال لا الحصر ضمانات القابلية للتسويق، والملاءمة لغرض معيّن، وعدم التعدّي. ولا يتحمّل المؤلفون أو أصحاب حقوق النشر في أي حال المسؤولية عن أي مطالبة أو أضرار أو التزام آخر، سواء في دعوى تعاقدية أو تقصيرية أو غير ذلك، تنشأ عن البرنامج أو استخدامه أو التعامل معه أو تتعلق بذلك.",
      },
    },
    {
      id: "measurements",
      title: { en: "About the measurements", ar: "عن القياسات" },
      body: [
        {
          en: "Results are measured from your computer. Wi-Fi signal, other devices on your network, a firewall, or the computer going to sleep can all affect them. Reports are a record of what your computer saw, not a certified measurement, and your provider may or may not accept them.",
          ar: "النتائج تُقاس من جهازك، وقد تتأثر بإشارة الـ Wi-Fi، أو الأجهزة الأخرى على شبكتك، أو الجدار الناري، أو دخول الجهاز في وضع السكون. التقارير سجلّ لما رآه جهازك، وليست قياساً معتمداً، وقد يقبلها المزوّد أو لا يقبلها.",
        },
      ],
    },
    {
      id: "fair",
      title: { en: "Use it fairly", ar: "استخدمه بإنصاف" },
      body: [
        {
          en: "Only point checks and speed tests at servers you're allowed to use. Don't set very short intervals against servers you don't run. When you use Cloudflare, GitHub, Discord, Slack or Telegram through the app, their own terms apply.",
          ar: "وجّه الفحوصات وقياسات السرعة فقط لخوادم مسموح لك باستخدامها، ولا تضبط فواصل قصيرة جداً على خوادم لا تديرها. وعند استخدام Cloudflare أو GitHub أو Discord أو Slack أو Telegram عبر التطبيق، تنطبق شروطها الخاصة.",
        },
      ],
    },
    {
      id: "changes",
      title: { en: "Changes", ar: "التغييرات" },
      body: [
        {
          en: "These terms may change in a new version. The date at the top shows the last update, and changes are listed in What's new.",
          ar: "قد تتغيّر هذه الشروط مع إصدار جديد. التاريخ في الأعلى يوضّح آخر تحديث، والتغييرات تُذكر في «ما الجديد».",
        },
      ],
    },
    {
      id: "contact",
      title: { en: "Questions", ar: "أسئلة" },
      body: [
        {
          en: "Reach us through any channel on the Contact page.",
          ar: "تواصل معنا عبر أي وسيلة في صفحة «تواصل معنا».",
        },
      ],
    },
  ],
};
