# 🌐 مراقب الإنترنت

[![Build & Release](https://github.com/FutureSolutionDev/internet-monitor/actions/workflows/build.yml/badge.svg)](https://github.com/FutureSolutionDev/internet-monitor/actions/workflows/build.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> [English README](README.md)

---

**أداة مفتوحة المصدر لمراقبة استقرار اتصال الإنترنت في الوقت الفعلي.**

تعمل بصمت في الخلفية، تسجّل كل انقطاع مع سببه ومدّته، وتعرض لوحة تحكم مرئية في المتصفح مع رسوم بيانية حيّة وإشعارات فورية.

---

## 💡 الفكرة والهدف

كثيرًا ما يشكو المستخدمون من تقطّع الإنترنت دون أن يملكوا دليلًا ملموسًا — لا توقيت الانقطاع، ولا سببه، ولا مدّته. **مراقب الإنترنت** يحلّ هذه المشكلة عمليًا:

- **المستخدم العادي** — يعمل تلقائيًا ويُرسل إشعارًا فوريًا عند كل انقطاع
- **فريق الدعم التقني** — لوحة بيانات كاملة وسجلات قابلة للتصدير كـ CSV
- **المطوّر** — Webhook مفصّل لـ Discord/Slack مع كامل بيانات الفحص

---

## ✨ المميزات

| الميزة | التفاصيل |
| ------ | -------- |
| 🔍 فحص متعدد الطبقات | TCP Ping + HTTP + DNS في نفس الوقت، كلها قابلة للتخصيص كـ arrays |
| 📊 لوحة تحكم حيّة | رسوم بيانية + إحصائيات + سجل الأحداث مع Tooltips |
| 🔔 إشعارات فورية | Windows Toast + Discord/Slack Webhook + صوت تنبيه مخصّص |
| 📋 سجلات منظّمة | JSONL يومي قابل للتصدير كـ CSV من لوحة التحكم |
| 🔄 تحديث تلقائي | يفحص GitHub Releases ويحدّث نفسه بنقرة واحدة |
| 🌐 ثنائي اللغة | عربي وإنجليزي مع دعم RTL كامل |
| 🖥️ نسختان | Tray (خلفية) + نافذة مستقلة Native |
| 🔒 نسخة واحدة | يمنع فتح أكثر من نسخة في نفس الوقت |

---

## 🚀 التثبيت السريع

> **ظهور تحذير Windows SmartScreen؟** اضغط **"مزيد من المعلومات" ← "تشغيل على أي حال"**. الملف غير موقّع رقمياً — وهذا أمر طبيعي للأدوات مفتوحة المصدر، والكود المصدري متاح كاملاً للمراجعة.

حمّل الملف المناسب من [Releases](https://github.com/FutureSolutionDev/internet-monitor/releases/latest):

| الملف | النظام | النوع |
| ----- | ------ | ----- |
| `internet-monitor-windows.exe` | Windows 10/11 | Tray — يعمل في الخلفية |
| `internet-monitor-gui-windows.exe` | Windows 10/11 | نافذة مستقلة |
| `internet-monitor-macos-arm64` | macOS M1/M2/M3 | Tray |
| `internet-monitor-macos-intel` | macOS Intel | Tray |
| `internet-monitor-linux` | Ubuntu/Debian | Tray |

**Windows** — بعد التحميل:

1. ضع الملف في مجلد ثابت مثل `C:\Tools\InternetMonitor\`
2. انقر عليه مرتين — ستظهر أيقونة في شريط المهام
3. كليك يمين على الأيقونة ← **Open Dashboard**

افتح لوحة التحكم على: **<http://localhost:8765>**

للتشغيل التلقائي مع Windows:

```bat
scripts\install.cmd
```

**macOS / Linux:**

```bash
chmod +x internet-monitor-*
./internet-monitor-macos-arm64
```

---

## 🧑‍💻 التشغيل في وضع التطوير

لا حاجة لبناء مسبق — شغّل مباشرة من المصدر:

```bash
git clone https://github.com/FutureSolutionDev/internet-monitor.git
cd internet-monitor
go mod tidy
```

**تشغيل (نسخة Tray — بدون GCC):**

```bat
scripts\run.cmd
```

أو يدوياً:

```bash
go run .
```

**إيقاف جميع النسخ الشغّالة:**

```bat
scripts\stop.cmd
```

افتح لوحة التحكم على **<http://localhost:8765>** — عدّل أي ملف `.go` وأعد التشغيل لترى التغييرات.

> نسخة النافذة المستقلة (GUI) تحتاج GCC — راجع **البناء من المصدر** أدناه.

---

## 🛠️ البناء من المصدر

### المتطلبات

| الأداة | الإصدار | ملاحظة |
| ------ | ------- | ------ |
| [Go](https://go.dev/dl/) | 1.21+ | مطلوب |
| GCC | أي إصدار | اختياري — للنافذة المستقلة فقط |

### Tray Version — بدون CGO (الأسهل)

```bash
git clone https://github.com/FutureSolutionDev/internet-monitor.git
cd internet-monitor
go mod tidy
go build -ldflags="-H=windowsgui -s -w" -o internet-monitor.exe .
```

### Native Window Version — يحتاج GCC

**Windows** — السكريبت يثبّت GCC تلقائيًا إن لم يكن موجودًا:

```bat
scripts\build-gui.cmd
```

**macOS / Linux** — GCC موجود افتراضيًا:

```bash
go build -ldflags="-s -w" -o internet-monitor-gui ./cmd/gui/
```

### السكريبتات المتاحة

```text
scripts\build.cmd        بناء tray exe
scripts\build-gui.cmd    بناء نافذة native (يثبّت GCC تلقائيًا إن لزم)
scripts\build-debug.cmd  بناء بـ console ظاهر (للتشخيص)
scripts\run.cmd          بناء وتشغيل
scripts\stop.cmd         إيقاف النسخة الشغّالة
scripts\install.cmd      تثبيت مع Windows Startup
scripts\uninstall.cmd    إزالة من Startup
scripts\logs.cmd         فتح مجلد السجلات
```

---

## ⚙️ الإعدادات (`config.json`)

يُنشأ تلقائيًا أول تشغيل. عدّله من **تبويب الإعدادات** في لوحة التحكم أو مباشرةً:

```json
{
  "check_interval_sec": 5,
  "ping_targets":  ["8.8.8.8:53", "1.1.1.1:53"],
  "http_targets":  ["https://connectivitycheck.gstatic.com/generate_204"],
  "dns_targets":   ["www.google.com", "www.cloudflare.com"],
  "fail_threshold": 3,
  "packet_loss_threshold": 20.0,
  "latency_threshold_ms": 500,
  "log_dir": "logs",
  "webhook_url": "",
  "dashboard_port": 8765
}
```

| الحقل | المعنى |
| ----- | ------ |
| `ping_targets` | عناوين TCP Ping — يجرّب كل العناوين، يكفي نجاح واحد |
| `http_targets` | روابط HTTP — يجرّب بالترتيب حتى يجد ناجحًا (200/204) |
| `dns_targets` | نطاقات DNS — يجرّب بالترتيب حتى يُحلّ أحدها |
| `fail_threshold` | عدد الفشل المتتالي قبل الإعلان عن الانقطاع |
| `webhook_url` | رابط Discord أو Slack (فارغ = معطّل) |

---

## 📡 Webhook — Discord و Slack فقط

يدعم **Discord** و **Slack** فقط. الـ payload يُنسَّق تلقائيًا حسب نوع الرابط مع كامل تفاصيل الفحص.

**مثال Discord embed عند الانقطاع:**

```json
{
  "username": "Internet Monitor",
  "embeds": [{
    "title": "❌ Internet Disconnected",
    "color": 15681604,
    "fields": [
      {"name": "🔌 TCP Ping",    "value": "❌ Failed", "inline": true},
      {"name": "🌐 HTTP",        "value": "❌ Failed", "inline": true},
      {"name": "🔍 DNS",         "value": "✅ OK",     "inline": true},
      {"name": "📉 Packet Loss", "value": "85.0%",    "inline": true},
      {"name": "⏱️ Duration",    "value": "2m 15s",   "inline": true}
    ]
  }]
}
```

---

## 📂 هيكل المشروع

```text
internet-monitor/
├── main.go                  نقطة دخول نسخة Tray
├── singleton_*.go           منع تشغيل أكثر من نسخة (حسب النظام)
├── cmd/gui/                 نسخة النافذة المستقلة
├── config/                  تحميل الإعدادات والـ migration التلقائي
├── monitor/                 محرك الفحص — TCP / HTTP / DNS
├── dashboard/               HTTP server + SSE + REST APIs
│   └── assets/              HTML / CSS / JS مدمج (Chart.js محلي)
├── logger/                  JSONL + تنسيق Webhook لـ Discord/Slack
├── tray/                    الأيقونة المدمجة مع الـ favicon
├── updater/                 GitHub Releases API + selfupdate
├── .github/workflows/       CI/CD — build.yml
└── scripts/                 سكريبتات Windows للبناء والتثبيت
```

---

## 🤝 المساهمة في المشروع

كل أنواع المساهمات مرحّب بها — إصلاح أخطاء، ميزات جديدة، ترجمات، توثيق.

### 1. Fork و Clone

```bash
git clone https://github.com/YOUR_USERNAME/internet-monitor.git
cd internet-monitor
go mod tidy
```

### 2. إنشاء Branch

```bash
git checkout -b feature/اسم-الميزة
```

### 3. صيغة Commit Messages

نتّبع [Conventional Commits](https://www.conventionalcommits.org/). نوع الـ commit يحدّد رقم الإصدار التالي تلقائيًا:

| النوع | الإصدار | مثال |
| ----- | ------- | ---- |
| `feat:` | minor (v1.1.0) | `feat: add dark mode toggle` |
| `fix:` | patch (v1.0.1) | `fix: DNS timeout on macOS` |
| `BREAKING CHANGE` | major (v2.0.0) | في نص الـ commit |
| `docs:`, `refactor:` | patch | بدون تغيير وظيفي |

### 4. فتح Pull Request

- صِف الهدف من التغيير بوضوح
- أرفق لقطة شاشة إذا كان هناك تغيير مرئي
- تأكّد أن `go build ./...` ينجح بدون أخطاء

---

## 📋 ملفات السجل

```text
logs/
  connectivity_2026-05-11.jsonl   حدث لكل سطر
  app.log                          أخطاء التطبيق وحالة الـ Webhook
```

---

## 📄 الترخيص

MIT — مجاني للاستخدام الشخصي والتجاري.
