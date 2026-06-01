/* ═══════════════════════════════════════════════════════════════
   Internet Monitor — Dashboard App
   ═══════════════════════════════════════════════════════════════ */

let lang = localStorage.getItem("lang") || "en";
function t(k) {
  return LANGS[lang][k] || k;
}

function applyLang() {
  document.documentElement.lang = lang;
  document.documentElement.dir = lang === "ar" ? "rtl" : "ltr";
  document.title = t("appName");
  document
    .querySelectorAll("[data-i18n]")
    .forEach((el) => (el.textContent = t(el.dataset.i18n)));
}

function toggleLang() {
  lang = lang === "ar" ? "en" : "ar";
  localStorage.setItem("lang", lang);
  // Persist to config so OS notifications (sent by the backend) use this
  // language too — keeps the toast/balloon language matching the UI.
  api.post("/api/language?lang=" + lang).catch(() => {});
  applyLang();
  if (lastData) process(lastData);
  if (logsData.length) renderLogTable(logsData); // re-translate logs tab
  renderPingTargets();
}

// Effective monitor thresholds, kept in sync with the backend config so the
// reason text matches the connected/degraded decision the server actually made.
let monitorThresholds = { loss: 20, lat: 500 };

// Builds a translated reason string from structured event data (EventEntry or JSONL reason object)
function formatEventReason(e) {
  const parts = [];
  const tcp = e.tcp_failed ?? e.tcp_ping_failed ?? false;
  const http = e.http_failed ?? false;
  const dns = e.dns_failed ?? false;
  const loss = e.packet_loss_pct ?? e.packet_loss ?? 0;
  const lat = e.latency_ms ?? e.avg_latency_ms ?? 0;

  if (tcp) parts.push(t("reason_tcp"));
  if (http) parts.push(t("reason_http"));
  if (dns) parts.push(t("reason_dns"));

  if (!parts.length) {
    if (loss > monitorThresholds.loss)
      parts.push(t("reason_loss") + " " + loss.toFixed(0) + "%");
    else if (lat > monitorThresholds.lat)
      parts.push(t("reason_latency") + " (" + lat + "ms)");
  } else if (loss > monitorThresholds.loss) {
    parts.push(t("reason_loss") + " " + loss.toFixed(0) + "%");
  }

  return parts.length ? parts.join(" + ") : t("reason_ok");
}

// ── Clock ──────────────────────────────────────────────────────
setInterval(() => {
  document.getElementById("hdr-time").textContent =
    new Date().toLocaleTimeString();
}, 1000);

// ── Tabs ──────────────────────────────────────────────────────
function showTab(name) {
  document
    .querySelectorAll(".tab-content")
    .forEach((el) => el.classList.remove("active"));
  document
    .querySelectorAll(".tab-btn")
    .forEach((el) => el.classList.remove("active"));
  document.getElementById("tab-" + name).classList.add("active");
  document.querySelector('[data-tab="' + name + '"]').classList.add("active");
  if (name === "logs") loadLogDates();
  if (name === "settings") {
    loadSettings();
    showSettingsTab("monitoring");
  }
  if (name === "speed") loadSpeedHistory();
}

function showSettingsTab(name) {
  document
    .querySelectorAll(".stab-content")
    .forEach((el) => el.classList.remove("active"));
  document
    .querySelectorAll(".settings-nav-btn")
    .forEach((el) => el.classList.remove("active"));
  document.getElementById("stab-" + name)?.classList.add("active");
  document.querySelector('[data-stab="' + name + '"]')?.classList.add("active");
}

// ── Chart ──────────────────────────────────────────────────────
const chartData = {
  labels: [],
  datasets: [
    {
      label: "ms",
      data: [],
      borderColor: "#22c55e",
      backgroundColor: "rgba(34,197,94,.07)",
      borderWidth: 2,
      tension: 0.35,
      fill: true,
      pointRadius: 0,
      pointHoverRadius: 5,
      hitRadius: 20,
    },
  ],
};

const chart = new Chart(document.getElementById("the-chart").getContext("2d"), {
  type: "line",
  data: chartData,
  options: {
    responsive: true,
    maintainAspectRatio: false,
    animation: { duration: 250 },
    interaction: { mode: "index", intersect: false },
    scales: {
      x: { display: false },
      y: {
        beginAtZero: true,
        grid: { color: "rgba(51,65,85,.5)" },
        ticks: { color: "#94a3b8", callback: (v) => v + "ms" },
        border: { display: false },
      },
    },
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: "#1e293b",
        borderColor: "#334155",
        borderWidth: 1,
        titleColor: "#94a3b8",
        bodyColor: "#f1f5f9",
        callbacks: {
          title: (items) => {
            const i = items[0].dataIndex;
            const total = chartData.datasets[0].data.length;
            const secsAgo = (total - 1 - i) * 5;
            if (secsAgo === 0) return lang === "ar" ? "الآن" : "Now";
            if (secsAgo < 60)
              return secsAgo + (lang === "ar" ? " ث مضت" : "s ago");
            return (
              Math.round(secsAgo / 60) + (lang === "ar" ? " د مضت" : " min ago")
            );
          },
          label: (ctx) => "  " + ctx.parsed.y + " ms",
        },
      },
    },
  },
});

// ── Status colors + quality ────────────────────────────────────
const STATUS_C = {
  connected: {
    dot: "#22c55e",
    border: "rgba(34,197,94,.35)",
    circle: "#22c55e",
    icon: "✅",
  },
  degraded: {
    dot: "#eab308",
    border: "rgba(234,179,8,.35)",
    circle: "#eab308",
    icon: "⚠️",
  },
  disconnected: {
    dot: "#ef4444",
    border: "rgba(239,68,68,.35)",
    circle: "#ef4444",
    icon: "❌",
  },
  checking: {
    dot: "#94a3b8",
    border: "rgba(148,163,184,.2)",
    circle: "#94a3b8",
    icon: "🌐",
  },
};

function qualityGrade(pct, loss, lat) {
  if (pct < 50)
    return {
      grade: "F",
      key: "q_critical",
      bg: "rgba(239,68,68,.2)",
      c: "#ef4444",
    };
  if (pct < 80)
    return {
      grade: "D",
      key: "q_poor",
      bg: "rgba(239,68,68,.15)",
      c: "#ef4444",
    };
  if (pct < 95)
    return {
      grade: "C",
      key: "q_fair",
      bg: "rgba(234,179,8,.15)",
      c: "#eab308",
    };
  if (loss > 5 || lat > 200)
    return {
      grade: "B",
      key: "q_good",
      bg: "rgba(234,179,8,.15)",
      c: "#eab308",
    };
  return {
    grade: "A",
    key: "q_excellent",
    bg: "rgba(34,197,94,.15)",
    c: "#22c55e",
  };
}

// ── API client — no-cache, unified settings ────────────────────
const api = {
  get(url) {
    return fetch(url, { cache: "no-store" });
  },
  post(url, body) {
    return fetch(url, {
      method: "POST",
      cache: "no-store",
      ...(body !== undefined && {
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    });
  },
};

// ── Helpers ────────────────────────────────────────────────────
function fmtDur(s) {
  if (!s || s < 1) return "—";
  if (s < 60) return s.toFixed(0) + "s";
  if (s < 3600) return Math.floor(s / 60) + "m " + ((s % 60) | 0) + "s";
  return Math.floor(s / 3600) + "h " + Math.floor((s % 3600) / 60) + "m";
}

function fmtUptime(s) {
  const h = (s / 3600) | 0,
    m = ((s % 3600) / 60) | 0,
    sec = (s % 60) | 0;
  return h > 0 ? h + "h " + m + "m" : m > 0 ? m + "m " + sec + "s" : sec + "s";
}

function escHtml(str) {
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

// ── Dashboard: process SSE data ────────────────────────────────
let avgSum = 0,
  avgCnt = 0,
  lastData = null,
  speedObservedRunning = false;

function process(d) {
  if (d.update_info && d.update_info.has_update)
    showUpdateBanner(d.update_info);
  if (d.type === "speed_test_progress") {
    handleSpeedProgress(d);
    return;
  }
  const prevStatus = lastData ? lastData.status : null;
  lastData = d;

  // Self-heal: if a speed test was running and the snapshot now reports it
  // finished, recover the UI in case the one-shot "done" message was dropped.
  if (d.speed_test_running === true) {
    speedObservedRunning = true;
  } else if (speedObservedRunning) {
    speedObservedRunning = false;
    const run = document.getElementById("speed-run-btn");
    if (run && run.disabled) {
      _speedReset();
      loadSpeedHistory();
    }
  }
  const st = d.status || "checking";

  // Browser notification on status change
  if (prevStatus && prevStatus !== st && prevStatus !== "checking") {
    _browserAlert(st, d);
  }
  const c = STATUS_C[st] || STATUS_C.checking;

  document.getElementById("dot").style.background = c.dot;
  document.getElementById("status-card").style.borderColor = c.border;
  document.getElementById("status-circle").style.background = c.circle;
  document.getElementById("status-circle").textContent = c.icon;
  document.getElementById("status-text").textContent = t("status_" + st);
  document.getElementById("hdr-status").textContent = t("status_" + st);

  const lat = d.latency_ms || 0;
  document.getElementById("latency-big").textContent =
    lat > 0 ? lat + "ms" : "—";
  document.getElementById("status-sub").textContent =
    d.tcp_ping_ok && d.http_ok && d.dns_ok
      ? t("status_sub_ok")
      : d.diagnosis && d.diagnosis !== "ok"
        ? t("diag_" + d.diagnosis)
        : t("loss_label") + ": " + (d.packet_loss || 0).toFixed(1) + "%";

  // Quality badge
  if (d.total_checks > 0) {
    const q = qualityGrade(d.uptime_pct || 0, d.packet_loss || 0, lat);
    const badge = document.getElementById("quality-badge");
    badge.style.display = "inline-block";
    badge.style.background = q.bg;
    badge.style.color = q.c;
    badge.textContent = q.grade + " · " + t(q.key);
  }

  document.getElementById("st-uptime").textContent = fmtUptime(
    d.uptime_seconds || 0,
  );
  document.getElementById("st-uptime-pct").textContent =
    d.total_checks > 0 ? (d.uptime_pct || 0).toFixed(1) + "%" : "—";
  document.getElementById("st-drops").textContent = d.disconnections || 0;
  document.getElementById("st-checks").textContent = (
    d.total_checks || 0
  ).toLocaleString();

  if (lat > 0) {
    avgSum += lat;
    avgCnt++;
  }
  document.getElementById("st-avg").textContent = avgCnt
    ? ((avgSum / avgCnt) | 0) + "ms"
    : "—";

  // Check badges
  function setChk(id, ok, label) {
    const el = document.getElementById(id);
    el.textContent = label + (ok ? " ✓" : " ✗");
    el.className = "chk-badge " + (ok ? "ok" : "fail");
  }
  setChk("chk-tcp", d.tcp_ping_ok, "TCP");
  setChk("chk-http", d.http_ok, "HTTP");
  setChk("chk-dns", d.dns_ok, "DNS");
  document.getElementById("loss-val").textContent =
    t("loss_label") + ": " + (d.packet_loss || 0).toFixed(1) + "%";
  const jv = document.getElementById("jitter-val");
  if (jv) jv.textContent = t("jitter_label") + ": " + (d.jitter_ms || 0) + "ms";
  renderTargets(d.targets || []);

  // Chart
  if (d.latency_history && d.latency_history.length) {
    chartData.labels = d.latency_history.map((_, i) => i);
    chartData.datasets[0].data = d.latency_history;
    chartData.datasets[0].borderColor = c.dot;
    chartData.datasets[0].backgroundColor = c.dot + "12";
    chart.update("none");
  }

  // Events table
  if (d.events && d.events.length) {
    document.getElementById("event-tbody").innerHTML = d.events
      .map(
        (e) => `
      <tr>
        <td class="mono">${e.time}</td>
        <td><span class="badge badge-${e.event_type}">${t("ev_" + e.event_type) || e.event_type}</span></td>
        <td class="mono">${fmtDur(e.duration_seconds)}</td>
        <td style="color:var(--muted);font-size:12px">${formatEventReason(e)}</td>
      </tr>`,
      )
      .join("");
  }

  // Ticks table
  if (d.ticks && d.ticks.length) {
    document.getElementById("ticks-tbody").innerHTML = d.ticks
      .map((t_) => {
        const ok = t_.tcp_ok && t_.http_ok && t_.dns_ok;
        const row = ok ? "" : ' class="tick-row-warn"';
        const chk = (v) =>
          v
            ? '<span class="tick-ok">✓</span>'
            : '<span class="tick-fail">✗</span>';
        return `<tr${row}>
        <td class="mono">${t_.time}</td>
        <td>${chk(t_.tcp_ok)}</td>
        <td>${chk(t_.http_ok)}</td>
        <td>${chk(t_.dns_ok)}</td>
        <td class="mono">${t_.latency_ms > 0 ? t_.latency_ms + "ms" : "—"}</td>
        <td class="mono">${t_.packet_loss_pct > 0 ? t_.packet_loss_pct.toFixed(1) + "%" : "—"}</td>
      </tr>`;
      })
      .join("");
  }
}

// Renders per-target check status (target strings are user-controlled -> escHtml).
function renderTargets(targets) {
  const el = document.getElementById("targets-list");
  if (!el) return;
  if (!targets.length) {
    el.innerHTML = "";
    return;
  }
  el.innerHTML = targets
    .map((tr) => {
      const mark = tr.ok ? "✓" : "✗";
      const lat = tr.ok && tr.latency_ms ? " " + tr.latency_ms + "ms" : "";
      return `<span class="chk-badge ${tr.ok ? "ok" : "fail"}" title="${tr.layer.toUpperCase()}">${tr.layer.toUpperCase()}: ${escHtml(tr.target)} ${mark}${lat}</span>`;
    })
    .join("");
}

// ── SSE ────────────────────────────────────────────────────────
function connect() {
  const es = new EventSource("/events");
  es.onopen = () =>
    (document.getElementById("hdr-status").textContent = t("live"));
  es.onmessage = (e) => {
    try {
      process(JSON.parse(e.data));
    } catch (_) {}
  };
  es.onerror = () => {
    document.getElementById("hdr-status").textContent = t("reconnecting");
    es.close();
    setTimeout(connect, 3000);
  };
}

// ══════════════════════════════════════════════════════════════
// LOGS TAB
// ══════════════════════════════════════════════════════════════
let logsData = [];

// Opens the printable monthly outage report for the selected month (or the
// current month) in a new tab; the report page offers Print / Save as PDF.
function openReport() {
  let m = document.getElementById("report-month")?.value;
  if (!m) {
    const n = new Date();
    m = n.getFullYear() + "-" + String(n.getMonth() + 1).padStart(2, "0");
  }
  window.open("/report?month=" + encodeURIComponent(m), "_blank");
}

async function loadLogDates() {
  const sel = document.getElementById("log-date-select");
  if (sel.options.length > 1) return; // already populated
  try {
    const dates = await (await api.get("/api/log-dates")).json();
    dates.forEach((d) => {
      const opt = document.createElement("option");
      opt.value = d;
      opt.textContent = d;
      sel.appendChild(opt);
    });
    if (dates.length > 0) {
      sel.value = dates[0];
      loadLogs();
    }
  } catch (e) {}
}

async function loadLogs() {
  const date = document.getElementById("log-date-select").value;
  if (!date) return;
  document.getElementById("log-tbody").innerHTML =
    `<tr><td colspan="6" class="empty">...</td></tr>`;
  try {
    logsData = await (await api.get("/api/logs?date=" + date)).json();
    renderLogTable(logsData);
  } catch (e) {
    document.getElementById("log-tbody").innerHTML =
      `<tr><td colspan="6" class="empty">${t("settings_error")}</td></tr>`;
  }
}

function renderLogTable(entries) {
  document.getElementById("log-count").textContent =
    entries.length + " " + t("log_count");
  if (!entries.length) {
    document.getElementById("log-tbody").innerHTML =
      `<tr><td colspan="6" class="empty">${t("logs_empty")}</td></tr>`;
    return;
  }
  document.getElementById("log-tbody").innerHTML = entries
    .map((e) => {
      const ts = new Date(e.timestamp);
      const time = isNaN(ts) ? e.timestamp : ts.toLocaleTimeString();
      const evType = e.event || "";
      const r = e.reason || {};
      // Pass reason fields to formatEventReason for client-side translation
      const reasonObj = {
        tcp_ping_failed: r.tcp_ping_failed || false,
        http_failed: r.http_failed || false,
        dns_failed: r.dns_failed || false,
        packet_loss_pct: r.packet_loss_pct || 0,
        avg_latency_ms: r.avg_latency_ms || 0,
      };
      return `<tr>
      <td class="mono">${time}</td>
      <td><span class="badge badge-${evType}">${t("ev_" + evType) || evType}</span></td>
      <td class="mono">${fmtDur(e.duration_seconds || 0)}</td>
      <td class="logs-reason">${formatEventReason(reasonObj)}</td>
      <td class="mono">${(r.packet_loss_pct || 0).toFixed(1)}%</td>
      <td class="mono">${r.avg_latency_ms > 0 ? r.avg_latency_ms + "ms" : "—"}</td>
    </tr>`;
    })
    .join("");
}

function buildCSV(entries) {
  const rows = [
    ["Time", "Event", "Duration(s)", "TCP", "HTTP", "DNS", "Loss%", "Latency(ms)"],
  ];
  entries.forEach((e) => {
    const r = e.reason || {};
    rows.push([
      new Date(e.timestamp).toLocaleString(),
      e.event || "",
      (e.duration_seconds || 0).toFixed(1),
      r.tcp_ping_failed ? "FAIL" : "OK",
      r.http_failed ? "FAIL" : "OK",
      r.dns_failed ? "FAIL" : "OK",
      (r.packet_loss_pct || 0).toFixed(1),
      r.avg_latency_ms || 0,
    ]);
  });
  return rows
    .map((r) => r.map((v) => '"' + String(v).replace(/"/g, '""') + '"').join(","))
    .join("\n");
}

function downloadCSV(csv, name) {
  // Blob + object URL (not a data: URI) so large multi-day exports aren't
  // truncated by browser data-URI length limits. ﻿ = UTF-8 BOM for Excel.
  const blob = new Blob(["﻿" + csv], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = name;
  a.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function exportCSV() {
  if (!logsData.length) return;
  downloadCSV(
    buildCSV(logsData),
    "internet-monitor-" +
      (document.getElementById("log-date-select").value || "logs") +
      ".csv",
  );
}

// Exports the last N days of connectivity logs as a single combined CSV.
// Day files are fetched concurrently (not one-at-a-time) and merged in date
// order; a failed day surfaces a warning instead of silently shrinking the CSV.
async function exportRangeCSV(days) {
  try {
    const dates = await (await api.get("/api/log-dates")).json();
    const pick = (dates || []).slice(0, days); // log-dates is newest-first
    if (!pick.length) return;
    const results = await Promise.all(
      pick.map((d) =>
        api
          .get("/api/logs?date=" + d)
          .then((r) => r.json())
          .catch(() => null),
      ),
    );
    if (results.some((r) => r === null)) {
      alert(t("export_partial") || "Some days failed to load; export may be incomplete.");
    }
    let all = [];
    results.forEach((rows) => {
      if (Array.isArray(rows)) all = all.concat(rows);
    });
    if (!all.length) return;
    downloadCSV(buildCSV(all), "internet-monitor-last-" + days + "d.csv");
  } catch (_) {}
}

// ══════════════════════════════════════════════════════════════
// SETTINGS TAB — target management + validation
// ══════════════════════════════════════════════════════════════
let pingTargets = [];
let httpTargets = [];
let dnsTargets = [];
let speedTargets = [];
let settingsTested = false;

function markUntested() {
  settingsTested = false;
  showWarnBanner(true);
}

function showWarnBanner(show) {
  const el = document.getElementById("test-warn-banner");
  if (el) el.style.display = show ? "flex" : "none";
}

// ── Ping targets list ──────────────────────────────────────────
function renderPingTargets() {
  const container = document.getElementById("ping-targets-container");
  if (!container) return;
  container.innerHTML = pingTargets
    .map(
      (target, i) => `
    <div class="target-ping-wrap">
      <div class="target-row" id="ping-row-${i}">
        <input type="text" id="ping-${i}" value="${escHtml(target)}"
               placeholder="host:port"
               oninput="pingTargets[${i}]=this.value; markUntested()">
        <button class="btn btn-secondary btn-sm" onclick="testSingle('ping',${i})">${t("test")}</button>
        <button class="btn-remove" onclick="removePingTarget(${i})" title="${t("remove")}">×</button>
      </div>
      <span class="test-result" id="ping-result-${i}"></span>
    </div>
  `,
    )
    .join("");
}

function addPingTarget() {
  pingTargets.push("");
  renderPingTargets();
  document.getElementById("ping-" + (pingTargets.length - 1))?.focus();
  markUntested();
}

// ── HTTP targets ──────────────────────────────────────────────
function renderHttpTargets() {
  const c = document.getElementById("http-targets-container");
  if (!c) return;
  c.innerHTML = httpTargets
    .map(
      (v, i) => `
    <div class="target-ping-wrap">
      <div class="target-row" id="http-row-${i}">
        <input type="text" id="http-${i}" value="${escHtml(v)}"
               placeholder="https://..."
               oninput="httpTargets[${i}]=this.value; markUntested()">
        <button class="btn btn-secondary btn-sm" onclick="testSingle('http',${i})">${t("test")}</button>
        <button class="btn-remove" onclick="removeHttpTarget(${i})" title="${t("remove")}">×</button>
      </div>
      <span class="test-result" id="http-result-${i}"></span>
    </div>`,
    )
    .join("");
}
function addHttpTarget() {
  httpTargets.push("");
  renderHttpTargets();
  document.getElementById("http-" + (httpTargets.length - 1))?.focus();
  markUntested();
}
function removeHttpTarget(i) {
  httpTargets.splice(i, 1);
  renderHttpTargets();
  markUntested();
}

// ── DNS targets ───────────────────────────────────────────────
function renderDnsTargets() {
  const c = document.getElementById("dns-targets-container");
  if (!c) return;
  c.innerHTML = dnsTargets
    .map(
      (v, i) => `
    <div class="target-ping-wrap">
      <div class="target-row" id="dns-row-${i}">
        <input type="text" id="dns-${i}" value="${escHtml(v)}"
               placeholder="www.google.com"
               oninput="dnsTargets[${i}]=this.value; markUntested()">
        <button class="btn btn-secondary btn-sm" onclick="testSingle('dns',${i})">${t("test")}</button>
        <button class="btn-remove" onclick="removeDnsTarget(${i})" title="${t("remove")}">×</button>
      </div>
      <span class="test-result" id="dns-result-${i}"></span>
    </div>`,
    )
    .join("");
}
function addDnsTarget() {
  dnsTargets.push("");
  renderDnsTargets();
  document.getElementById("dns-" + (dnsTargets.length - 1))?.focus();
  markUntested();
}
function removeDnsTarget(i) {
  dnsTargets.splice(i, 1);
  renderDnsTargets();
  markUntested();
}

function removePingTarget(i) {
  pingTargets.splice(i, 1);
  renderPingTargets();
  markUntested();
}

// ── Speed download targets ─────────────────────────────────────
function renderSpeedTargets() {
  const c = document.getElementById("speed-targets-container");
  if (!c) return;
  c.innerHTML = speedTargets
    .map(
      (target, i) => `
    <div class="target-row" id="speed-row-${i}">
      <input type="text" id="speed-target-${i}" value="${escHtml(target)}"
             placeholder="https://speed.cloudflare.com/__down"
             oninput="speedTargets[${i}]=this.value">
      <button class="btn-remove" onclick="removeSpeedTarget(${i})" title="${t("remove")}">×</button>
    </div>
  `,
    )
    .join("");
}

function addSpeedTarget() {
  speedTargets.push("");
  renderSpeedTargets();
  document.getElementById("speed-target-" + (speedTargets.length - 1))?.focus();
}

function removeSpeedTarget(i) {
  if (speedTargets.length > 1) {
    speedTargets.splice(i, 1);
    renderSpeedTargets();
  }
}

// ── Single target test ─────────────────────────────────────────
async function testSingle(type, index) {
  const req = { ping_targets: [], http_target: "", dns_target: "" };
  let resultId = "";

  if (type === "ping") {
    const val = (pingTargets[index] || "").trim();
    if (!val) return;
    req.ping_targets = [val];
    resultId = "ping-result-" + index;
  } else if (type === "http") {
    const val = (
      httpTargets[index] ||
      document.getElementById("http-" + index)?.value ||
      ""
    ).trim();
    if (!val) return;
    req.http_target = val;
    resultId = "http-result-" + index;
  } else if (type === "dns") {
    const val = (
      dnsTargets[index] ||
      document.getElementById("dns-" + index)?.value ||
      ""
    ).trim();
    if (!val) return;
    req.dns_target = val;
    resultId = "dns-result-" + index;
  }

  setResult(resultId, null, t("testing"));

  try {
    const res = await api.post("/api/test-targets", req);
    const data = await res.json();

    if (type === "ping" && data.ping_targets.length > 0) {
      showTestResult(resultId, data.ping_targets[0]);
    } else if (type === "http") {
      showTestResult(resultId, data.http_target);
    } else if (type === "dns") {
      showTestResult(resultId, data.dns_target);
    }
  } catch (e) {
    setResult(resultId, "test-err", "❌ error");
  }
}

// ── Test All ───────────────────────────────────────────────────
async function testAllTargets() {
  const btn = document.getElementById("test-all-btn");
  if (btn) btn.disabled = true;

  // Sync arrays from inputs
  pingTargets.forEach((_, i) => {
    const el = document.getElementById("ping-" + i);
    if (el) pingTargets[i] = el.value;
  });
  httpTargets.forEach((_, i) => {
    const el = document.getElementById("http-" + i);
    if (el) httpTargets[i] = el.value;
  });
  dnsTargets.forEach((_, i) => {
    const el = document.getElementById("dns-" + i);
    if (el) dnsTargets[i] = el.value;
  });

  // Test each HTTP and DNS target individually (so we can show per-item results)
  const pingList = pingTargets.map((v) => v.trim()).filter(Boolean);
  const httpList = httpTargets.map((v) => v.trim()).filter(Boolean);
  const dnsList = dnsTargets.map((v) => v.trim()).filter(Boolean);

  // Show loading for all
  pingList.forEach((_, i) => setResult("ping-result-" + i, null, t("testing")));
  httpList.forEach((_, i) => setResult("http-result-" + i, null, t("testing")));
  dnsList.forEach((_, i) => setResult("dns-result-" + i, null, t("testing")));
  setResult("test-all-result", null, t("testing"));

  // Build one request with all targets
  const req = {
    ping_targets: pingList,
    http_targets: httpList,
    dns_targets: dnsList,
  };

  try {
    const res = await api.post("/api/test-targets", req);
    const data = await res.json();

    (data.ping_targets || []).forEach((r, i) =>
      showTestResult("ping-result-" + i, r),
    );
    (data.http_targets || []).forEach((r, i) =>
      showTestResult("http-result-" + i, r),
    );
    (data.dns_targets || []).forEach((r, i) =>
      showTestResult("dns-result-" + i, r),
    );

    const allResults = [
      ...(data.ping_targets || []),
      ...(data.http_targets || []),
      ...(data.dns_targets || []),
    ].filter(Boolean);
    const allOk = allResults.every((r) => r.ok);
    setResult(
      "test-all-result",
      allOk ? "test-ok" : "test-warn",
      allOk ? t("test_all_ok") : t("test_all_warn"),
    );

    settingsTested = true;
    showWarnBanner(false);
  } catch (e) {
    setResult("test-all-result", "test-err", "❌ error");
  } finally {
    if (btn) btn.disabled = false;
  }
}

// Error code → human-readable text (bilingual)
const ERR_CODES = {
  ar: {
    timeout: "لا استجابة — انتهى الوقت",
    refused: "البورت مغلق أو الخادم رفض الاتصال",
    not_found: "العنوان غير موجود (DNS)",
    unreachable: "الشبكة غير متاحة",
    no_permission: "لا توجد صلاحية للوصول",
    error: "خطأ في الاتصال",
  },
  en: {
    timeout: "No response — timed out",
    refused: "Port closed or connection refused",
    not_found: "Host not found (DNS)",
    unreachable: "Network unreachable",
    no_permission: "Access denied",
    error: "Connection error",
  },
};

function translateError(code) {
  const codes = ERR_CODES[lang] || ERR_CODES.en;
  return codes[code] || code;
}

function showTestResult(id, r) {
  if (!r) {
    setResult(id, null, "");
    return;
  }
  if (r.ok) {
    setResult(id, "test-ok", "✅ " + r.latency_ms + "ms");
  } else {
    setResult(id, "test-warn", "⚠️ " + translateError(r.error));
  }
}

function setResult(id, cls, text) {
  const el = document.getElementById(id);
  if (!el) return;
  el.className = "test-result" + (cls ? " " + cls : "");
  el.textContent = text || "";
}

// ── Load settings from API ─────────────────────────────────────
async function loadSettings() {
  try {
    const cfg = await (await api.get("/api/config")).json();

    document.getElementById("cfg-interval").value = cfg.check_interval_sec || 5;
    document.getElementById("cfg-fail-threshold").value =
      cfg.fail_threshold || 3;
    document.getElementById("cfg-loss-threshold").value =
      cfg.packet_loss_threshold || 20;
    document.getElementById("cfg-latency-threshold").value =
      cfg.latency_threshold_ms || 500;
    // Use ?? so an explicit 0 threshold is preserved (|| would silently
    // replace 0 with the default and the reason text would diverge).
    monitorThresholds = {
      loss: cfg.packet_loss_threshold ?? 20,
      lat: cfg.latency_threshold_ms ?? 500,
    };
    document.getElementById("cfg-webhook").value = cfg.webhook_url || "";
    validateWebhookURL(cfg.webhook_url || "");
    const tgT = document.getElementById("cfg-tg-token");
    const tgC = document.getElementById("cfg-tg-chat");
    if (tgT) tgT.value = cfg.telegram_bot_token || "";
    if (tgC) tgC.value = cfg.telegram_chat_id || "";
    document.getElementById("cfg-log-dir").value = cfg.log_dir || "logs";
    document.getElementById("cfg-port").value = cfg.dashboard_port || 8765;
    const st = cfg.speed_test || {};
    speedTargets = [
      ...(st.download_targets || ["https://speed.cloudflare.com/__down"]),
    ];
    renderSpeedTargets();
    const parallel = document.getElementById("cfg-speed-parallel");
    if (parallel) parallel.value = st.parallel_connections || 4;
    const timeoutEl = document.getElementById("cfg-speed-timeout");
    if (timeoutEl) timeoutEl.value = st.timeout_seconds || 10;
    const alertEl = document.getElementById("cfg-speed-alert");
    if (alertEl) alertEl.value = st.alert_threshold_mbps || 0;
    const schedEl = document.getElementById("cfg-speed-schedule");
    if (schedEl) schedEl.value = st.schedule_minutes || 0;
    const ulEl = document.getElementById("cfg-speed-upload");
    if (ulEl) ulEl.value = st.upload_target ?? "https://speed.cloudflare.com/__up";

    pingTargets = Array.isArray(cfg.ping_targets)
      ? [...cfg.ping_targets]
      : ["8.8.8.8:53"];
    httpTargets = Array.isArray(cfg.http_targets)
      ? [...cfg.http_targets]
      : ["https://connectivitycheck.gstatic.com/generate_204"];
    dnsTargets = Array.isArray(cfg.dns_targets)
      ? [...cfg.dns_targets]
      : ["www.google.com"];

    renderPingTargets();
    renderHttpTargets();
    renderDnsTargets();

    setResult("test-all-result", null, "");
    settingsTested = false;
    showWarnBanner(false);
  } catch (e) {}

  loadStartup();
  loadSoundState();
}

// ── Startup toggle ─────────────────────────────────────────────
async function loadStartup() {
  try {
    const data = await (await api.get("/api/startup")).json();
    const group = document.getElementById("startup-group");
    if (!data.supported) {
      if (group) group.style.display = "none";
      return;
    }
    if (group) group.style.display = "";
    const cb = document.getElementById("cfg-startup");
    if (cb) cb.checked = data.enabled;
    setResult(
      "startup-status",
      data.enabled ? "test-ok" : null,
      data.enabled ? t("startup_on") : "",
    );
  } catch (e) {}
}

async function toggleStartup(enabled) {
  try {
    const res = await api.post("/api/startup", { enabled });
    const data = await res.json();
    if (data.ok) {
      setResult(
        "startup-status",
        data.enabled ? "test-ok" : null,
        data.enabled ? t("startup_on") : t("startup_off"),
      );
    } else {
      setResult("startup-status", "test-err", t("startup_err"));
      const cb = document.getElementById("cfg-startup");
      if (cb) cb.checked = !enabled;
    }
  } catch (e) {
    setResult("startup-status", "test-err", t("startup_err"));
    const cb = document.getElementById("cfg-startup");
    if (cb) cb.checked = !enabled;
  }
}

// ── Save settings ──────────────────────────────────────────────
async function saveSettings() {
  const allMsgs = [
    "save-msg",
    "save-msg-targets",
    "save-msg-notifs",
    "save-msg-speed",
  ]
    .map((id) => document.getElementById(id))
    .filter(Boolean);
  const msg = allMsgs.find((el) => el.offsetParent !== null) || allMsgs[0];
  if (msg) msg.textContent = "";

  // Sync all target arrays from inputs
  pingTargets.forEach((_, i) => {
    const el = document.getElementById("ping-" + i);
    if (el) pingTargets[i] = el.value.trim();
  });
  httpTargets.forEach((_, i) => {
    const el = document.getElementById("http-" + i);
    if (el) httpTargets[i] = el.value.trim();
  });
  dnsTargets.forEach((_, i) => {
    const el = document.getElementById("dns-" + i);
    if (el) dnsTargets[i] = el.value.trim();
  });

  try {
    const cfg = await (await api.get("/api/config")).json();

    cfg.check_interval_sec =
      parseInt(document.getElementById("cfg-interval").value) || 5;
    cfg.fail_threshold =
      parseInt(document.getElementById("cfg-fail-threshold").value) || 3;
    cfg.packet_loss_threshold =
      parseFloat(document.getElementById("cfg-loss-threshold").value) || 20;
    cfg.latency_threshold_ms =
      parseInt(document.getElementById("cfg-latency-threshold").value) || 500;
    cfg.webhook_url = document.getElementById("cfg-webhook").value.trim();
    cfg.telegram_bot_token = (
      document.getElementById("cfg-tg-token")?.value || ""
    ).trim();
    cfg.telegram_chat_id = (
      document.getElementById("cfg-tg-chat")?.value || ""
    ).trim();
    cfg.log_dir = document.getElementById("cfg-log-dir").value.trim() || "logs";
    cfg.dashboard_port =
      parseInt(document.getElementById("cfg-port").value) || 8765;
    cfg.ping_targets = pingTargets.filter((v) => v.trim());
    cfg.http_targets = httpTargets.filter((v) => v.trim());
    cfg.dns_targets = dnsTargets.filter((v) => v.trim());

    const stTargets = speedTargets.map((s) => s.trim()).filter(Boolean);
    cfg.speed_test = {
      download_targets: stTargets.length
        ? stTargets
        : ["https://speed.cloudflare.com/__down"],
      parallel_connections:
        parseInt(document.getElementById("cfg-speed-parallel")?.value) || 4,
      timeout_seconds:
        parseInt(document.getElementById("cfg-speed-timeout")?.value) || 10,
      upload_target: (document.getElementById("cfg-speed-upload")?.value || "").trim(),
      alert_threshold_mbps:
        parseFloat(document.getElementById("cfg-speed-alert")?.value) || 0,
      schedule_minutes:
        parseInt(document.getElementById("cfg-speed-schedule")?.value) || 0,
    };

    const res = await api.post("/api/config", cfg);

    if (res.ok) {
      let saved = t("settings_saved");
      try {
        const r = await res.json();
        if (r.restart_required) saved += " — " + t("restart_required");
      } catch (_) {}
      msg.textContent = saved;
      msg.className = "msg-ok";
    } else {
      throw new Error(await res.text());
    }
  } catch (e) {
    msg.textContent = t("settings_error") + ": " + e.message;
    msg.className = "msg-err";
  }
  setTimeout(() => (msg.textContent = ""), 4000);
}

// ══════════════════════════════════════════════════════════════
// BROWSER NOTIFICATIONS + AUDIO
// ══════════════════════════════════════════════════════════════

// Request browser notification permission on load
(function askPermission() {
  if ("Notification" in window && Notification.permission === "default") {
    Notification.requestPermission();
  }
})();

// playAlert is the SINGLE entry point for all sound in the UI (preview, test,
// live alerts). It delegates to the backend's one native player (sound.Play),
// which stops any in-progress sound and plays the latest — so there is exactly
// one playback channel and overlapping audio is impossible no matter how fast
// the user clicks. (We deliberately do NOT use an HTML <audio> element, which
// would be a second, uncoordinated channel.)
function playAlert() {
  api.post("/api/play-sound").catch(() => {});
}

async function loadSoundState() {
  try {
    const res = await fetch("/notification-sound", { method: "HEAD", cache: "no-store" });
    const custom = res.headers.get("X-Custom-Sound") === "1";
    const resetBtn = document.getElementById("sound-reset-btn");
    const fileLabel = document.getElementById("sound-filename");
    if (resetBtn) resetBtn.style.display = custom ? "" : "none";
    if (fileLabel) fileLabel.textContent = custom ? t("sound_current") + " notification.mp3" : "";
  } catch (_) {}
}

async function uploadSound(input) {
  const file = input.files[0];
  if (!file) return;
  const status = document.getElementById("sound-status");
  const form = new FormData();
  form.append("sound", file);
  try {
    const res = await fetch("/notification-sound", { method: "POST", body: form });
    const data = await res.json();
    if (data.ok) {
      if (status) status.textContent = t("sound_ok");
      const fileLabel = document.getElementById("sound-filename");
      if (fileLabel) fileLabel.textContent = t("sound_current") + " " + file.name;
      const resetBtn = document.getElementById("sound-reset-btn");
      if (resetBtn) resetBtn.style.display = "";
    } else {
      if (status) status.textContent = t("sound_err");
    }
  } catch (_) {
    if (status) status.textContent = t("sound_err");
  }
  input.value = "";
}

async function resetSound() {
  const status = document.getElementById("sound-status");
  try {
    await fetch("/notification-sound", { method: "DELETE" });
    if (status) status.textContent = t("sound_reset_ok");
    const fileLabel = document.getElementById("sound-filename");
    if (fileLabel) fileLabel.textContent = "";
    const resetBtn = document.getElementById("sound-reset-btn");
    if (resetBtn) resetBtn.style.display = "none";
  } catch (_) {
    if (status) status.textContent = t("sound_err");
  }
}

function showBrowserNotification(title, body) {
  if (!("Notification" in window) || Notification.permission !== "granted")
    return;
  try {
    new Notification(title, {
      body,
      icon: "/assets/favicon.png",
      silent: true,
    });
  } catch (_) {}
}

function _browserAlert(status, d) {
  // Skip browser audio/notification when the Go backend handles it natively.
  if (lastData?.system_notifs) return;

  const loss = (d.packet_loss || 0).toFixed(1);
  const lat = d.latency_ms || 0;
  switch (status) {
    case "disconnected":
      playAlert();
      showBrowserNotification(
        lang === "ar" ? "🔴 النت انقطع!" : "🔴 Disconnected",
        lang === "ar" ? `فقدان: ${loss}%` : `Loss: ${loss}%`,
      );
      break;
    case "degraded":
      playAlert();
      showBrowserNotification(
        lang === "ar" ? "⚠️ الاتصال ضعيف" : "⚠️ Connection Degraded",
        lang === "ar"
          ? `فقدان ${loss}% — زمن ${lat}ms`
          : `Loss ${loss}% — ${lat}ms`,
      );
      break;
    case "connected":
      showBrowserNotification(
        lang === "ar" ? "✅ الإنترنت عاد" : "✅ Internet Restored",
        lang === "ar" ? `زمن الاستجابة: ${lat}ms` : `Latency: ${lat}ms`,
      );
      break;
  }
}

// ══════════════════════════════════════════════════════════════
// NOTIFICATION TEST
// ══════════════════════════════════════════════════════════════
// isNativeGUI: true when running inside the Go webview window (not a regular browser)
const isNativeGUI =
  typeof window["nativeMinimizeToTray"] !== "undefined" ||
  (document.location.hostname === "127.0.0.1" &&
    navigator.userAgent.includes("Chrome") &&
    !navigator.userAgent.includes("Electron"));

async function testNotification() {
  const btn = document.getElementById("test-notif-btn");
  const res = document.getElementById("test-notif-result");
  if (btn) btn.disabled = true;
  if (res) {
    res.className = "test-result";
    res.textContent = "...";
  }

  // Show a browser banner only (no sound here): the server call below plays the
  // chime via the one native player, so calling playAlert() too would double it.
  if (!lastData?.system_notifs) {
    if ("Notification" in window && Notification.permission === "default") {
      await Notification.requestPermission();
    }
    showBrowserNotification(
      lang === "ar" ? "🔔 اختبار الإشعار" : "🔔 Test Notification",
      lang === "ar"
        ? "الصوت والإشعار يعملان ✅"
        : "Sound and notification are working ✅",
    );
  }

  // Always call server API: plays the chime (one native player) + OS banner.
  try {
    const r = await api.post("/api/test-notification?lang=" + lang);
    if (res) {
      res.className = r.ok ? "test-result test-ok" : "test-result test-warn";
      res.textContent = r.ok ? t("test_notif_ok") : "⚠️ server";
    }
  } catch (_) {
    if (res) {
      res.className = "test-result test-warn";
      res.textContent = "⚠️ offline";
    }
  } finally {
    if (btn) btn.disabled = false;
    setTimeout(() => {
      if (res) res.textContent = "";
    }, 4000);
  }
}

// NOTE: client-side UX hint only. The authoritative webhook classification
// lives in logger/webhook_format.go (IsDiscord/IsSlack); keep these in sync.
function isDiscordURL(url) {
  return (
    url.includes("discord.com/api/webhooks") ||
    url.includes("discordapp.com/api/webhooks")
  );
}
function isSlackURL(url) {
  return url.includes("hooks.slack.com") || url.includes("slack.com/services/");
}
function isSupportedWebhook(url) {
  return isDiscordURL(url) || isSlackURL(url);
}

function validateWebhookURL(url) {
  const warn = document.getElementById("webhook-url-warn");
  if (!warn) return;
  warn.style.display = url && !isSupportedWebhook(url) ? "block" : "none";
}

async function testWebhook() {
  const btn = document.getElementById("test-webhook-btn");
  const res = document.getElementById("webhook-test-result");
  const url = document.getElementById("cfg-webhook")?.value?.trim();
  if (btn) btn.disabled = true;
  if (res) {
    res.className = "test-result";
    res.textContent = "...";
  }

  if (!url) {
    if (res) {
      res.className = "test-result test-warn";
      res.textContent = t("test_webhook_no_url");
    }
    if (btn) btn.disabled = false;
    return;
  }

  try {
    const r = await api.post("/api/test-webhook", { url });
    const data = await r.json();
    if (data.ok) {
      if (res) {
        res.className = "test-result test-ok";
        res.textContent = t("test_webhook_ok");
      }
    } else {
      if (res) {
        res.className = "test-result test-err";
        res.textContent = t("test_webhook_err") + ": " + (data.error || "");
      }
    }
  } catch (e) {
    if (res) {
      res.className = "test-result test-err";
      res.textContent = t("test_webhook_err");
    }
  } finally {
    if (btn) btn.disabled = false;
    setTimeout(() => {
      if (res) res.textContent = "";
    }, 6000);
  }
}

// ══════════════════════════════════════════════════════════════
// NATIVE GUI INTEGRATION
// ══════════════════════════════════════════════════════════════

// Show "Minimize to Tray" button only when running inside the native GUI window
// (the Go code binds window.nativeMinimizeToTray on the webview)
function checkNativeMode() {
  if (typeof window["nativeMinimizeToTray"] === "function") {
    const btn = document.getElementById("tray-minimize-btn");
    if (btn) btn.style.display = "inline-flex";
  }
}

function minimizeToTray() {
  if (typeof window["nativeMinimizeToTray"] === "function") {
    window["nativeMinimizeToTray"]();
  }
}

// ══════════════════════════════════════════════════════════════
// INIT
// ══════════════════════════════════════════════════════════════
applyLang();
connect();

// ── Availability widget (today / 7d / 30d) ─────────────────────
function fmtPct(v) {
  return v == null ? "—" : v.toFixed(2) + "%";
}
async function loadAvailability() {
  try {
    const a = await (await api.get("/api/availability")).json();
    const set = (id, v) => {
      const el = document.getElementById(id);
      if (el) el.textContent = fmtPct(v);
    };
    set("av-today", a.today);
    set("av-week", a.week);
    set("av-month", a.month);
  } catch (_) {}
}
loadAvailability();
setInterval(loadAvailability, 60000);

// Copies a short, paste-ready outage summary to the clipboard.
async function copySummary() {
  const d = lastData || {};
  let avail = {};
  try {
    avail = await (await api.get("/api/availability")).json();
  } catch (_) {}
  const lines = [
    "Internet Monitor — summary",
    `Status: ${d.status || "—"}`,
    `Latency: ${d.latency_ms || 0}ms · Jitter: ${d.jitter_ms || 0}ms · Loss: ${(d.packet_loss || 0).toFixed(1)}%`,
    `Uptime — today: ${fmtPct(avail.today)} · 7d: ${fmtPct(avail.week)} · 30d: ${fmtPct(avail.month)}`,
    `Disconnections this session: ${d.disconnections || 0}`,
    `Generated: ${new Date().toLocaleString()}`,
  ];
  const text = lines.join("\n");
  try {
    await navigator.clipboard.writeText(text);
    alert(t("copied") || "Copied");
  } catch (_) {
    prompt(t("copy_summary") || "Copy", text);
  }
}

// Show version in header
api
  .get("/api/version")
  .then((r) => r.json())
  .then((d) => {
    const el = document.getElementById("app-version");
    if (el && d.version) el.textContent = d.version;
  })
  .catch(() => {});

// Load effective thresholds once so event-reason text matches the backend
// before the settings tab is opened.
api
  .get("/api/config")
  .then((r) => r.json())
  .then((cfg) => {
    // Use ?? so an explicit 0 threshold is preserved (|| would silently
    // replace 0 with the default and the reason text would diverge).
    monitorThresholds = {
      loss: cfg.packet_loss_threshold ?? 20,
      lat: cfg.latency_threshold_ms ?? 500,
    };
  })
  .catch(() => {});

// ── Auto-update ──────────────────────────────────────────────
function showUpdateBanner(info) {
  const banner = document.getElementById("update-banner");
  const verEl = document.getElementById("update-version");
  if (!banner || !info.has_update) return;
  if (verEl) verEl.textContent = info.latest_version;
  banner.style.display = "flex";
  // Also update i18n text in case language changed
  banner
    .querySelectorAll("[data-i18n]")
    .forEach((el) => (el.textContent = t(el.dataset.i18n)));
}

async function applyUpdate() {
  const btn = document.getElementById("update-btn");
  const status = document.getElementById("update-status");
  if (btn) btn.disabled = true;

  try {
    if (status) {
      status.className = "";
      status.textContent = t("update_downloading");
    }

    const r = await api.post("/api/update");
    const d = await r.json();

    if (d.ok) {
      if (status) {
        status.textContent = t("update_done");
      }
      // App will restart itself; show countdown
      let secs = 5;
      const iv = setInterval(() => {
        if (status) status.textContent = t("update_done") + " (" + secs + ")";
        if (--secs < 0) clearInterval(iv);
      }, 1000);
    } else {
      if (status) {
        status.textContent = t("update_err") + ": " + (d.error || "");
      }
      if (btn) btn.disabled = false;
    }
  } catch (e) {
    if (status) {
      status.textContent = t("update_err");
    }
    if (btn) btn.disabled = false;
  }
}

// Check for available update on page load
api
  .get("/api/update")
  .then((r) => r.json())
  .then((d) => {
    if (d.has_update) showUpdateBanner(d);
  })
  .catch(() => {});

// Check after a short delay so the Go binding has time to register
setTimeout(checkNativeMode, 500);

// ── Speed Test ─────────────────────────────────────────────────

// ── Circular gauge (speedtest.net style) ───────────────────────
// The arc spans 160° (from 200° to 20°, i.e. the bottom semicircle opening up).
// Mbps is mapped on a log scale so 1→1000 Mbps all read nicely.
const GAUGE_R = 80,
  GAUGE_CX = 100,
  GAUGE_CY = 120,
  GAUGE_A0 = 180, // start angle (deg) — left
  GAUGE_A1 = 0; // end angle (deg) — right

function _mbpsToFrac(mbps) {
  if (mbps <= 0) return 0;
  // log scale: 0 at 0.1 Mbps, 1 at 1000 Mbps
  const f = (Math.log10(mbps) + 1) / 4;
  return Math.max(0, Math.min(1, f));
}
function _polar(cx, cy, r, deg) {
  const a = (deg * Math.PI) / 180;
  return [cx + r * Math.cos(a), cy - r * Math.sin(a)];
}
function _arcPath(frac) {
  const a = GAUGE_A0 + (GAUGE_A1 - GAUGE_A0) * frac;
  const [x0, y0] = _polar(GAUGE_CX, GAUGE_CY, GAUGE_R, GAUGE_A0);
  const [x1, y1] = _polar(GAUGE_CX, GAUGE_CY, GAUGE_R, a);
  const large = 0;
  const sweep = 1;
  return `M${x0.toFixed(1)} ${y0.toFixed(1)} A${GAUGE_R} ${GAUGE_R} 0 ${large} ${sweep} ${x1.toFixed(1)} ${y1.toFixed(1)}`;
}
function setGauge(mbps, phase) {
  const frac = _mbpsToFrac(mbps);
  const fill = document.getElementById("gauge-fill");
  const needle = document.getElementById("gauge-needle");
  const val = document.getElementById("gauge-value");
  const ph = document.getElementById("gauge-phase");
  if (fill) fill.setAttribute("d", _arcPath(frac));
  if (needle) {
    const a = GAUGE_A0 + (GAUGE_A1 - GAUGE_A0) * frac;
    const [nx, ny] = _polar(GAUGE_CX, GAUGE_CY, GAUGE_R - 16, a);
    needle.setAttribute("x2", nx.toFixed(1));
    needle.setAttribute("y2", ny.toFixed(1));
  }
  if (val) val.textContent = (mbps || 0).toFixed(mbps >= 100 ? 0 : 1);
  if (ph && phase) ph.textContent = phase;
  const g = document.getElementById("speed-gauge");
  if (g) g.setAttribute("data-phase", phase || "");
}

function _activeCard(id) {
  ["speed-card-ping", "speed-card-dl", "speed-card-ul"].forEach((c) => {
    const el = document.getElementById(c);
    if (el) el.classList.toggle("active", c === id);
  });
}

function handleSpeedProgress(d) {
  if (d.cancelled) {
    setGauge(0, t("speed_cancelled"));
    _activeCard(null);
    _speedReset();
    return;
  }

  if (d.phase === "ping") {
    setGauge(0, t("speed_ping"));
    _activeCard("speed-card-ping");
    if (d.latency_ms != null) {
      const p = document.getElementById("speed-ping-result");
      if (p) p.textContent = d.latency_ms;
    }
    return;
  }

  if (d.done) {
    const dl = document.getElementById("speed-dl-result");
    const ul = document.getElementById("speed-ul-result");
    const p = document.getElementById("speed-ping-result");
    if (dl && d.download_mbps != null) dl.textContent = d.download_mbps.toFixed(1);
    if (p && d.latency_ms != null) p.textContent = d.latency_ms;
    if (ul && d.result && d.result.upload_mbps != null)
      ul.textContent = d.result.upload_mbps.toFixed(1);
    setGauge(d.download_mbps || 0, t("speed_done"));
    _activeCard(null);
    _speedReset();
    loadSpeedHistory();
    return;
  }

  // Live download/upload tick.
  const mbps = d.current_mbps || 0;
  if (d.phase === "upload") {
    _activeCard("speed-card-ul");
    setGauge(mbps, t("speed_uploading"));
    const ul = document.getElementById("speed-ul-result");
    if (ul) ul.textContent = mbps.toFixed(1);
  } else {
    _activeCard("speed-card-dl");
    setGauge(mbps, t("speed_downloading"));
    const dl = document.getElementById("speed-dl-result");
    if (dl) dl.textContent = mbps.toFixed(1);
  }
}

function _speedReset() {
  const run = document.getElementById("speed-run-btn");
  const cancel = document.getElementById("speed-cancel-btn");
  if (run) run.disabled = false;
  if (cancel) cancel.style.display = "none";
}

async function startSpeedTest() {
  const run = document.getElementById("speed-run-btn");
  const cancel = document.getElementById("speed-cancel-btn");

  if (run) run.disabled = true;
  if (cancel) cancel.style.display = "";
  // Reset readouts + gauge for a fresh run.
  ["speed-ping-result", "speed-dl-result", "speed-ul-result"].forEach((id) => {
    const el = document.getElementById(id);
    if (el) el.textContent = "—";
  });
  setGauge(0, t("speed_starting"));

  try {
    const r = await api.post("/api/speed-test/start");
    if (r.status === 409) {
      setGauge(0, t("speed_running"));
    }
  } catch (_) {
    _speedReset();
  }
}

async function cancelSpeedTest() {
  try {
    await api.post("/api/speed-test/cancel");
  } catch (_) {}
  _speedReset();
  const status = document.getElementById("speed-status");
  if (status) status.textContent = t("speed_cancelled");
}

async function loadSpeedHistory() {
  try {
    const rows = await (await api.get("/api/speed-test/history")).json();
    const tbody = document.getElementById("speed-history-tbody");
    if (!tbody) return;
    if (!rows || rows.length === 0) {
      tbody.innerHTML = `<tr><td colspan="5" class="empty">${t("no_events")}</td></tr>`;
      return;
    }
    tbody.innerHTML = rows
      .map(
        (r) => `<tr>
      <td class="mono">${(r.timestamp || "").slice(11, 19)}</td>
      <td class="mono">${(r.download_mbps || 0).toFixed(1)}</td>
      <td class="mono">${r.upload_mbps != null ? r.upload_mbps.toFixed(1) : "—"}</td>
      <td class="mono">${(r.duration_seconds || 0).toFixed(1)}</td>
      <td>${r.triggered_by === "user" ? t("speed_user") : t("speed_schedule")}</td>
    </tr>`,
      )
      .join("");
  } catch (_) {}
}

async function exportSpeedCSV() {
  try {
    const rows = await (
      await api.get("/api/speed-test/history?limit=1000")
    ).json();
    if (!rows || rows.length === 0) return;
    const header = "Timestamp,Download (Mbps),Duration (s),Triggered By\n";
    const csv =
      header +
      rows
        .map(
          (r) =>
            `${r.timestamp},${(r.download_mbps || 0).toFixed(1)},${(r.duration_seconds || 0).toFixed(1)},${r.triggered_by}`,
        )
        .join("\n");
    const a = document.createElement("a");
    a.href = URL.createObjectURL(new Blob([csv], { type: "text/csv" }));
    a.download = "speedtest-history.csv";
    a.click();
  } catch (_) {}
}
