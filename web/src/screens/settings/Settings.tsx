import {
  Bell,
  CircleAlert,
  Gauge,
  type LucideIcon,
  Music,
  Radar,
  RefreshCw,
  Save,
  Settings2,
  Target,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ConfirmDialog } from "../../components/ConfirmDialog";
import { toast } from "../../components/Toast";
import { Banner, Button, Skeleton } from "../../components/ui";
import { ApiError, api } from "../../lib/api";
import { useI18n } from "../../lib/i18n";
import { useLive } from "../../lib/live";
import { setNavigationBlocker } from "../../lib/router";
import {
  buildChanges,
  type Draft,
  draftFrom,
  type FieldKey,
  type Issues,
  rebaseDraft,
  sectionOfField,
  serverFieldToDraft,
  validateDraft,
} from "../../lib/settingsForm";
import type { MessageKey } from "../../lib/strings";
import type { ConfigView } from "../../lib/types";
import { AppSection } from "./AppSection";
import { AlertsSection, MonitoringSection, SpeedSection, TargetsSection } from "./Sections";
import { SoundCard } from "./SoundCard";

interface Loaded {
  base: ConfigView;
  revision: number;
  draft: Draft;
}

const SECTIONS: { id: string; key: MessageKey; icon: LucideIcon }[] = [
  { id: "monitoring", key: "sec_monitoring", icon: Radar },
  { id: "targets", key: "sec_targets", icon: Target },
  { id: "alerts", key: "sec_alerts", icon: Bell },
  { id: "sound", key: "sec_sound", icon: Music },
  { id: "speed", key: "sec_speed", icon: Gauge },
  { id: "app", key: "sec_app", icon: Settings2 },
];

export function Settings() {
  const { t } = useI18n();
  const serverRevision = useLive((s) => s.snapshot?.config_revision ?? 0);
  const [state, setState] = useState<Loaded | null>(null);
  const [loadError, setLoadError] = useState(false);
  const [saving, setSaving] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [serverIssue, setServerIssue] = useState<Issues>({});
  const [remote, setRemote] = useState<{ config: ConfigView; revision: number } | null>(null);
  const [readOnly, setReadOnly] = useState(false);
  const [restart, setRestart] = useState(false);
  const [pendingLeave, setPendingLeave] = useState<(() => void) | null>(null);
  const [active, setActive] = useState("monitoring");
  const loading = useRef(false);
  const stateRef = useRef(state);
  stateRef.current = state;

  const adopt = useCallback((config: ConfigView, revision: number): boolean => {
    const s = stateRef.current;
    if (!s) {
      setState({ base: config, revision, draft: draftFrom(config) });
      return true;
    }
    if (revision <= s.revision) return true;
    const r = rebaseDraft(s.draft, s.base, config);
    if (r.conflicts.length) {
      setRemote({ config, revision });
      return false;
    }
    setState({ base: config, revision, draft: r.draft });
    setRemote(null);
    return true;
  }, []);

  const sync = useCallback(async () => {
    if (loading.current) return;
    loading.current = true;
    try {
      const { config, revision } = await api.config();
      adopt(config, revision);
      setLoadError(false);
    } catch {
      setLoadError(true);
    } finally {
      loading.current = false;
    }
  }, [adopt]);

  const reset = useCallback(async () => {
    try {
      const { config, revision } = await api.config();
      setState({ base: config, revision, draft: draftFrom(config) });
      setRemote(null);
      setServerIssue({});
      setSubmitted(false);
      setLoadError(false);
    } catch {
      setLoadError(true);
    }
  }, []);

  const keepMine = () => {
    const s = stateRef.current;
    if (!s || !remote) return;
    setState({
      base: remote.config,
      revision: remote.revision,
      draft: rebaseDraft(s.draft, s.base, remote.config).draft,
    });
    setRemote(null);
  };

  useEffect(() => {
    void sync();
  }, [sync]);

  const changes = useMemo(() => (state ? buildChanges(state.draft, state.base) : {}), [state]);
  const dirty = Object.keys(changes).length > 0;
  const issues = useMemo(
    () => (state ? { ...validateDraft(state.draft), ...serverIssue } : {}),
    [state, serverIssue],
  );

  useEffect(() => {
    if (!state || serverRevision <= state.revision || saving || remote) return;
    void sync();
  }, [serverRevision, state, saving, remote, sync]);

  useEffect(() => {
    if (!dirty) {
      setNavigationBlocker(null);
      return;
    }
    setNavigationBlocker((proceed) => {
      setPendingLeave(() => proceed);
      return true;
    });
    const onUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
    };
    window.addEventListener("beforeunload", onUnload);
    return () => {
      setNavigationBlocker(null);
      window.removeEventListener("beforeunload", onUnload);
    };
  }, [dirty]);

  // biome-ignore lint/correctness/useExhaustiveDependencies: observe once the sections are rendered
  useEffect(() => {
    if (typeof IntersectionObserver === "undefined") return;
    const io = new IntersectionObserver(
      (entries) => {
        const top = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)[0];
        if (top) setActive(top.target.id);
      },
      { rootMargin: "-20% 0px -60% 0px" },
    );
    for (const s of SECTIONS) {
      const el = document.getElementById(s.id);
      if (el) io.observe(el);
    }
    return () => io.disconnect();
  }, [state !== null]);

  const set = useCallback(<K extends keyof Draft>(key: K, value: Draft[K]) => {
    setState((s) => (s ? { ...s, draft: { ...s.draft, [key]: value } } : s));
    setServerIssue((prev) => {
      if (!(key in prev) && !("targets" in prev)) return prev;
      const next = { ...prev };
      delete next[key as FieldKey];
      delete next.targets;
      return next;
    });
  }, []);

  const err = useCallback(
    (key: FieldKey): string | undefined => {
      const issue = issues[key];
      if (!issue || !state) return undefined;
      const touched =
        key in serverIssue ||
        submitted ||
        key === "targets" ||
        JSON.stringify(state.draft[key as keyof Draft]) !==
          JSON.stringify(draftFrom(state.base)[key as keyof Draft]);
      return touched ? t(issue.key, issue.vars) : undefined;
    },
    [issues, serverIssue, submitted, state, t],
  );

  const save = async () => {
    if (!state || !dirty) return;
    const invalid = Object.keys(issues) as FieldKey[];
    if (invalid.length) {
      setSubmitted(true);
      toast(t("save_invalid"), "error");
      const first = invalid[0];
      if (first)
        document
          .getElementById(sectionOfField(first))
          ?.scrollIntoView({ behavior: "smooth", block: "start" });
      return;
    }
    setSaving(true);
    try {
      let r: Awaited<ReturnType<typeof api.saveConfig>>;
      try {
        r = await api.saveConfig(changes, state.revision);
      } catch (e) {
        const cfg =
          e instanceof ApiError && e.code === "conflict"
            ? (e.body.config as ConfigView | undefined)
            : undefined;
        if (!cfg) throw e;
        const rebased = rebaseDraft(state.draft, state.base, cfg);
        if (rebased.conflicts.length) throw e;
        r = await api.saveConfig(
          buildChanges(rebased.draft, cfg),
          Number(e instanceof ApiError ? e.body.revision : 0),
        );
      }
      setState({ base: r.config, revision: r.revision, draft: draftFrom(r.config) });
      setSubmitted(false);
      setServerIssue({});
      setRemote(null);
      setRestart(r.restart_required);
      toast(t("saved"), "ok");
    } catch (e) {
      if (e instanceof ApiError && e.code === "conflict") {
        const cfg = e.body.config as ConfigView | undefined;
        const rev = Number(e.body.revision);
        if (cfg) setRemote({ config: cfg, revision: rev });
      } else if (e instanceof ApiError && e.code === "read_only") {
        setReadOnly(true);
      } else if (e instanceof ApiError && e.code === "invalid") {
        const field = serverFieldToDraft(e.field);
        if (field)
          setServerIssue({
            [field]: {
              key: field === "webhook" ? "err_webhook" : field === "log_dir" ? "err_logdir" : "err_targets",
            },
          });
        toast(t("save_invalid"), "error");
      } else {
        toast(t("save_failed"), "error");
      }
    } finally {
      setSaving(false);
    }
  };

  const discard = () => {
    if (!state) return;
    setState({ ...state, draft: draftFrom(state.base) });
    setSubmitted(false);
    setServerIssue({});
  };

  if (loadError && !state) {
    return (
      <Banner
        tone="danger"
        icon={CircleAlert}
        actions={
          <Button size="sm" icon={RefreshCw} onClick={reset}>
            {t("retry")}
          </Button>
        }
      >
        {t("load_failed")}
      </Banner>
    );
  }

  if (!state) {
    return (
      <div className="stack" aria-busy="true">
        <Skeleton height={48} width={240} />
        <Skeleton height={260} />
        <Skeleton height={320} />
      </div>
    );
  }

  const badge = (id: string) =>
    (submitted || Object.keys(serverIssue).length > 0) &&
    (Object.keys(issues) as FieldKey[]).some((k) => sectionOfField(k) === id);
  const props = { draft: state.draft, set, err };

  return (
    <>
      <div className="page-head">
        <div>
          <h1>{t("settings_title")}</h1>
          <p>{t("settings_sub")}</p>
        </div>
      </div>

      {readOnly ? (
        <Banner tone="danger" icon={CircleAlert}>
          {t("read_only")}
        </Banner>
      ) : null}
      {remote ? (
        <Banner
          tone="warn"
          icon={CircleAlert}
          actions={
            <>
              <Button size="sm" variant="ghost" onClick={keepMine}>
                {t("keep_mine")}
              </Button>
              <Button size="sm" variant="solid" icon={RefreshCw} onClick={reset}>
                {t("load_latest")}
              </Button>
            </>
          }
        >
          {t("changed_elsewhere")}
        </Banner>
      ) : null}
      {restart ? (
        <Banner tone="warn" icon={CircleAlert}>
          {t("restart_needed")}
        </Banner>
      ) : null}

      <div className="settings-layout">
        <nav className="settings-nav" aria-label={t("settings_title")}>
          {SECTIONS.map(({ id, key, icon: Icon }) => (
            <a
              key={id}
              href={`#${id}`}
              aria-current={active === id ? "true" : undefined}
              onClick={(e) => {
                e.preventDefault();
                setActive(id);
                document.getElementById(id)?.scrollIntoView({ behavior: "smooth", block: "start" });
              }}
            >
              <Icon size={16} aria-hidden />
              {t(key)}
              {badge(id) ? <span className="badge" aria-hidden /> : null}
            </a>
          ))}
        </nav>
        <div className="stack">
          <MonitoringSection {...props} />
          <TargetsSection {...props} />
          <AlertsSection {...props} base={state.base} />
          <SoundCard />
          <SpeedSection {...props} />
          <AppSection {...props} />
        </div>
      </div>

      {dirty ? (
        <section className="savebar" aria-label={t("unsaved")}>
          <span style={{ flex: 1 }}>{t("unsaved")}</span>
          <Button variant="ghost" onClick={discard} disabled={saving}>
            {t("discard")}
          </Button>
          <Button variant="solid" icon={Save} busy={saving} onClick={save} disabled={readOnly}>
            {saving ? t("saving") : t("save")}
          </Button>
        </section>
      ) : null}

      <ConfirmDialog
        open={pendingLeave !== null}
        title={t("leave_title")}
        body={t("leave_body")}
        confirmLabel={t("leave_confirm")}
        cancelLabel={t("stay")}
        danger
        onCancel={() => setPendingLeave(null)}
        onConfirm={() => {
          const go = pendingLeave;
          setPendingLeave(null);
          setNavigationBlocker(null);
          go?.();
        }}
      />
    </>
  );
}
