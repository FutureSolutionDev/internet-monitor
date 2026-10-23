import { useSyncExternalStore } from "react";
import { onUnauthorized } from "./api";
import type { Snapshot, SpeedProgress } from "./types";

export type Connection = "connecting" | "live" | "reconnecting" | "unauthorized" | "busy";

export interface LiveState {
  snapshot: Snapshot | null;
  connection: Connection;
  speed: SpeedProgress | null;
  soundRevision: number;
  receivedAt: number;
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

export function createLiveStore(deps: LiveDeps) {
  let state: LiveState = {
    snapshot: null,
    connection: "connecting",
    speed: null,
    soundRevision: 0,
    receivedAt: 0,
  };
  const listeners = new Set<Listener>();
  let source: EventSourceLike | null = null;
  let retryTimer: unknown = null;
  let attempts = 0;
  let firstBoot: { id: string; version: string } | null = null;
  let stopped = true;

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
    if (!snap.speed_test_running && state.speed && !state.speed.done) {
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
    const typed = msg as { type?: string };
    if (typed.type === "speed_test_progress") {
      const progress = msg as SpeedProgress;
      const prev = state.speed;
      const running = prev && !prev.done;
      const merged: SpeedProgress = running ? { ...prev, ...progress } : progress;
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
      connect();
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

  const connect = () => {
    source?.close();
    const es = deps.open("/events");
    source = es;
    es.onopen = () => {
      attempts = 0;
      set({ connection: "live" });
    };
    es.onmessage = (ev) => {
      if (state.connection !== "live") set({ connection: "live" });
      handleMessage(ev);
    };
    es.onerror = () => {
      if (source !== es) return;
      if (es.readyState === CLOSED) {
        es.close();
        source = null;
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
      if (retryTimer !== null) deps.clearTimer(retryTimer);
      retryTimer = null;
      source?.close();
      source = null;
    },
    markUnauthorized() {
      if (state.connection === "unauthorized") return;
      source?.close();
      source = null;
      stopped = true;
      set({ connection: "unauthorized" });
    },
    beginSpeedTest() {
      set({ speed: { type: "speed_test_progress", phase: "ping", done: false } });
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
