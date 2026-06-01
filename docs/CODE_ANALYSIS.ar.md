# تحليل شامل لمشروع Internet Monitor

> مستند مرجعي: مشاكل، أخطاء، تحسينات، وخريطة تطوير (Upgrade Roadmap) + أفكار مميزات جديدة.
> **حالة التنفيذ:** أُنجز على هذا الـ branch الكثير — الإصلاحات الحرجة، 23 مشكلة إجماع من مراجعة متعددة، **Phase 0 و Phase 2 كاملتين**، بوابات CI، تقرير PDF شهري، و `/metrics`. التفاصيل في القسم 7.

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
| M1 | **سباق بيانات (data race) محتمل على الإعدادات** عند التطبيق الحي (الـ checker/logger/engine يشاركون نفس `*config.Config`). | عام | ✅ عولج ضمن حل C1: التغيير يُمرَّر عبر goroutine المحرك + `Logger.SetConfig` تحت mutex. |
| M2 | **لا يوجد graceful shutdown للسيرفر** (لا توجد إشارة `http.Server` ولا `Shutdown`)؛ عملاء SSE لا يُغلَقون بنظام عند الخروج. | `dashboard/server.go` | ✅ أُصلح: `http.Server` + `Shutdown(ctx)` + قناة shutdown لعملاء SSE (اختبار). |
| M3 | **قراءة هشّة للـ body** في `serveTestWebhook` (`Read` واحدة في buffer بحجم 2048). | `dashboard/server.go` | ✅ أُصلح: `io.ReadAll` + `MaxBytesReader`. |
| M4 | **لا يوجد تحقق على `/api/config` POST**؛ أي قيم سالبة/خارج المدى كانت تُحفظ. | `dashboard/server.go` | ✅ أُصلح عبر `config.Sanitize()` + رفض `log_dir` فيه `..`. |
| M5 | **صفر اختبارات** في المشروع كله. | — | ✅ أُضيفت اختبارات لـ config/core/updater/speedtest/notifytext/report/dashboard/types/monitor. |
| M6 | **عدم تطابق نسخة Go في الـ CI:** الـ workflow يستخدم `GO_VERSION: '1.21'` بينما `go.mod` يطلب `go 1.24.0`. | `.github/workflows/build.yml` | ✅ أُصلح: وُحِّدت إلى `1.24`. |
| M7 | **التحديث الذاتي دون تحقق توقيع/checksum.** التنزيل عبر HTTPS من GitHub لكن `selfupdate.Apply(..., Options{})` بلا verifier. | `updater/updater.go` | ✅ أُصلح: فرض HTTPS + تحقق الحجم + **تحقق `SHA256SUMS`** من الـ release. (التوقيع الكامل minisign لا يزال اختياريًا — يحتاج مفتاح من جهة الإصدار.) |

### 2.3 أمان (Security)

| # | المشكلة | ملاحظة |
| --- | --- | --- |
| S1 | **REST API على localhost بلا مصادقة ولا حماية CSRF/Origin.** صفحة ويب خبيثة قد تحاول POST إلى `/api/config` أو `/api/update` أو `/api/speed-test/start`. | ✅ أُصلح: حارس CSRF يرفض الـ Origin المختلف على كل طلبات التعديل. |
| S2 | **رفع ملف الصوت** يقبل أي محتوى حتى 10MB دون تحقق نوع MIME. | ⬜ تحقق امتداد/MIME — Phase 1. |
| S3 | الواجهة تستخدم `escHtml` لمدخلات المستخدم في الإعدادات (جيّد)، لكن أصناف الشارات `badge-${evType}` تُحقن دون escape (المصدر بيانات التطبيق فالمخاطر منخفضة). | تحسين بسيط. |

### 2.4 بسيطة وجودة الكود (Minor / Quality)

- ✅ **L1** قياس الـ latency غير دقيق (يشمل timeout الأهداف الفاشلة) — أُصلح: القياس لكل محاولة على حدة.
- ✅ **L2** `os.Chdir` يتجاهل الخطأ — أُصلح: يُسجَّل الخطأ.
- ✅ **L3** `strings.Title` مهجور — أُصلح: دالة `capitalize`.
- ✅ **L4** الكود ليس `gofmt`-clean — أُصلح: `gofmt` للريبو كله + بوابة CI.
- ⬜ **L5** الحقل `stLast` يبدو غير مستخدَم بعد كتابته — `dashboard/server.go`.
- ✅ **L6** لا يوجد تدوير للـ logs — أُصلح: حذف تلقائي للملفات الأقدم من 90 يوم.
- ⬜ **L7** تكرار قراءة JSONL بين `serveLogs` و`serveSpeedTestHistory` — لا يزال (أُضيف `readMonthlyJSONL` للتقرير فقط).
- ⬜ **L8** `latencyHistory` يُضاف لها 0 وقت الانقطاع (رسم مضلِّل).
- ⬜ **L9** ثوابت سحرية (`histSize`, `notifyCooldown`, `maxHistory/...`) — يفضّل ضبطها.
- ⬜ **L10** أخطاء `writeDefault`/`os.WriteFile` مُتجاهَلة في `config.go`.

> **إضافي من مراجعة الـ 4 reviewers** (اتفق عليها 2+): عولِجت أيضًا — captive-portal (عدم اتباع redirects)، سباق `WaitGroup` في speedtest، TOCTOU في بدء الاختبار، حقن AppleScript على macOS (escape)، uptime يحتسب degraded، حدث "connected" زائف عند الإقلاع، تكرار بُناة الإشعارات/نغمة الرنين، `parseInt`/`itoa` اليدوية، وغيرها.

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

- ✅ **تمّ:** أُضيف job `quality` في CI يشغّل `gofmt -l` + `go vet ./...` + `go test ./...` (مع مكتبات GTK لحزم الـ CGO)، و **`golangci-lint` كبوابة حاجزة** (الحزم النقية + نسخة Windows، 0 ملاحظات؛ `errcheck` مؤجَّل). كل build jobs معلّقة على نجاح `quality`. وُحِّدت نسخة Go.
- ✅ اختبارات وحدة: `Sanitize`, `CheckInterval`, `DetermineStatus`, `ApplyConfig`, `compareVersions`, `parseChecksums`, `Diagnose`, `Summarize` (التقرير), `MeasureUpload`, `jitterOf`, `probe`/per-target, `parseProcRoute`, graceful shutdown, و `/metrics`.
- ⬜ **متبقٍّ:** تغطية أوسع لـ HTTP handlers، وتفعيل `errcheck` بعد تنظيف.

---

## 5. خريطة التطوير (Upgrade Roadmap)

**Phase 0 — استقرار: ✅ مكتملة**
- ✅ ربط `OnConfigChange` (تطبيق حي للإعدادات).
- ✅ حماية من panic عند الفاصل الزمني الصفري + `Sanitize`.
- ✅ إظهار خطأ `ListenAndServe`.
- ✅ graceful shutdown، ✅ توحيد نسخة Go في CI، ✅ تحقق سلامة التحديث الذاتي (SHA256SUMS).

**Phase 1 — جودة وأمان: ◑ جزئيًا**
- ✅ بوابات CI (`golangci-lint`/vet/test/fmt)، ✅ حماية CSRF/Origin، ✅ تدوير الـ logs (90 يوم)، ✅ تحقق إعدادات.
- ⬜ structured logging (`slog`)، تفعيل `errcheck`، تحقق MIME لرفع الصوت، rate-limit للـ webhooks.

**Phase 2 — قدرات المراقبة: ✅ مكتملة**
- ✅ اختبار رفع (upload)، ✅ jitter، ✅ سجل لكل هدف (per-target، فحص متوازي)، ✅ تصنيف العطل (DNS/HTTP/down)، ✅ ICMP مع fallback لـ TCP، ✅ تمييز LAN/ISP (اكتشاف البوابة + probe بـ ICMP).
- ملاحظة: ICMP وتمييز LAN/ISP **opt-in** (`use_icmp`) ويحتاجان صلاحيات ICMP (Linux/macOS؛ مع fallback آمن).

**Phase 3 — تجربة المستخدم: ◑ بدأت**
- ✅ تقرير شهري PDF (Outage report) + سلاسل تاريخية + تصدير CSV (موجود).
- ⬜ تقارير يومية/أسبوعية، Outage timeline في الداشبورد، theme فاتح/داكن، إشعارات أكثر تخصيصًا.

**Phase 4 — توسّع: ◑ بدأت**
- ✅ `/metrics` Prometheus exporter.
- ⬜ تخزين SQLite، REST API موثّق، PWA/موبايل، أجهزة/ملفات تعريف متعددة.

---

## 6. أفكار مميزات جديدة

1. ✅ **ICMP ping حقيقي** مع fallback لـ TCP — **نُفِّذ** (opt-in).
2. ✅ **اختبار سرعة الرفع (Upload)** — **نُفِّذ**.
3. ⬜ **اختبارات سرعة مجدوَلة** (cron-like) + رسم بياني للسرعة عبر الزمن.
4. ⬜ **Outage timeline** في الداشبورد + ملخص "كم مرة انقطع النت اليوم".
5. ✅ **تقرير شهري** (Outage report PDF) قابل للطباعة — **نُفِّذ**.
6. ⬜ **قنوات تنبيه إضافية:** Telegram، البريد (SMTP)، webhook عام، ntfy.sh.
7. ✅ **اكتشاف نوع العطل** (LAN vs ISP vs DNS) — **نُفِّذ**.
8. ✅ **`/metrics` Prometheus exporter** — **نُفِّذ**.
9. ⬜ **تخزين SQLite** لاستعلامات تاريخية أسرع.
10. ⬜ **Maintenance windows / إيقاف مؤقت** للمراقبة دون ضجيج تنبيهات.
11. ⬜ **ملفات تعريف متعددة** (بيت/شغل) بإعدادات مختلفة.
12. ⬜ **PWA / واجهة موبايل** اعتمادًا على الداشبورد الحالي.
13. ⬜ **حماية اختيارية بـ token** للداشبورد عند تعريضه على الشبكة.
14. ⬜ **توسعة i18n** لمزيد من اللغات.

---

## 7. ملخّص ما تمّ تنفيذه في هذا الـ branch

**أ) إصلاحات حرجة وأساسية:**
- تطبيق حي وآمن للإعدادات (`Engine.ApplyConfig` عبر قناة في goroutine المحرك + `Checker.SetConfig` + `Logger.SetConfig` تحت mutex) مربوط بـ `OnConfigChange`.
- `Config.Sanitize()` + حماية `CheckInterval` من panic، وتطبيقها عند التحميل والحفظ.
- إظهار خطأ `ListenAndServe`، قراءة آمنة للـ body، تطبيق `log_dir` حيًّا وتنبيه إعادة التشغيل للبورت.

**ب) 23 مشكلة إجماع من مراجعة 4 reviewers:** تصحيح قياس latency، captive-portal، uptime يحتسب degraded، حذف حدث الإقلاع الزائف، إعادة كتابة speedtest (إزالة سباق `WaitGroup`، CAS، دورة إلغاء، حد أقصى للاتصالات)، تقوية المُحدِّث، حارس CSRF، panic recovery، توحيد بُناة الإشعارات/الرنين في حزمتي `notifytext`/`sound`، escape لـ osascript، `strconv` بدل اليدوي، إلخ.

**ج) Phase 0 كاملة:** graceful shutdown، توحيد نسخة Go، بوابات CI (fmt/vet/test) + `golangci-lint` حاجز، تحقق `SHA256SUMS` للتحديث.

**د) Phase 2 كاملة:** اختبار رفع، jitter، per-target (فحص متوازي)، تصنيف العطل، ICMP+fallback، تمييز LAN/ISP.

**هـ) مميزات:** تقرير انقطاع شهري PDF (مع تسجيل عيّنات دقيقة `metrics_*.jsonl` + تدوير 90 يوم)، و `/metrics` Prometheus exporter.

**و) حزم/ملفات جديدة:** `report/`, `notifytext/`, `sound/`, `dashboard/assets/report.html`, واختبارات في معظم الحزم النقية.

> **قيود معروفة:** تغيير `dashboard_port` يتطلّب إعادة تشغيل. ICMP/تمييز LAN-ISP اختياريان (`use_icmp`) ويحتاجان صلاحيات ICMP (Linux/macOS، مع fallback آمن لـ TCP) — لم يُتحقَّق منهما runtime في بيئة التطوير. حزم الـ CGO (`tray`/`cmd/gui`) و الطباعة في المتصفح تحتاج بناء/تجربة على Windows/macOS.
