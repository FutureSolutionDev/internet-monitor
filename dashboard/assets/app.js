/* ═══════════════════════════════════════════════════════════════
   Internet Monitor — Dashboard App
   ═══════════════════════════════════════════════════════════════ */

// ── i18n ──────────────────────────────────────────────────────
const LANGS = {
  ar: {
    appName:'مراقب الإنترنت', nav_dashboard:'لوحة التحكم', nav_logs:'السجلات', nav_settings:'الإعدادات',
    status_connected:'متصل', status_disconnected:'منقطع', status_degraded:'ضعيف الإشارة',
    status_checking:'جاري الفحص...', status_wait:'في انتظار أول فحص',
    status_sub_ok:'جميع الفحوصات ناجحة', latency:'زمن الاستجابة',
    uptime:'وقت التشغيل', uptime_pct:'نسبة الاتصال',
    disconnections:'انقطاعات', avg_latency:'متوسط الاستجابة', total_checks:'إجمالي الفحوصات',
    chart_title:'زمن الاستجابة — آخر 60 فحص', events_title:'سجل الأحداث',
    col_time:'الوقت', col_event:'الحدث', col_duration:'المدة', col_reason:'السبب',
    no_events:'لا توجد أحداث بعد', loss_label:'فقدان الحزم',
    ev_connected:'متصل', ev_disconnected:'انقطاع', ev_degraded:'ضعيف',
    logs_title:'عرض السجلات المخزنة', logs_select:'اختر تاريخاً...',
    logs_select_hint:'اختر تاريخاً لعرض السجلات', logs_empty:'لا توجد سجلات لهذا التاريخ',
    export_csv:'تصدير CSV', log_count:'سجل',
    grp_notif_test:'اختبار الإشعارات والصوت', notif_test_note:'يشغّل النغمة ويعرض إشعار تجريبي على سطح المكتب',
    test_notif:'اختبار الإشعار والصوت', test_notif_ok:'✅ تم الإرسال', test_notif_err:'❌ فشل',
    test_webhook:'اختبار الـ Webhook', test_webhook_ok:'✅ وصل بنجاح', test_webhook_err:'❌ فشل',
    test_webhook_no_url:'⚠️ لم يتم تعيين webhook_url',
    settings_title:'الإعدادات', settings_save:'حفظ الإعدادات',
    settings_saved:'✅ تم الحفظ بنجاح', settings_error:'❌ فشل الحفظ',
    requires_restart:'يتطلب إعادة تشغيل',
    grp_behaviour:'سلوك المراقبة', grp_targets:'العناوين المستهدفة', grp_storage:'التخزين والإشعارات',
    tag_affects_perf:'تؤثر على الأداء',
    check_interval:'فترة الفحص (ثانية)', fail_threshold:'عدد الفشل المتتالي للانقطاع',
    fail_threshold_note:'بعد كم فشل متتالي يُعتبر الاتصال منقطعاً',
    loss_threshold:'حد فقدان الحزم (%)', loss_threshold_note:'فوق هذه النسبة يُعتبر ضعيفاً',
    latency_threshold:'حد زمن الاستجابة (مللي ثانية)', latency_threshold_note:'فوق هذا الحد يُعتبر ضعيفاً',
    ping_targets_label:'عناوين TCP Ping', ping_targets_note:'صيغة: host:port — مثال: 8.8.8.8:53',
    add_target:'+ إضافة عنوان', remove:'حذف',
    http_target_label:'عنوان HTTP للاختبار', http_target_note:'يجب أن يرجع 200 أو 204',
    dns_target_label:'اسم النطاق لاختبار DNS', dns_target_note:'مثال: www.google.com',
    test:'اختبار', test_all:'⚡ اختبار جميع العناوين',
    testing:'جاري الاختبار...', test_all_ok:'✅ جميع العناوين تعمل', test_all_warn:'⚠️ بعض العناوين لا تستجيب',
    test_warn_banner:'العناوين المؤثرة على الأداء لم تُختبر — انقر "اختبار الكل" للتحقق قبل الحفظ',
    webhook_url:'عنوان الـ Webhook (اختياري)', webhook_note:'يُرسل إشعار JSON عند كل حدث — Discord / Slack / أي خدمة',
    log_dir:'مجلد السجلات', dashboard_port:'منفذ لوحة التحكم',
    q_excellent:'ممتاز', q_good:'جيد', q_fair:'متوسط', q_poor:'ضعيف', q_critical:'حرج',
    live:'مباشر', reconnecting:'جاري إعادة الاتصال...',
    reason_tcp:'فشل TCP Ping', reason_http:'فشل HTTP', reason_dns:'فشل DNS',
    reason_loss:'فقدان', reason_latency:'زمن استجابة عالٍ', reason_ok:'اتصال طبيعي',
    update_available:'إصدار جديد متاح', update_now:'تحديث الآن',
    update_downloading:'⏳ جاري التحميل...', update_applying:'⚙️ جاري التطبيق...',
    update_done:'✅ تم — سيُعاد التشغيل', update_err:'❌ فشل التحديث',
  },
  en: {
    appName:'Internet Monitor', nav_dashboard:'Dashboard', nav_logs:'Logs', nav_settings:'Settings',
    status_connected:'Connected', status_disconnected:'Disconnected', status_degraded:'Degraded',
    status_checking:'Checking...', status_wait:'Waiting for first check',
    status_sub_ok:'All checks passing', latency:'Latency',
    uptime:'Uptime', uptime_pct:'Connection %',
    disconnections:'Drops', avg_latency:'Avg Latency', total_checks:'Total Checks',
    chart_title:'Latency — Last 60 Checks', events_title:'Event Log',
    col_time:'Time', col_event:'Event', col_duration:'Duration', col_reason:'Reason',
    no_events:'No events yet', loss_label:'Packet Loss',
    ev_connected:'Connected', ev_disconnected:'Disconnected', ev_degraded:'Degraded',
    logs_title:'View Stored Logs', logs_select:'Select a date...',
    logs_select_hint:'Select a date to view logs', logs_empty:'No logs for this date',
    export_csv:'Export CSV', log_count:'record',
    grp_notif_test:'Notification & Sound Test', notif_test_note:'Plays the ringtone and shows a desktop notification',
    test_notif:'Test Notification & Sound', test_notif_ok:'✅ Sent', test_notif_err:'❌ Failed',
    test_webhook:'Test Webhook', test_webhook_ok:'✅ Delivered', test_webhook_err:'❌ Failed',
    test_webhook_no_url:'⚠️ webhook_url not set',
    settings_title:'Settings', settings_save:'Save Settings',
    settings_saved:'✅ Saved successfully', settings_error:'❌ Save failed',
    requires_restart:'Requires restart',
    grp_behaviour:'Monitoring Behaviour', grp_targets:'Check Targets', grp_storage:'Storage & Notifications',
    tag_affects_perf:'Affects performance',
    check_interval:'Check interval (seconds)', fail_threshold:'Failures before disconnect',
    fail_threshold_note:'How many consecutive failures trigger disconnect status',
    loss_threshold:'Packet loss threshold (%)', loss_threshold_note:'Above this % = degraded',
    latency_threshold:'Latency threshold (ms)', latency_threshold_note:'Above this ms = degraded',
    ping_targets_label:'TCP Ping Targets', ping_targets_note:'Format: host:port — e.g. 8.8.8.8:53',
    add_target:'+ Add Target', remove:'Remove',
    http_target_label:'HTTP Check URL', http_target_note:'Should return 200 or 204',
    dns_target_label:'DNS Resolution Domain', dns_target_note:'e.g. www.google.com',
    test:'Test', test_all:'⚡ Test All Targets',
    testing:'Testing...', test_all_ok:'✅ All targets responding', test_all_warn:'⚠️ Some targets not responding',
    test_warn_banner:'Performance-critical targets have not been tested — click "Test All" before saving',
    webhook_url:'Webhook URL (optional)', webhook_note:'Discord / Slack / any JSON-compatible service',
    log_dir:'Logs directory', dashboard_port:'Dashboard port',
    q_excellent:'Excellent', q_good:'Good', q_fair:'Fair', q_poor:'Poor', q_critical:'Critical',
    live:'Live', reconnecting:'Reconnecting...',
    reason_tcp:'TCP Ping failed', reason_http:'HTTP failed', reason_dns:'DNS failed',
    reason_loss:'Loss', reason_latency:'High latency', reason_ok:'All checks passing',
    update_available:'New version available', update_now:'Update Now',
    update_downloading:'⏳ Downloading...', update_applying:'⚙️ Applying...',
    update_done:'✅ Done — restarting', update_err:'❌ Update failed',
  }
};

let lang = localStorage.getItem('lang') || 'ar';
function t(k) { return LANGS[lang][k] || k; }

function applyLang() {
  document.documentElement.lang = lang;
  document.documentElement.dir  = lang === 'ar' ? 'rtl' : 'ltr';
  document.title = t('appName');
  document.querySelectorAll('[data-i18n]').forEach(el => el.textContent = t(el.dataset.i18n));
}

function toggleLang() {
  lang = lang === 'ar' ? 'en' : 'ar';
  localStorage.setItem('lang', lang);
  applyLang();
  if (lastData) process(lastData);
  if (logsData.length) renderLogTable(logsData); // re-translate logs tab
  renderPingTargets();
}

// Builds a translated reason string from structured event data (EventEntry or JSONL reason object)
function formatEventReason(e) {
  const parts = [];
  const tcp  = e.tcp_failed  ?? e.tcp_ping_failed  ?? false;
  const http = e.http_failed ?? false;
  const dns  = e.dns_failed  ?? false;
  const loss = e.packet_loss_pct ?? e.packet_loss ?? 0;
  const lat  = e.latency_ms ?? e.avg_latency_ms ?? 0;

  if (tcp)  parts.push(t('reason_tcp'));
  if (http) parts.push(t('reason_http'));
  if (dns)  parts.push(t('reason_dns'));

  if (!parts.length) {
    if (loss > 20) parts.push(t('reason_loss') + ' ' + loss.toFixed(0) + '%');
    else if (lat > 500) parts.push(t('reason_latency') + ' (' + lat + 'ms)');
  } else if (loss > 20) {
    parts.push(t('reason_loss') + ' ' + loss.toFixed(0) + '%');
  }

  return parts.length ? parts.join(' + ') : t('reason_ok');
}

// ── Clock ──────────────────────────────────────────────────────
setInterval(() => { document.getElementById('hdr-time').textContent = new Date().toLocaleTimeString(); }, 1000);

// ── Tabs ──────────────────────────────────────────────────────
function showTab(name) {
  document.querySelectorAll('.tab-content').forEach(el => el.classList.remove('active'));
  document.querySelectorAll('.tab-btn').forEach(el => el.classList.remove('active'));
  document.getElementById('tab-' + name).classList.add('active');
  document.querySelector('[data-tab="' + name + '"]').classList.add('active');
  if (name === 'logs')     loadLogDates();
  if (name === 'settings') loadSettings();
}

// ── Chart ──────────────────────────────────────────────────────
const chartData = {
  labels: [],
  datasets: [{
    label: 'ms', data: [],
    borderColor: '#22c55e', backgroundColor: 'rgba(34,197,94,.07)',
    borderWidth: 2, tension: .35, fill: true,
    pointRadius: 0, pointHoverRadius: 5, hitRadius: 20
  }]
};

const chart = new Chart(document.getElementById('the-chart').getContext('2d'), {
  type: 'line', data: chartData,
  options: {
    responsive: true, maintainAspectRatio: false,
    animation: { duration: 250 },
    interaction: { mode: 'index', intersect: false },
    scales: {
      x: { display: false },
      y: {
        beginAtZero: true,
        grid: { color: 'rgba(51,65,85,.5)' },
        ticks: { color: '#94a3b8', callback: v => v + 'ms' },
        border: { display: false }
      }
    },
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: '#1e293b', borderColor: '#334155', borderWidth: 1,
        titleColor: '#94a3b8', bodyColor: '#f1f5f9',
        callbacks: {
          title: items => {
            const i = items[0].dataIndex;
            const total = chartData.datasets[0].data.length;
            const secsAgo = (total - 1 - i) * 5;
            if (secsAgo === 0) return lang === 'ar' ? 'الآن' : 'Now';
            if (secsAgo < 60) return secsAgo + (lang === 'ar' ? ' ث مضت' : 's ago');
            return Math.round(secsAgo / 60) + (lang === 'ar' ? ' د مضت' : ' min ago');
          },
          label: ctx => '  ' + ctx.parsed.y + ' ms'
        }
      }
    }
  }
});

// ── Status colors + quality ────────────────────────────────────
const STATUS_C = {
  connected:    { dot:'#22c55e', border:'rgba(34,197,94,.35)',   circle:'#22c55e', icon:'✅' },
  degraded:     { dot:'#eab308', border:'rgba(234,179,8,.35)',   circle:'#eab308', icon:'⚠️' },
  disconnected: { dot:'#ef4444', border:'rgba(239,68,68,.35)',   circle:'#ef4444', icon:'❌' },
  checking:     { dot:'#94a3b8', border:'rgba(148,163,184,.2)', circle:'#94a3b8', icon:'🌐' },
};

function qualityGrade(pct, loss, lat) {
  if (pct < 50) return { grade:'F', key:'q_critical', bg:'rgba(239,68,68,.2)',  c:'#ef4444' };
  if (pct < 80) return { grade:'D', key:'q_poor',     bg:'rgba(239,68,68,.15)', c:'#ef4444' };
  if (pct < 95) return { grade:'C', key:'q_fair',     bg:'rgba(234,179,8,.15)', c:'#eab308' };
  if (loss > 5 || lat > 200) return { grade:'B', key:'q_good', bg:'rgba(234,179,8,.15)', c:'#eab308' };
  return { grade:'A', key:'q_excellent', bg:'rgba(34,197,94,.15)', c:'#22c55e' };
}

// ── Helpers ────────────────────────────────────────────────────
function fmtDur(s) {
  if (!s || s < 1) return '—';
  if (s < 60)   return s.toFixed(0) + 's';
  if (s < 3600) return Math.floor(s / 60) + 'm ' + (s % 60 | 0) + 's';
  return Math.floor(s / 3600) + 'h ' + Math.floor((s % 3600) / 60) + 'm';
}

function fmtUptime(s) {
  const h = s / 3600 | 0, m = (s % 3600 / 60) | 0, sec = s % 60 | 0;
  return h > 0 ? h + 'h ' + m + 'm' : m > 0 ? m + 'm ' + sec + 's' : sec + 's';
}

function escHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;').replace(/</g, '&lt;')
    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

// ── Dashboard: process SSE data ────────────────────────────────
let avgSum = 0, avgCnt = 0, lastData = null;

function process(d) {
  const prevStatus = lastData ? lastData.status : null;
  lastData = d;
  const st = d.status || 'checking';

  // Browser notification on status change
  if (prevStatus && prevStatus !== st && prevStatus !== 'checking') {
    _browserAlert(st, d);
  }
  const c  = STATUS_C[st] || STATUS_C.checking;

  document.getElementById('dot').style.background            = c.dot;
  document.getElementById('status-card').style.borderColor  = c.border;
  document.getElementById('status-circle').style.background = c.circle;
  document.getElementById('status-circle').textContent      = c.icon;
  document.getElementById('status-text').textContent        = t('status_' + st);
  document.getElementById('hdr-status').textContent         = t('status_' + st);

  const lat = d.latency_ms || 0;
  document.getElementById('latency-big').textContent = lat > 0 ? lat + 'ms' : '—';
  document.getElementById('status-sub').textContent  =
    d.tcp_ping_ok && d.http_ok && d.dns_ok ? t('status_sub_ok') :
    t('loss_label') + ': ' + (d.packet_loss || 0).toFixed(1) + '%';

  // Quality badge
  if (d.total_checks > 0) {
    const q = qualityGrade(d.uptime_pct || 0, d.packet_loss || 0, lat);
    const badge = document.getElementById('quality-badge');
    badge.style.display    = 'inline-block';
    badge.style.background = q.bg;
    badge.style.color      = q.c;
    badge.textContent      = q.grade + ' · ' + t(q.key);
  }

  document.getElementById('st-uptime').textContent     = fmtUptime(d.uptime_seconds || 0);
  document.getElementById('st-uptime-pct').textContent = d.total_checks > 0 ? (d.uptime_pct || 0).toFixed(1) + '%' : '—';
  document.getElementById('st-drops').textContent      = d.disconnections || 0;
  document.getElementById('st-checks').textContent     = (d.total_checks || 0).toLocaleString();

  if (lat > 0) { avgSum += lat; avgCnt++; }
  document.getElementById('st-avg').textContent = avgCnt ? (avgSum / avgCnt | 0) + 'ms' : '—';

  // Check badges
  function setChk(id, ok, label) {
    const el = document.getElementById(id);
    el.textContent = label + (ok ? ' ✓' : ' ✗');
    el.className   = 'chk-badge ' + (ok ? 'ok' : 'fail');
  }
  setChk('chk-tcp',  d.tcp_ping_ok, 'TCP');
  setChk('chk-http', d.http_ok,     'HTTP');
  setChk('chk-dns',  d.dns_ok,      'DNS');
  document.getElementById('loss-val').textContent = t('loss_label') + ': ' + (d.packet_loss || 0).toFixed(1) + '%';

  // Chart
  if (d.latency_history && d.latency_history.length) {
    chartData.labels           = d.latency_history.map((_, i) => i);
    chartData.datasets[0].data = d.latency_history;
    chartData.datasets[0].borderColor     = c.dot;
    chartData.datasets[0].backgroundColor = c.dot + '12';
    chart.update('none');
  }

  // Events table
  if (d.events && d.events.length) {
    document.getElementById('event-tbody').innerHTML = d.events.map(e => `
      <tr>
        <td class="mono">${e.time}</td>
        <td><span class="badge badge-${e.event_type}">${t('ev_' + e.event_type) || e.event_type}</span></td>
        <td class="mono">${fmtDur(e.duration_seconds)}</td>
        <td style="color:var(--muted);font-size:12px">${formatEventReason(e)}</td>
      </tr>`).join('');
  }
}

// ── SSE ────────────────────────────────────────────────────────
function connect() {
  const es = new EventSource('/events');
  es.onopen    = () => document.getElementById('hdr-status').textContent = t('live');
  es.onmessage = e  => { try { process(JSON.parse(e.data)); } catch (_) {} };
  es.onerror   = () => {
    document.getElementById('hdr-status').textContent = t('reconnecting');
    es.close();
    setTimeout(connect, 3000);
  };
}

// ══════════════════════════════════════════════════════════════
// LOGS TAB
// ══════════════════════════════════════════════════════════════
let logsData = [];

async function loadLogDates() {
  const sel = document.getElementById('log-date-select');
  if (sel.options.length > 1) return; // already populated
  try {
    const dates = await (await fetch('/api/log-dates')).json();
    dates.forEach(d => {
      const opt = document.createElement('option');
      opt.value = d; opt.textContent = d;
      sel.appendChild(opt);
    });
    if (dates.length > 0) { sel.value = dates[0]; loadLogs(); }
  } catch (e) {}
}

async function loadLogs() {
  const date = document.getElementById('log-date-select').value;
  if (!date) return;
  document.getElementById('log-tbody').innerHTML =
    `<tr><td colspan="6" class="empty">...</td></tr>`;
  try {
    logsData = await (await fetch('/api/logs?date=' + date)).json();
    renderLogTable(logsData);
  } catch (e) {
    document.getElementById('log-tbody').innerHTML =
      `<tr><td colspan="6" class="empty">${t('settings_error')}</td></tr>`;
  }
}

function renderLogTable(entries) {
  document.getElementById('log-count').textContent = entries.length + ' ' + t('log_count');
  if (!entries.length) {
    document.getElementById('log-tbody').innerHTML =
      `<tr><td colspan="6" class="empty">${t('logs_empty')}</td></tr>`;
    return;
  }
  document.getElementById('log-tbody').innerHTML = entries.map(e => {
    const ts     = new Date(e.timestamp);
    const time   = isNaN(ts) ? e.timestamp : ts.toLocaleTimeString();
    const evType = e.event || '';
    const r      = e.reason || {};
    // Pass reason fields to formatEventReason for client-side translation
    const reasonObj = {
      tcp_ping_failed: r.tcp_ping_failed || false,
      http_failed:     r.http_failed     || false,
      dns_failed:      r.dns_failed      || false,
      packet_loss_pct: r.packet_loss_pct || 0,
      avg_latency_ms:  r.avg_latency_ms  || 0,
    };
    return `<tr>
      <td class="mono">${time}</td>
      <td><span class="badge badge-${evType}">${t('ev_' + evType) || evType}</span></td>
      <td class="mono">${fmtDur(e.duration_seconds || 0)}</td>
      <td class="logs-reason">${formatEventReason(reasonObj)}</td>
      <td class="mono">${(r.packet_loss_pct || 0).toFixed(1)}%</td>
      <td class="mono">${r.avg_latency_ms > 0 ? r.avg_latency_ms + 'ms' : '—'}</td>
    </tr>`;
  }).join('');
}

function exportCSV() {
  if (!logsData.length) return;
  const rows = [['Time','Event','Duration(s)','TCP','HTTP','DNS','Loss%','Latency(ms)']];
  logsData.forEach(e => {
    const r = e.reason || {};
    rows.push([
      new Date(e.timestamp).toLocaleString(), e.event || '',
      (e.duration_seconds || 0).toFixed(1),
      r.tcp_ping_failed ? 'FAIL' : 'OK',
      r.http_failed     ? 'FAIL' : 'OK',
      r.dns_failed      ? 'FAIL' : 'OK',
      (r.packet_loss_pct || 0).toFixed(1),
      r.avg_latency_ms || 0
    ]);
  });
  const csv = rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n');
  const a = document.createElement('a');
  a.href = 'data:text/csv;charset=utf-8,﻿' + encodeURIComponent(csv);
  a.download = 'internet-monitor-' + (document.getElementById('log-date-select').value || 'logs') + '.csv';
  a.click();
}

// ══════════════════════════════════════════════════════════════
// SETTINGS TAB — target management + validation
// ══════════════════════════════════════════════════════════════
let pingTargets   = [];
let settingsTested = false; // tracks whether user tested before saving

function markUntested() {
  settingsTested = false;
  showWarnBanner(true);
}

function showWarnBanner(show) {
  const el = document.getElementById('test-warn-banner');
  if (el) el.style.display = show ? 'flex' : 'none';
}

// ── Ping targets list ──────────────────────────────────────────
function renderPingTargets() {
  const container = document.getElementById('ping-targets-container');
  if (!container) return;
  container.innerHTML = pingTargets.map((target, i) => `
    <div class="target-ping-wrap">
      <div class="target-row" id="ping-row-${i}">
        <input type="text" id="ping-${i}" value="${escHtml(target)}"
               placeholder="host:port"
               oninput="pingTargets[${i}]=this.value; markUntested()">
        <button class="btn btn-secondary btn-sm" onclick="testSingle('ping',${i})">${t('test')}</button>
        <button class="btn-remove" onclick="removePingTarget(${i})" title="${t('remove')}">×</button>
      </div>
      <span class="test-result" id="ping-result-${i}"></span>
    </div>
  `).join('');
}

function addPingTarget() {
  pingTargets.push('');
  renderPingTargets();
  const idx = pingTargets.length - 1;
  document.getElementById('ping-' + idx)?.focus();
  markUntested();
}

function removePingTarget(i) {
  pingTargets.splice(i, 1);
  renderPingTargets();
  markUntested();
}

// ── Single target test ─────────────────────────────────────────
async function testSingle(type, index) {
  const req = { ping_targets: [], http_target: '', dns_target: '' };
  let resultId = '';

  if (type === 'ping') {
    const val = (pingTargets[index] || '').trim();
    if (!val) return;
    req.ping_targets = [val];
    resultId = 'ping-result-' + index;
  } else if (type === 'http') {
    req.http_target = document.getElementById('cfg-http').value.trim();
    resultId = 'http-result';
  } else if (type === 'dns') {
    req.dns_target = document.getElementById('cfg-dns').value.trim();
    resultId = 'dns-result';
  }

  setResult(resultId, null, t('testing'));

  try {
    const res  = await fetch('/api/test-targets', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req)
    });
    const data = await res.json();

    if (type === 'ping' && data.ping_targets.length > 0) {
      showTestResult(resultId, data.ping_targets[0]);
    } else if (type === 'http') {
      showTestResult(resultId, data.http_target);
    } else if (type === 'dns') {
      showTestResult(resultId, data.dns_target);
    }
  } catch (e) {
    setResult(resultId, 'test-err', '❌ error');
  }
}

// ── Test All ───────────────────────────────────────────────────
async function testAllTargets() {
  const btn = document.getElementById('test-all-btn');
  if (btn) btn.disabled = true;

  // Sync pingTargets from inputs first
  pingTargets.forEach((_, i) => {
    const inp = document.getElementById('ping-' + i);
    if (inp) pingTargets[i] = inp.value;
  });

  const req = {
    ping_targets: pingTargets.map(t => t.trim()).filter(Boolean),
    http_target:  document.getElementById('cfg-http').value.trim(),
    dns_target:   document.getElementById('cfg-dns').value.trim(),
  };

  // Show loading
  req.ping_targets.forEach((_, i) => setResult('ping-result-' + i, null, t('testing')));
  setResult('http-result', null, t('testing'));
  setResult('dns-result',  null, t('testing'));
  setResult('test-all-result', null, t('testing'));

  try {
    const res  = await fetch('/api/test-targets', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req)
    });
    const data = await res.json();

    (data.ping_targets || []).forEach((r, i) => showTestResult('ping-result-' + i, r));
    showTestResult('http-result', data.http_target);
    showTestResult('dns-result',  data.dns_target);

    // Summary
    const allOk = [...(data.ping_targets || []), data.http_target, data.dns_target]
      .filter(Boolean).every(r => r.ok);
    setResult('test-all-result', allOk ? 'test-ok' : 'test-warn',
      allOk ? t('test_all_ok') : t('test_all_warn'));

    settingsTested = true;
    showWarnBanner(false);
  } catch (e) {
    setResult('test-all-result', 'test-err', '❌ error');
  } finally {
    if (btn) btn.disabled = false;
  }
}

function showTestResult(id, r) {
  if (!r) { setResult(id, null, ''); return; }
  if (r.ok) {
    setResult(id, 'test-ok', '✅ ' + r.latency_ms + 'ms');
  } else {
    setResult(id, 'test-warn', '⚠️ ' + (r.error || 'failed'));
  }
}

function setResult(id, cls, text) {
  const el = document.getElementById(id);
  if (!el) return;
  el.className   = 'test-result' + (cls ? ' ' + cls : '');
  el.textContent = text || '';
}

// ── Load settings from API ─────────────────────────────────────
async function loadSettings() {
  try {
    const cfg = await (await fetch('/api/config')).json();

    // Simple fields
    document.getElementById('cfg-interval').value         = cfg.check_interval_sec   || 5;
    document.getElementById('cfg-fail-threshold').value   = cfg.fail_threshold        || 3;
    document.getElementById('cfg-loss-threshold').value   = cfg.packet_loss_threshold || 20;
    document.getElementById('cfg-latency-threshold').value = cfg.latency_threshold_ms || 500;
    document.getElementById('cfg-webhook').value          = cfg.webhook_url           || '';
    document.getElementById('cfg-log-dir').value          = cfg.log_dir               || 'logs';
    document.getElementById('cfg-port').value             = cfg.dashboard_port        || 8765;
    document.getElementById('cfg-http').value             = cfg.http_target           || '';
    document.getElementById('cfg-dns').value              = cfg.dns_target            || '';

    // Ping targets list
    pingTargets = Array.isArray(cfg.ping_targets) ? [...cfg.ping_targets] : ['8.8.8.8:53'];
    renderPingTargets();

    // Clear any previous test results
    ['http-result','dns-result','test-all-result'].forEach(id => setResult(id, null, ''));
    settingsTested = false;
    showWarnBanner(false); // fresh load — no warning yet
  } catch (e) {}
}

// ── Save settings ──────────────────────────────────────────────
async function saveSettings() {
  const msg = document.getElementById('save-msg');
  msg.textContent = '';

  // Sync ping targets from inputs
  pingTargets.forEach((_, i) => {
    const inp = document.getElementById('ping-' + i);
    if (inp) pingTargets[i] = inp.value.trim();
  });

  try {
    // Fetch current config to preserve unknown fields
    const cfg = await (await fetch('/api/config')).json();

    cfg.check_interval_sec    = parseInt(document.getElementById('cfg-interval').value)         || 5;
    cfg.fail_threshold        = parseInt(document.getElementById('cfg-fail-threshold').value)   || 3;
    cfg.packet_loss_threshold = parseFloat(document.getElementById('cfg-loss-threshold').value) || 20;
    cfg.latency_threshold_ms  = parseInt(document.getElementById('cfg-latency-threshold').value) || 500;
    cfg.webhook_url           = document.getElementById('cfg-webhook').value.trim();
    cfg.log_dir               = document.getElementById('cfg-log-dir').value.trim() || 'logs';
    cfg.dashboard_port        = parseInt(document.getElementById('cfg-port').value)  || 8765;
    cfg.http_target           = document.getElementById('cfg-http').value.trim();
    cfg.dns_target            = document.getElementById('cfg-dns').value.trim();
    cfg.ping_targets          = pingTargets.filter(t => t.trim());

    const res = await fetch('/api/config', {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify(cfg)
    });

    if (res.ok) {
      msg.textContent = t('settings_saved');
      msg.className   = 'msg-ok';
    } else {
      throw new Error(await res.text());
    }
  } catch (e) {
    msg.textContent = t('settings_error') + ': ' + e.message;
    msg.className   = 'msg-err';
  }
  setTimeout(() => msg.textContent = '', 4000);
}

// ══════════════════════════════════════════════════════════════
// BROWSER NOTIFICATIONS + AUDIO
// ══════════════════════════════════════════════════════════════

// Request browser notification permission on load
(function askPermission() {
  if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission();
  }
})();

function playAlert() {
  try {
    const audio = new Audio('/assets/Ringtone.mp3');
    audio.volume = 0.85;
    audio.play().catch(() => {}); // ignore autoplay block
  } catch(_) {}
}

function showBrowserNotification(title, body) {
  if (!('Notification' in window) || Notification.permission !== 'granted') return;
  try {
    new Notification(title, { body, icon: '/assets/favicon.png', silent: true });
  } catch(_) {}
}

function _browserAlert(status, d) {
  const loss = (d.packet_loss || 0).toFixed(1);
  const lat  = d.latency_ms || 0;
  switch (status) {
    case 'disconnected':
      playAlert();
      showBrowserNotification(
        lang === 'ar' ? '🔴 النت انقطع!' : '🔴 Disconnected',
        lang === 'ar' ? `فقدان: ${loss}%` : `Loss: ${loss}%`
      );
      break;
    case 'degraded':
      playAlert();
      showBrowserNotification(
        lang === 'ar' ? '⚠️ الاتصال ضعيف' : '⚠️ Connection Degraded',
        lang === 'ar' ? `فقدان ${loss}% — زمن ${lat}ms` : `Loss ${loss}% — ${lat}ms`
      );
      break;
    case 'connected':
      showBrowserNotification(
        lang === 'ar' ? '✅ الإنترنت عاد' : '✅ Internet Restored',
        lang === 'ar' ? `زمن الاستجابة: ${lat}ms` : `Latency: ${lat}ms`
      );
      break;
  }
}

// ══════════════════════════════════════════════════════════════
// NOTIFICATION TEST
// ══════════════════════════════════════════════════════════════
// isNativeGUI: true when running inside the Go webview window (not a regular browser)
const isNativeGUI = typeof window['nativeMinimizeToTray'] !== 'undefined'
  || document.location.hostname === '127.0.0.1'
    && navigator.userAgent.includes('Chrome') && !navigator.userAgent.includes('Electron');

async function testNotification() {
  const btn = document.getElementById('test-notif-btn');
  const res = document.getElementById('test-notif-result');
  if (btn) btn.disabled = true;
  if (res) { res.className = 'test-result'; res.textContent = '...'; }

  // In native GUI, the server-side will play sound + OS toast — skip browser duplicate
  const isNative = typeof window['nativeMinimizeToTray'] === 'function';

  if (!isNative) {
    // Browser-only mode: play audio + show Web Notification
    if ('Notification' in window && Notification.permission === 'default') {
      await Notification.requestPermission();
    }
    playAlert();
    showBrowserNotification(
      lang === 'ar' ? '🔔 اختبار الإشعار' : '🔔 Test Notification',
      lang === 'ar' ? 'الصوت والإشعار يعملان ✅' : 'Sound and notification are working ✅'
    );
  }

  // Always call server API (triggers OS toast + sound in native/tray mode)
  try {
    const r = await fetch('/api/test-notification', { method: 'POST' });
    if (res) {
      res.className = r.ok ? 'test-result test-ok' : 'test-result test-warn';
      res.textContent = r.ok ? t('test_notif_ok') : '⚠️ server';
    }
  } catch (_) {
    if (res) { res.className = 'test-result test-warn'; res.textContent = '⚠️ offline'; }
  } finally {
    if (btn) btn.disabled = false;
    setTimeout(() => { if (res) res.textContent = ''; }, 4000);
  }
}

async function testWebhook() {
  const btn = document.getElementById('test-webhook-btn');
  const res = document.getElementById('webhook-test-result');
  const url = document.getElementById('cfg-webhook')?.value?.trim();
  if (btn) btn.disabled = true;
  if (res) { res.className = 'test-result'; res.textContent = '...'; }

  if (!url) {
    if (res) { res.className = 'test-result test-warn'; res.textContent = t('test_webhook_no_url'); }
    if (btn) btn.disabled = false;
    return;
  }

  try {
    const r = await fetch('/api/test-webhook', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url })
    });
    const data = await r.json();
    if (data.ok) {
      if (res) { res.className = 'test-result test-ok'; res.textContent = t('test_webhook_ok'); }
    } else {
      if (res) { res.className = 'test-result test-err'; res.textContent = t('test_webhook_err') + ': ' + (data.error||''); }
    }
  } catch (e) {
    if (res) { res.className = 'test-result test-err'; res.textContent = t('test_webhook_err'); }
  } finally {
    if (btn) btn.disabled = false;
    setTimeout(() => { if (res) res.textContent = ''; }, 6000);
  }
}

// ══════════════════════════════════════════════════════════════
// NATIVE GUI INTEGRATION
// ══════════════════════════════════════════════════════════════

// Show "Minimize to Tray" button only when running inside the native GUI window
// (the Go code binds window.nativeMinimizeToTray on the webview)
function checkNativeMode() {
  if (typeof window['nativeMinimizeToTray'] === 'function') {
    const btn = document.getElementById('tray-minimize-btn');
    if (btn) btn.style.display = 'inline-flex';
  }
}

function minimizeToTray() {
  if (typeof window['nativeMinimizeToTray'] === 'function') {
    window['nativeMinimizeToTray']();
  }
}

// ══════════════════════════════════════════════════════════════
// INIT
// ══════════════════════════════════════════════════════════════
applyLang();
connect();

// Show version in header
fetch('/api/version').then(r=>r.json()).then(d=>{
  const el = document.getElementById('app-version');
  if (el && d.version) el.textContent = d.version;
}).catch(()=>{});

// ── Auto-update ──────────────────────────────────────────────
function showUpdateBanner(info) {
  const banner = document.getElementById('update-banner');
  const verEl  = document.getElementById('update-version');
  if (!banner || !info.has_update) return;
  if (verEl) verEl.textContent = info.latest_version;
  banner.style.display = 'flex';
  // Also update i18n text in case language changed
  banner.querySelectorAll('[data-i18n]').forEach(el => el.textContent = t(el.dataset.i18n));
}

async function applyUpdate() {
  const btn    = document.getElementById('update-btn');
  const status = document.getElementById('update-status');
  if (btn) btn.disabled = true;

  try {
    if (status) { status.className = ''; status.textContent = t('update_downloading'); }

    const r = await fetch('/api/update', { method: 'POST' });
    const d = await r.json();

    if (d.ok) {
      if (status) { status.textContent = t('update_done'); }
      // App will restart itself; show countdown
      let secs = 5;
      const iv = setInterval(() => {
        if (status) status.textContent = t('update_done') + ' (' + secs + ')';
        if (--secs < 0) clearInterval(iv);
      }, 1000);
    } else {
      if (status) { status.textContent = t('update_err') + ': ' + (d.error || ''); }
      if (btn) btn.disabled = false;
    }
  } catch (e) {
    if (status) { status.textContent = t('update_err'); }
    if (btn) btn.disabled = false;
  }
}

// Check for available update on page load
fetch('/api/update').then(r=>r.json()).then(d=>{
  if (d.has_update) showUpdateBanner(d);
}).catch(()=>{});

// Also check via SSE snapshot (server pushes update info)
const _origConnect = connect;
// Patch process() to also handle update info from snapshot
const _procOrig = process;
// Note: update info is separate from SSE; we poll /api/update every 30min
setInterval(() => {
  fetch('/api/update').then(r=>r.json()).then(d=>{
    if (d.has_update) showUpdateBanner(d);
  }).catch(()=>{});
}, 30 * 60 * 1000);
// Check after a short delay so the Go binding has time to register
setTimeout(checkNativeMode, 500);
