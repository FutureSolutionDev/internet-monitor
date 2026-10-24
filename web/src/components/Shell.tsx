import {
  ArrowDownToLine,
  BellRing,
  CircleAlert,
  CircleHelp,
  Gauge as GaugeIcon,
  History,
  Info,
  LayoutDashboard,
  type LucideIcon,
  Menu,
  MessagesSquare,
  Moon,
  Scale,
  Settings,
  ShieldCheck,
  Sparkles,
  Sun,
  X,
} from "lucide-react";
import {
  type CSSProperties,
  type MouseEvent,
  type ReactNode,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { api } from "../lib/api";
import { errorKey } from "../lib/errors";
import { formatAgo } from "../lib/format";
import { useNow } from "../lib/hooks";
import { type MessageKey, useI18n } from "../lib/i18n";
import { useLive } from "../lib/live";
import { hrefOf, navigate, type Route } from "../lib/router";
import { useTheme } from "../lib/theme";
import { changeThemeWithTransition } from "../lib/themeTransition";
import { statusTone } from "../styles/categories";
import { LanguageSelect } from "./LanguagePicker";
import { NotificationBell } from "./NotificationBell";
import { toast } from "./Toast";
import { Banner, Button } from "./ui";
import { WhatsNew } from "./WhatsNew";

const NAV: {
  route: Route;
  key: "nav_overview" | "nav_history" | "nav_notifications" | "nav_speed" | "nav_settings";
  icon: LucideIcon;
}[] = [
  { route: "overview", key: "nav_overview", icon: LayoutDashboard },
  { route: "history", key: "nav_history", icon: History },
  { route: "notifications", key: "nav_notifications", icon: BellRing },
  { route: "speed", key: "nav_speed", icon: GaugeIcon },
  { route: "settings", key: "nav_settings", icon: Settings },
];

const ABOUT: { route: Route; key: MessageKey; icon: LucideIcon }[] = [
  { route: "about", key: "nav_about", icon: Info },
  { route: "changelog", key: "nav_changelog", icon: Sparkles },
  { route: "faq", key: "nav_faq", icon: CircleHelp },
  { route: "contact", key: "nav_contact", icon: MessagesSquare },
  { route: "privacy", key: "nav_privacy", icon: ShieldCheck },
  { route: "terms", key: "nav_terms", icon: Scale },
];

const TITLES: Record<Route, MessageKey> = {
  overview: "nav_overview",
  history: "nav_history",
  notifications: "nav_notifications",
  speed: "nav_speed",
  settings: "nav_settings",
  report: "report_open",
  changelog: "nav_changelog",
  faq: "nav_faq",
  privacy: "nav_privacy",
  terms: "nav_terms",
  about: "nav_about",
  contact: "nav_contact",
};

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
      toast(t(errorKey(e, "update_failed_plain")), "error");
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
        ? t("update_failed_plain")
        : t("update_available", { v: info?.latest_version ?? "" })}
    </Banner>
  );
}

const PHONE = "(max-width: 767px)";

function usePhone(): boolean {
  return useSyncExternalStore(
    (fn) => {
      const mq = window.matchMedia(PHONE);
      mq.addEventListener("change", fn);
      return () => mq.removeEventListener("change", fn);
    },
    () => window.matchMedia(PHONE).matches,
    () => false,
  );
}

export function Shell({ route, children }: { route: Route; children: ReactNode }) {
  const { t } = useI18n();
  const theme = useTheme();
  const version = useLive((s) => s.snapshot?.version);
  const connection = useLive((s) => s.connection);
  const configError = useLive((s) => s.snapshot?.config_error);
  const [drawer, setDrawer] = useState(false);
  const phone = usePhone();
  const sidebar = useRef<HTMLElement>(null);
  const menuButton = useRef<HTMLButtonElement>(null);
  const mainRef = useRef<HTMLElement>(null);
  const firstRoute = useRef(true);

  useEffect(() => {
    document.title = `${t(TITLES[route])} · ${t("app_name")}`;
  }, [route, t]);

  // biome-ignore lint/correctness/useExhaustiveDependencies: move focus to the page whenever the route changes
  useEffect(() => {
    if (firstRoute.current) {
      firstRoute.current = false;
      return;
    }
    mainRef.current?.focus({ preventScroll: true });
  }, [route]);

  useEffect(() => {
    if (!phone) return;
    if (drawer) sidebar.current?.querySelector<HTMLElement>("a[href]")?.focus();
  }, [drawer, phone]);

  const closeDrawer = (restoreFocus: boolean) => {
    setDrawer(false);
    if (restoreFocus) menuButton.current?.focus();
  };

  useEffect(() => {
    if (!drawer) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setDrawer(false);
        menuButton.current?.focus();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [drawer]);

  const go = (e: MouseEvent<HTMLAnchorElement>, r: Route) => {
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
    e.preventDefault();
    setDrawer(false);
    navigate(r);
  };

  return (
    <div className="shell" data-drawer={drawer ? "open" : "closed"}>
      <a className="skip-link" href="#main">
        {t("skip_to_content")}
      </a>
      <aside ref={sidebar} className="sidebar" inert={phone && !drawer}>
        <div className="brand">
          <img src="/favicon.png" alt="" width={30} height={30} />
          <span>{t("app_name")}</span>
        </div>
        <nav className="nav" aria-label={t("nav_main")}>
          {NAV.map(({ route: r, key, icon: Icon }) => (
            <a
              key={r}
              className="nav-item"
              href={hrefOf(r)}
              aria-current={r === route ? "page" : undefined}
              title={t(key)}
              onClick={(e) => go(e, r)}
            >
              <Icon size={19} aria-hidden />
              <span>{t(key)}</span>
            </a>
          ))}
        </nav>
        <div className="sidebar-foot">
          <nav className="about-nav" aria-label={t("nav_more")}>
            {ABOUT.map(({ route: r, key, icon: Icon }) => (
              <a
                key={r}
                href={hrefOf(r)}
                aria-current={r === route ? "page" : undefined}
                title={t(key)}
                onClick={(e) => go(e, r)}
              >
                <Icon size={15} aria-hidden />
                <span>{t(key)}</span>
              </a>
            ))}
          </nav>
          {version ? (
            <a className="num version-link" href={hrefOf("changelog")} onClick={(e) => go(e, "changelog")}>
              <span className="version-long">{t("version_current", { v: version })}</span>
              <span className="version-short" aria-hidden>
                {version}
              </span>
            </a>
          ) : null}
        </div>
      </aside>
      <button
        type="button"
        className="backdrop"
        aria-label={t("close_menu")}
        tabIndex={-1}
        onClick={() => closeDrawer(true)}
      />
      <div className="main">
        <header className="topbar">
          <Button
            ref={menuButton}
            className="menu-button"
            variant="ghost"
            iconOnly
            icon={drawer ? X : Menu}
            aria-label={t(drawer ? "close_menu" : "open_menu")}
            aria-expanded={drawer}
            onClick={() => setDrawer((d) => !d)}
          />
          <div className="topbar-title">
            <strong>{t(TITLES[route])}</strong>
          </div>
          <div className="topbar-actions">
            <ConnectionPill />
            <NotificationBell />
            <LanguageSelect iconOnly />
            <Button
              variant="edge"
              size="sm"
              iconOnly
              icon={theme === "dark" ? Sun : Moon}
              aria-label={t(theme === "dark" ? "theme_light" : "theme_dark")}
              title={t(theme === "dark" ? "theme_light" : "theme_dark")}
              onClick={(e) => changeThemeWithTransition(theme === "dark" ? "light" : "dark", e)}
            />
          </div>
        </header>
        <main ref={mainRef} id="main" className="content" tabIndex={-1}>
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
        <WhatsNew />
      </div>
    </div>
  );
}
