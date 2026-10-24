import { useSyncExternalStore } from "react";
import { onUnauthorized } from "./api";
import type { NotificationRecord, Snapshot, SpeedProgress } from "./types";

export type Connection = "connecting" | "live" | "reconnecting" | "unauthorized" | "busy";

export interface LiveState {
  snapshot: Snapshot | null;
  connection: Connection;
  speed: SpeedProgress | null;
  soundRevision: number;
  receivedAt: number;
  notice: { revision: number; unread: number | null; last: NotificationRecord | null };
}

type Listener = () => void;

export interface EventSourceLike {
  readyState: number;
  onopen: ((ev: Event) => void) | null;
  onmessage: ((ev: MessageEvent) => void) | null;
  onerror: ((ev: Event) => void) | null;
  close(): void;
}

export interface LiveDeps {
  open: (url: string) => EventSourceLike;
  probe: () => Promise<number>;
  reload: () => void;
  setTimer: (fn: () => void, ms: number) => unknown;
  clearTimer: (id: unknown) => void;
  now: () => number;
}

const CLOSED = 2;
const SILENCE_LIMIT = 40_000;
const WATCH_EVERY = 5_000;

export function createLiveStore(deps: LiveDeps) {
  let state: LiveState = {
    snapshot: null,
    connection: "connecting",
    speed: null,
    soundRevision: 0,
    receivedAt: 0,
    notice: { revision: 0, unread: null, last: null },
  };
  const listeners = new Set<Listener>();
  let source: EventSourceLike | null = null;
  let retryTimer: unknown = null;
  let attempts = 0;
  let opened = false;
  let lastFrame = 0;
  let watchTimer: unknown = null;
  let firstBoot: { id: string; version: string } | null = null;
  let stopped = true;
  let speedPendingSince = 0;

  const set = (patch: Partial<LiveState>) => {
    state = { ...state, ...patch };
    for (const fn of listeners) fn();
  };

  const applySnapshot = (snap: Snapshot) => {
    if (snap.type === "init") {
      if (!firstBoot) {
        firstBoot = { id: snap.boot_id, version: snap.version };
      } else if (snap.boot_id !== firstBoot.id && snap.version !== firstBoot.version) {
        deps.reload();
        return;
      }
    }
    const patch: Partial<LiveState> = { snapshot: snap, receivedAt: deps.now() };
    if (snap.type === "sound") patch.soundRevision = state.soundRevision + 1;
    if (snap.type === "speed_state" || snap.speed_test_running) speedPendingSince = 0;
    const pending = speedPendingSince > 0 && deps.now() - speedPendingSince < 10_000;
    if (!snap.speed_test_running && !pending && state.speed && !state.speed.done) {
      patch.speed = { ...state.speed, done: true, phase: "cancelled", cancelled: true };
    }
    set(patch);
  };

  const handleMessage = (ev: MessageEvent) => {
    let msg: unknown;
    try {
      msg = JSON.parse(String(ev.data));
    } catch {
      return;
    }
    if (!msg || typeof msg !== "object") return;
    const typed = msg as { type?: string; unread?: number; notification?: NotificationRecord };
    if (typed.type === "ping") return;
    if (typed.type === "notification" || typed.type === "notifications_read") {
      set({
        notice: {
          revision: state.notice.revision + 1,
          unread: typeof typed.unread === "number" ? typed.unread : state.notice.unread,
          last: typed.type === "notification" ? (typed.notification ?? null) : state.notice.last,
        },
      });
      return;
    }
    if (typed.type === "speed_test_progress") {
      const progress = msg as SpeedProgress;
      const prev = state.speed;
      const running = prev && !prev.done;
      speedPendingSince = 0;
      const merged: SpeedProgress = running ? { ...prev, ...progress } : progress;
      if (running && progress.phase && progress.phase !== prev.phase) {
        merged.current_mbps = progress.current_mbps ?? 0;
      }
      if (running && prev.phase === "download" && progress.phase !== "download" && !progress.download_mbps) {
        merged.download_mbps = prev.current_mbps;
      }
      set({ speed: merged, receivedAt: deps.now() });
      return;
    }
    applySnapshot(msg as Snapshot);
  };

  const scheduleReconnect = () => {
    if (stopped || retryTimer !== null) return;
    const delay = Math.min(30_000, 1000 * 2 ** Math.min(attempts, 5));
    attempts++;
    retryTimer = deps.setTimer(() => {
      retryTimer = null;
      if (!stopped) connect();
    }, delay);
  };

  const handleClosed = async () => {
    let status = 0;
    try {
      status = await deps.probe();
    } catch {
      status = 0;
    }
    if (stopped) return;
    if (status === 401) {
      set({ connection: "unauthorized" });
      return;
    }
    set({ connection: status === 503 ? "busy" : "reconnecting" });
    scheduleReconnect();
  };

  const stopWatch = () => {
    if (watchTimer !== null) deps.clearTimer(watchTimer);
    watchTimer = null;
  };

  const watch = () => {
    stopWatch();
    watchTimer = deps.setTimer(() => {
      watchTimer = null;
      if (stopped || !source) return;
      if (deps.now() - lastFrame > SILENCE_LIMIT) {
        source.close();
        source = null;
        set({ connection: "reconnecting" });
        void handleClosed();
        return;
      }
      watch();
    }, WATCH_EVERY);
  };

  const connect = () => {
    source?.close();
    const es = deps.open("/events");
    source = es;
    lastFrame = deps.now();
    watch();
    es.onopen = () => {
      lastFrame = deps.now();
      attempts = 0;
      if (!opened) {
        opened = true;
        set({ connection: "live" });
        return;
      }
      set({ connection: "live", notice: { ...state.notice, revision: state.notice.revision + 1 } });
    };
    es.onmessage = (ev) => {
      lastFrame = deps.now();
      if (state.connection !== "live") set({ connection: "live" });
      handleMessage(ev);
    };
    es.onerror = () => {
      if (source !== es) return;
      if (es.readyState === CLOSED) {
        es.close();
        source = null;
        stopWatch();
        set({ connection: state.connection === "connecting" ? "connecting" : "reconnecting" });
        void handleClosed();
      } else if (state.connection === "live") {
        set({ connection: "reconnecting" });
      }
    };
  };

  return {
    getState: () => state,
    subscribe(fn: Listener) {
      listeners.add(fn);
      return () => listeners.delete(fn);
    },
    start() {
      if (!stopped) return;
      stopped = false;
      connect();
    },
    stop() {
      stopped = true;
      stopWatch();
      if (retryTimer !== null) deps.clearTimer(retryTimer);
      retryTimer = null;
      source?.close();
      source = null;
    },
    markUnauthorized() {
      if (state.connection === "unauthorized") return;
      stopped = true;
      stopWatch();
      if (retryTimer !== null) deps.clearTimer(retryTimer);
      retryTimer = null;
      source?.close();
      source = null;
      set({ connection: "unauthorized" });
    },
    beginSpeedTest() {
      speedPendingSince = deps.now();
      set({ speed: { type: "speed_test_progress", phase: "ping", done: false } });
    },
    abortSpeedTest() {
      speedPendingSince = 0;
      if (state.speed && !state.speed.done) set({ speed: null });
    },
  };
}

export type LiveStore = ReturnType<typeof createLiveStore>;

export const live = createLiveStore({
  open: (url) => new EventSource(url),
  probe: async () => (await fetch("/api/version", { cache: "no-store" })).status,
  reload: () => window.location.reload(),
  setTimer: (fn, ms) => window.setTimeout(fn, ms),
  clearTimer: (id) => window.clearTimeout(id as number),
  now: () => Date.now(),
});

onUnauthorized(() => live.markUnauthorized());

export function useLive<T>(select: (s: LiveState) => T): T {
  return useSyncExternalStore(live.subscribe, () => select(live.getState()));
}
