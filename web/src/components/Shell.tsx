import {
  ArrowDownToLine,
  CircleAlert,
  Gauge as GaugeIcon,
  History,
  LayoutDashboard,
  type LucideIcon,
  Menu,
  Moon,
  Settings,
  Sun,
  X,
} from "lucide-react";
import { type CSSProperties, type ReactNode, useEffect, useState } from "react";
import { api } from "../lib/api";
import { formatAgo } from "../lib/format";
import { useNow } from "../lib/hooks";
import { setLangOverride, useI18n } from "../lib/i18n";
import { useLive } from "../lib/live";
import { hrefOf, navigate, type Route } from "../lib/router";
import { setTheme, useTheme } from "../lib/theme";
import { statusTone } from "../styles/categories";
import { toast } from "./Toast";
import { Banner, Button } from "./ui";

const NAV: {
  route: Route;
  key: "nav_overview" | "nav_history" | "nav_speed" | "nav_settings";
  icon: LucideIcon;
}[] = [
  { route: "overview", key: "nav_overview", icon: LayoutDashboard },
  { route: "history", key: "nav_history", icon: History },
  { route: "speed", key: "nav_speed", icon: GaugeIcon },
  { route: "settings", key: "nav_settings", icon: Settings },
];

function ConnectionPill() {
  const { t } = useI18n();
  const connection = useLive((s) => s.connection);
  const status = useLive((s) => s.snapshot?.status ?? "checking");
  const lastCheck = useLive((s) => s.snapshot?.last_check ?? 0);
  const now = useNow(5000);
  const label =
    connection === "live"
      ? t(
          status === "connected"
            ? "status_connected"
            : status === "degraded"
              ? "status_degraded"
              : status === "disconnected"
                ? "status_disconnected"
                : "status_checking",
        )
      : connection === "connecting"
        ? t("conn_connecting")
        : t("conn_reconnecting");
  const tone = connection === "live" ? statusTone(status) : "var(--accent-amber)";
  return (
    <span
      className="pill"
      role="status"
      title={
        connection === "live" && lastCheck
          ? t("last_check", { ago: formatAgo(t, lastCheck, now) })
          : undefined
      }
      style={{ "--tone": tone } as CSSProperties}
    >
      <span className={connection === "live" ? "dot glow" : "dot"} aria-hidden />
      {label}
    </span>
  );
}

function UpdateBanner() {
  const { t } = useI18n();
  const info = useLive((s) => s.snapshot?.update_info);
  const status = useLive((s) => s.snapshot?.update_status ?? "idle");
  const error = useLive((s) => s.snapshot?.update_error);
  const [busy, setBusy] = useState(false);
  if (!info?.has_update && status === "idle") return null;
  if (status === "downloading" || status === "restarting") {
    return (
      <Banner icon={ArrowDownToLine}>
        {t(status === "downloading" ? "update_downloading" : "update_restarting")}
      </Banner>
    );
  }
  const install = async () => {
    setBusy(true);
    try {
      await api.applyUpdate();
    } catch (e) {
      toast(t("update_failed", { error: e instanceof Error ? e.message : String(e) }), "error");
    } finally {
      setBusy(false);
    }
  };
  return (
    <Banner
      tone={status === "error" ? "danger" : "info"}
      icon={status === "error" ? CircleAlert : ArrowDownToLine}
      actions={
        info?.has_update ? (
          <Button variant="solid" size="sm" busy={busy} onClick={install}>
            {t("update_install")}
          </Button>
        ) : null
      }
    >
      {status === "error"
        ? t("update_failed", { error: error ?? "" })
        : t("update_available", { v: info?.latest_version ?? "" })}
    </Banner>
  );
}

export function Shell({ route, children }: { route: Route; children: ReactNode }) {
  const { t, lang } = useI18n();
  const theme = useTheme();
  const version = useLive((s) => s.snapshot?.version);
  const connection = useLive((s) => s.connection);
  const configError = useLive((s) => s.snapshot?.config_error);
  const [drawer, setDrawer] = useState(false);

  useEffect(() => {
    if (!drawer) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setDrawer(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [drawer]);

  const switchLang = async () => {
    const next = lang === "ar" ? "en" : "ar";
    setLangOverride(next);
    try {
      await api.setLanguage(next);
    } catch {
      setLangOverride(null);
      toast(t("save_failed"), "error");
    }
  };

  const current = NAV.find((n) => n.route === route) ?? NAV[0];

  return (
    <div className="shell" data-drawer={drawer ? "open" : "closed"}>
      <a className="skip-link" href="#main">
        {t("skip_to_content")}
      </a>
      <aside className="sidebar" aria-label={t("nav_main")}>
        <div className="brand">
          <img src="/favicon.png" alt="" width={30} height={30} />
          <span>{t("app_name")}</span>
        </div>
        <nav className="nav">
          {NAV.map(({ route: r, key, icon: Icon }) => (
            <a
              key={r}
              className="nav-item"
              href={hrefOf(r)}
              aria-current={r === route ? "page" : undefined}
              title={t(key)}
              onClick={(e) => {
                if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
                e.preventDefault();
                setDrawer(false);
                navigate(r);
              }}
            >
              <Icon size={19} aria-hidden />
              <span>{t(key)}</span>
            </a>
          ))}
        </nav>
        <div className="sidebar-foot">
          {version ? <span className="num">{t("version_current", { v: version })}</span> : null}
        </div>
      </aside>
      <button
        type="button"
        className="backdrop"
        aria-label={t("close_menu")}
        tabIndex={-1}
        onClick={() => setDrawer(false)}
      />
      <div className="main">
        <header className="topbar">
          <Button
            className="menu-button"
            variant="ghost"
            iconOnly
            icon={drawer ? X : Menu}
            aria-label={t(drawer ? "close_menu" : "open_menu")}
            aria-expanded={drawer}
            onClick={() => setDrawer((d) => !d)}
          />
          <div className="topbar-title">
            <strong>{current ? t(current.key) : ""}</strong>
          </div>
          <div className="topbar-actions">
            <ConnectionPill />
            <Button
              variant="edge"
              size="sm"
              onClick={switchLang}
              aria-label={t("lang_toggle_label")}
              lang={lang === "ar" ? "en" : "ar"}
            >
              {t("lang_toggle")}
            </Button>
            <Button
              variant="edge"
              size="sm"
              iconOnly
              icon={theme === "dark" ? Sun : Moon}
              aria-label={t(theme === "dark" ? "theme_light" : "theme_dark")}
              title={t(theme === "dark" ? "theme_light" : "theme_dark")}
              onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            />
          </div>
        </header>
        <main id="main" className="content" tabIndex={-1}>
          {connection === "reconnecting" || connection === "busy" ? (
            <Banner tone="warn" icon={CircleAlert}>
              {t(connection === "busy" ? "gate_busy_body" : "conn_reconnecting")} · {t("conn_stale")}
            </Banner>
          ) : null}
          {configError ? (
            <Banner tone="danger" icon={CircleAlert}>
              {t("config_error", { error: configError })}
            </Banner>
          ) : null}
          <UpdateBanner />
          {children}
        </main>
      </div>
    </div>
  );
}
