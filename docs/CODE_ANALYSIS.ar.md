# تحليل شامل لمشروع Internet Monitor

> مستند مرجعي: مشاكل، أخطاء، تحسينات، وخريطة تطوير (Upgrade Roadmap) + أفكار مميزات جديدة.
> تمّ في هذا الـ branch تنفيذ **الإصلاحات الحرجة الآمنة** فقط؛ باقي البنود مقترحات للمراحل القادمة.

---

## 1. نظرة عامة على المعمارية

المشروع أداة Go لمراقبة الإنترنت، مبنية بشكل نظيف ومعماريتها سليمة بشكل عام:

| الحزمة | المسؤولية |
| --- | --- |
| `monitor` | الفحص الفعلي (TCP ping / HTTP / DNS) |
| `core` | محرك المراقبة (`Engine`) + حساب الحالة (`DetermineStatus`) + توزيع الأحداث (`MultiNotifier`) |
| `dashboard` | سيرفر HTTP + REST API + SSE + الواجهة (assets) |
| `logger` | كتابة logs بصيغة JSONL + إرسال webhooks (Discord/Slack) |
| `speedtest` | اختبار سرعة التحميل التكيّفي |
| `updater` | التحقق من التحديثات + التحديث الذاتي عبر GitHub Releases |
| `tray` / `cmd/gui` | أيقونة شريط المهام + نسخة GUI (WebView) |
| `config` / `types` | الإعدادات والأنواع المشتركة |

**نقاط قوة:** فصل واضح للمسؤوليات، واجهة `Notifier` قابلة للتوسّع، استخدام صحيح للـ `context`، `embed.FS` للأصول، ربط بـ `127.0.0.1` فقط (آمن افتراضياً)، دعم متعدد المنصّات.

**المشكلة الجوهرية:** الإعدادات تُقرأ في أماكن متعددة (في الذاكرة + من الملف مباشرة)، وتعديلها من الواجهة لم يكن يُطبَّق على المحرك الشغّال — وهي أهم نقطة عالَجناها.

---

## 2. المشاكل والأخطاء

### 2.1 حرجة (Critical) — تمّ إصلاحها في هذا الـ branch ✅

| # | المشكلة | الموضع | الأثر |
| --- | --- | --- | --- |
| C1 | **تعديل الإعدادات من الواجهة لا يُطبَّق إلا بعد إعادة التشغيل.** `OnConfigChange` معرَّف ومُستدعى في السيرفر لكنه **لم يكن مربوطًا** في `main.go` ولا `cmd/gui/main.go`. | `dashboard/server.go:191,379` + `main.go` + `cmd/gui/main.go` | المستخدم يغيّر الأهداف/العتبات/الفاصل الزمني فتُحفظ على القرص بينما المحرك يستمر بالقديم. |
| C2 | **انهيار (panic) محتمل عند `check_interval_sec = 0`.** `time.NewTicker(0)` يعمل panic، والواجهة كانت تسمح بحفظ 0. | `core/engine.go` + `config/config.go` | crash-loop عند بدء التشغيل بعد حفظ قيمة غير صالحة. |
| C3 | **`http.ListenAndServe` يبتلع الخطأ بصمت.** | `dashboard/server.go` | لو البورت مشغول، الداشبورد يموت دون أي إشارة بينما التطبيق يكمل عمله "بلا واجهة". |

### 2.2 متوسطة (Medium)

| # | المشكلة | الموضع | ملاحظة |
| --- | --- | --- | --- |
| M1 | **سباق بيانات (data race) محتمل على الإعدادات** عند التطبيق الحي (الـ checker/logger/engine يشاركون نفس `*config.Config`). | عام | عولج ضمن حل C1: التغيير يُمرَّر عبر goroutine المحرك + `Logger.SetConfig` تحت mutex. |
| M2 | **لا يوجد graceful shutdown للسيرفر** (لا توجد إشارة `http.Server` ولا `Shutdown`)؛ عملاء SSE لا يُغلَقون بنظام عند الخروج. | `dashboard/server.go` | مقترح Phase 0. |
| M3 | **قراءة هشّة للـ body** في `serveTestWebhook` (`Read` واحدة في buffer بحجم 2048). | `dashboard/server.go` | ✅ أُصلح: `io.ReadAll` + `MaxBytesReader`. |
| M4 | **لا يوجد تحقق على `/api/config` POST**؛ أي قيم سالبة/خارج المدى كانت تُحفظ. | `dashboard/server.go` | ✅ أُصلح جزئيًا عبر `config.Sanitize()`. |
| M5 | **صفر اختبارات** في المشروع كله. | — | ✅ بدأنا بإضافة اختبارات للوحدات النقية. |
| M6 | **عدم تطابق نسخة Go في الـ CI:** الـ workflow يستخدم `GO_VERSION: '1.21'` بينما `go.mod` يطلب `go 1.24.0`. | `.github/workflows/build.yml` | يؤدي لتنزيل toolchain 1.24 تلقائيًا (بطء) أو فشل حسب إعداد `GOTOOLCHAIN`. |
| M7 | **التحديث الذاتي دون تحقق توقيع/checksum.** التنزيل عبر HTTPS من GitHub لكن `selfupdate.Apply(..., Options{})` بلا verifier. مكتبة `minisign` موجودة كاعتماد لكنها غير مستخدمة. | `updater/updater.go` | خطر سلسلة التوريد؛ مقترح Phase 0/1. |

### 2.3 أمان (Security)

| # | المشكلة | ملاحظة |
| --- | --- | --- |
| S1 | **REST API على localhost بلا مصادقة ولا حماية CSRF/Origin.** صفحة ويب خبيثة قد تحاول POST إلى `/api/config` أو `/api/update` أو `/api/speed-test/start`. الـ `application/json` يفرض preflight يحمي جزئيًا، لكن لا يوجد فحص `Origin` ولا حماية من DNS rebinding. | إضافة فحص `Origin`/`Host` + token اختياري — Phase 1. |
| S2 | **رفع ملف الصوت** يقبل أي محتوى حتى 10MB دون تحقق نوع MIME. | تحقق امتداد/MIME — Phase 1. |
| S3 | الواجهة تستخدم `escHtml` لمدخلات المستخدم في الإعدادات (جيّد)، لكن أصناف الشارات `badge-${evType}` تُحقن دون escape (المصدر بيانات التطبيق فالمخاطر منخفضة). | تحسين بسيط. |

### 2.4 بسيطة وجودة الكود (Minor / Quality)

- **L1** قياس الـ latency غير دقيق: يُحسب من بداية اللوب، فإذا فشل أول هدف ping تُضاف مدة الـ timeout (حتى 2 ثانية) للقياس — `monitor/checker.go`.
- **L2** `os.Chdir` يتجاهل الخطأ والتطبيق يعتمد على الـ cwd — `main.go`.
- **L3** `strings.Title` مهجور (deprecated) — `logger/webhook_format.go` (البديل `golang.org/x/text/cases`).
- **L4** الكود ليس `gofmt`-clean (ترتيب imports، محاذاة structs، مسافات تعليقات).
- **L5** الحقل `stLast` يبدو غير مستخدَم بعد كتابته — `dashboard/server.go`.
- **L6** لا يوجد تدوير/تنظيف للـ logs؛ ملفات `connectivity_*.jsonl` و`speedtest_*.jsonl` تتراكم بلا حدود.
- **L7** تكرار شبه كامل في قراءة JSONL بين `serveLogs` و`serveSpeedTestHistory` — يصلح كـ helper مشترك.
- **L8** `latencyHistory` يُضاف لها 0 وقت الانقطاع، فالرسم البياني يُظهر هبوطًا مضلِّلًا إلى 0.
- **L9** ثوابت سحرية: `histSize = 10`، `notifyCooldown = 4s`، حدود `maxHistory/maxEvents/maxTicks` — يفضّل جعلها قابلة للضبط.
- **L10** أخطاء `writeDefault`/`os.WriteFile` مُتجاهَلة في `config.go`.

---

## 3. التحسينات المقترحة

**موثوقية:**
- استخدام `http.Server{}` مع `Shutdown(ctx)` للإغلاق النظيف وإغلاق عملاء SSE.
- تدوير وأرشفة الـ logs (احتفاظ N يوم + ضغط القديم).
- إعادة المحاولة (backoff) لإرسال الـ webhooks الفاشلة بدل المحاولة لمرة واحدة.

**أمان:**
- فحص `Origin`/`Host` على كل `/api/*` + token اختياري عند التعريض على الشبكة.
- تحقق توقيع/checksum للتحديث الذاتي عبر `minisign` (موجودة بالفعل كاعتماد).
- تحقق نوع/امتداد ملف الصوت المرفوع.

**جودة الكود:**
- `slog` للـ structured logging بدل `log` القياسي.
- استخراج helper موحَّد لقراءة JSONL (L7) ولإرسال الـ webhook.
- نقل الثوابت السحرية إلى الإعدادات (L9).
- `gofmt`/`goimports` على كل الملفات + `golangci-lint`.

**أداء:**
- قياس latency حقيقي مستقل عن منطق الـ failover (L1).
- إعادة استخدام اتصالات HTTP حيث يناسب (الفحص يستخدم `DisableKeepAlives` عمدًا للدقة — مقبول).

---

## 4. الاختبارات والـ CI

- **الآن:** لا يوجد سوى بناء/إصدار في الـ CI، وصفر اختبارات.
- **مقترح:**
  - خطوة `go test ./...` + `go vet ./...` + فحص `gofmt -l`.
  - `golangci-lint` كبوابة جودة.
  - اختبارات وحدة لـ `DetermineStatus`, `compareVersions`, `config.Sanitize`, `simplifyError`, وبناة الـ webhook payloads.
  - اختبارات HTTP handlers عبر `httptest`.
  - توحيد نسخة Go (M6).

---

## 5. خريطة التطوير (Upgrade Roadmap)

**Phase 0 — استقرار (جارٍ):**
- ✅ ربط `OnConfigChange` (تطبيق حي للإعدادات).
- ✅ حماية من panic عند الفاصل الزمني الصفري + `Sanitize`.
- ✅ إظهار خطأ `ListenAndServe`.
- ⬜ graceful shutdown، توحيد نسخة Go في CI، توقيع التحديث الذاتي.

**Phase 1 — جودة وأمان:**
- اختبارات شاملة + `golangci-lint` في CI، حماية CSRF/Origin، تدوير الـ logs، structured logging، تحقق إعدادات كامل.

**Phase 2 — قدرات المراقبة:**
- قياس latency حقيقي (ICMP مع fallback لـ TCP)، قياس jitter، اختبار رفع (upload) للسرعة، سجلّ لكل هدف على حدة، التمييز بين مشاكل DNS/البوابة/مزود الخدمة.

**Phase 3 — تجربة المستخدم:**
- تقارير يومية/أسبوعية، تصدير CSV/PDF، رسوم بيانية تاريخية، Outage timeline، إشعارات قابلة للتخصيص، theme فاتح/داكن.

**Phase 4 — توسّع:**
- تخزين في SQLite بدل JSONL (استعلامات تاريخية)، `/metrics` لـ Prometheus، REST API موثّق، PWA/تطبيق مرافق، دعم عدة أجهزة/ملفات تعريف.

---

## 6. أفكار مميزات جديدة

1. **ICMP ping حقيقي** مع fallback لـ TCP للحصول على RTT أدق.
2. **اختبار سرعة الرفع (Upload)** — الحقل محجوز بالفعل في الإعدادات (`upload_target`).
3. **اختبارات سرعة مجدوَلة** (cron-like) + رسم بياني للسرعة عبر الزمن.
4. **Outage timeline** + ملخص "كم مرة/كم دقيقة انقطع النت اليوم".
5. **تقرير SLA/Uptime شهري** قابل للتصدير.
6. **قنوات تنبيه إضافية:** Telegram، البريد (SMTP)، webhook عام، ntfy.sh.
7. **اكتشاف نوع العطل** بفحص البوابة المحلية (LAN vs ISP vs DNS).
8. **`/metrics` Prometheus exporter** للدمج مع Grafana.
9. **تخزين SQLite** لاستعلامات وتحليلات تاريخية أسرع.
10. **Maintenance windows / إيقاف مؤقت** للمراقبة دون ضجيج تنبيهات.
11. **ملفات تعريف متعددة** (بيت/شغل) بإعدادات مختلفة.
12. **PWA / واجهة موبايل** اعتمادًا على الداشبورد الحالي.
13. **حماية اختيارية بـ token** للداشبورد عند تعريضه على الشبكة.
14. **توسعة i18n** (البنية موجودة في `assets/i18n.js`) لمزيد من اللغات.

---

## 7. ملخّص ما تمّ تنفيذه في هذا الـ branch

| التغيير | الملفات |
| --- | --- |
| تطبيق حي وآمن للإعدادات (`Engine.ApplyConfig` عبر قناة داخل goroutine المحرك + `Checker.SetConfig` + `Logger.SetConfig` تحت mutex) وربطه بـ `OnConfigChange` | `core/engine.go`, `monitor/checker.go`, `logger/logger.go`, `main.go`, `cmd/gui/main.go` |
| منع panic للفاصل الزمني غير الصالح + `Config.Sanitize()` لتثبيت القيم خارج المدى، وتطبيقه عند التحميل وعند الحفظ من الواجهة | `config/config.go`, `dashboard/server.go` |
| إظهار خطأ `ListenAndServe` في الـ log بدل ابتلاعه | `dashboard/server.go` |
| قراءة آمنة للـ body في `serveTestWebhook` | `dashboard/server.go` |
| اختبارات وحدة أولى (`Sanitize`, `CheckInterval`, `DetermineStatus`, `compareVersions`, `ApplyConfig`) | `config/config_test.go`, `core/engine_test.go`, `updater/updater_test.go` |

> القيود المعروفة بعد هذه الإصلاحات: تغيير `log_dir` أو `dashboard_port` من الواجهة لا يزال يتطلّب إعادة تشغيل (بنية تحتية نادرة التغيير)؛ أما الأهداف/العتبات/الفاصل الزمني/الـ webhook فتُطبَّق فورًا.
