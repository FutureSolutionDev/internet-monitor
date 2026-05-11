# مراقب الإنترنت — Internet Monitor

أداة مجانية بتشتغل في الخلفية وبتراقب اتصال النت عندك.
كل ما النت بينقطع أو بيتعمل حاجة، بتسجّل الحدث تلقائي بالوقت والسبب والمدة، وفيه dashboard في المتصفح بالبيانات live.

---

## بتعمل إيه بالظبط؟

- بتفحص النت كل 5 ثواني (ممكن تعدّلها)
- بتكتشف 3 حالات:
  - 🟢 **متصل** — كل حاجة تمام
  - 🟡 **ضعيف** — في فقدان حزم أو latency عالي
  - 🔴 **منقطع** — النت وقع خالص
- بتسجّل كل حدث في ملف يومي بصيغة JSONL
- بتبعت إشعار Windows لما النت يوقع أو يرجع
- ممكن تربطها بـ Discord / Slack / أي webhook
- فيها dashboard في المتصفح فيه charts وإحصائيات live

---

## التثبيت والتشغيل

### الطريقة السريعة (للمستخدم العادي)

1. حمّل ملف `internet-monitor-windows.exe`
2. حطّه في مجلد ثابت — مثلاً: `C:\Tools\InternetMonitor\`
3. حمّل `config.json` وحطّه جنبه في نفس المجلد
4. دبّل كليك على `internet-monitor-windows.exe`
5. هتلاقي أيقونة اتضافت في شريط المهام (System Tray)
6. كليك يمين على الأيقونة ← **Open Dashboard** تفتح اللوحة في المتصفح

### عايز يشتغل تلقائي مع Windows؟

```bat
scripts\install.cmd
```

ده بيثبّت البرنامج ويضيفه في الـ startup عشان يشتغل كل ما تشغّل الجهاز.

---

## أيقونة الـ Tray معناها إيه؟

| الأيقونة | يعني إيه |
| -------- | -------- |
| 🟢 خضرا | النت شغّال وكويس |
| 🟡 صفرا | الاتصال ضعيف أو بطيء |
| 🔴 حمرا | النت وقع |

---

## لوحة التحكم (Dashboard)

افتح المتصفح على: **<http://localhost:8765>**

أو كليك يمين على الأيقونة في الـ Tray ← **Open Dashboard**

### اللوحة فيها إيه؟

**تبويب لوحة التحكم:**

- بطاقة الحالة الكبيرة + زمن الاستجابة
- درجة جودة الاتصال (A / B / C / D / F)
- 5 بطاقات: وقت التشغيل، نسبة الاتصال، الانقطاعات، متوسط الاستجابة، عدد الفحوصات
- رسم بياني live لزمن الاستجابة — لو وقفت الماوس عليه هيظهر tooltip
- حالة فحوصات TCP / HTTP / DNS ونسبة فقدان الحزم
- جدول آخر الأحداث

**تبويب السجلات:**

- اختار أي يوم تشوف سجله كامل
- كل حدث فيه: الوقت، النوع، مدة الانقطاع، السبب، فقدان الحزم، زمن الاستجابة
- زرار تصدير CSV

**تبويب الإعدادات:**

- تعدّل كل الإعدادات من المتصفح مباشرةً
- فيه زرار اختبار لكل عنوان قبل الحفظ
- الحفظ فوري من غير ما تعيد التشغيل (ما عدا المنفذ)

---

## ملف الإعدادات (config.json)

```json
{
  "check_interval_sec": 5,
  "ping_targets": ["8.8.8.8:53", "1.1.1.1:53"],
  "http_target": "https://connectivitycheck.gstatic.com/generate_204",
  "dns_target": "www.google.com",
  "fail_threshold": 3,
  "packet_loss_threshold": 20.0,
  "latency_threshold_ms": 500,
  "log_dir": "logs",
  "webhook_url": "",
  "dashboard_port": 8765
}
```

| الإعداد | معناه | القيمة الافتراضية |
| ------- | ----- | ----------------- |
| `check_interval_sec` | كل كام ثانية بيتم الفحص | 5 |
| `fail_threshold` | كام فشل متتالي قبل ما يعتبره "منقطع" | 3 |
| `packet_loss_threshold` | نسبة فقدان الحزم اللي بيعتبرها "ضعيف" | 20% |
| `latency_threshold_ms` | زمن استجابة بيعتبره "ضعيف" | 500ms |
| `webhook_url` | رابط Discord/Slack للإشعارات | فاضي |
| `dashboard_port` | بورت لوحة التحكم | 8765 |

---

## Webhook — تبعت الإشعارات على Discord أو Slack

حط رابط الـ webhook في إعداد `webhook_url`.

عند كل انقطاع أو رجوع بيوصل payload زي ده:

```json
{
  "timestamp": "2026-05-11T14:30:00Z",
  "event": "disconnected",
  "duration_seconds": 45.2,
  "reason": {
    "tcp_ping_failed": true,
    "http_failed": true,
    "dns_failed": false,
    "packet_loss_pct": 80.0,
    "avg_latency_ms": 0
  }
}
```

لو عملت اختبار يدوي من تبويب الإعدادات، بيبعت كمان payload بنتيجة الاختبار.

---

## ملفات السجل

بتتحفظ في مجلد `logs/` — ملف لكل يوم:

```text
logs/
  connectivity_2026-05-11.jsonl
  connectivity_2026-05-12.jsonl
  ...
```

كل سطر حدث لوحده — ممكن تفتحه بأي text editor.

---

## درجة جودة الاتصال

| الدرجة | يعني إيه |
| ------ | -------- |
| **A — ممتاز** | أكتر من 95% وقت متصل، latency أقل من 200ms |
| **B — جيد** | أكتر من 95% لكن الاتصال شوية بطيء |
| **C — متوسط** | نسبة الاتصال 80-95% |
| **D — ضعيف** | نسبة الاتصال 50-80% |
| **F — حرج** | أقل من 50% وقت متصل |

---

## الأوامر (للمطورين)

```bash
scripts\build.cmd        # بناء tray exe
scripts\build-gui.cmd    # بناء نافذة native (محتاج TDM-GCC)
scripts\build-debug.cmd  # بناء بـ console ظاهر للتشخيص
scripts\run.cmd          # تشغيل مباشر
scripts\stop.cmd         # إيقاف
scripts\install.cmd      # تثبيت مع Windows Startup
scripts\uninstall.cmd    # إزالة التثبيت
scripts\logs.cmd         # فتح مجلد السجلات
```

أو باستخدام `make`:

```bash
make build
make build-gui
make install
make stop
make logs
```

---

## البناء من المصدر

**Tray version (بدون CGO):**

```bash
go build -ldflags="-H=windowsgui -s -w" -o internet-monitor.exe .
```

**GUI version (محتاج TDM-GCC على Windows):**

```bash
winget install TDMGcc.TDMGcc
go build -ldflags="-H=windowsgui -s -w" -o internet-monitor-gui.exe ./cmd/gui/
```

**على macOS/Linux** — gcc موجود بالفعل:

```bash
go build -ldflags="-s -w" -o internet-monitor-gui ./cmd/gui/
```

---

## تحميل الإصدارات الجاهزة

**[GitHub Releases](../../releases)** — فيها ملفات جاهزة لكل نظام:

| الملف | النظام |
| ----- | ------ |
| `internet-monitor-windows.exe` | Windows 10/11 — Tray |
| `internet-monitor-gui-windows.exe` | Windows 10/11 — نافذة |
| `internet-monitor-macos-arm64` | macOS M1/M2/M3 — Tray |
| `internet-monitor-macos-intel` | macOS Intel — Tray |
| `internet-monitor-gui-macos-arm64` | macOS M1/M2/M3 — نافذة |
| `internet-monitor-gui-macos-intel` | macOS Intel — نافذة |
| `internet-monitor-linux` | Linux — Tray |
| `internet-monitor-gui-linux` | Linux — نافذة |

---

## الترخيص

MIT — مجاني للاستخدام الشخصي والتجاري.
