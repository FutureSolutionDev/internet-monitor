import { describe, expect, test, vi } from "vitest";
import { createLiveStore, type EventSourceLike } from "./live";

const WATCH_EVERY = 5_000;

class FakeSource implements EventSourceLike {
  readyState = 0;
  onopen: ((ev: Event) => void) | null = null;
  onmessage: ((ev: MessageEvent) => void) | null = null;
  onerror: ((ev: Event) => void) | null = null;
  closed = false;
  close() {
    this.closed = true;
    this.readyState = 2;
  }
  open() {
    this.readyState = 1;
    this.onopen?.(new Event("open"));
  }
  send(data: unknown) {
    this.onmessage?.(new MessageEvent("message", { data: JSON.stringify(data) }));
  }
  fail(closed: boolean) {
    this.readyState = closed ? 2 : 0;
    this.onerror?.(new Event("error"));
  }
}

function setup(probeStatus = 200) {
  const sources: FakeSource[] = [];
  const timers: (() => void)[] = [];
  const watchers = new Map<number, () => void>();
  let seq = 0;
  let clock = 1000;
  const reload = vi.fn();
  const store = createLiveStore({
    open: () => {
      const s = new FakeSource();
      sources.push(s);
      return s;
    },
    probe: async () => probeStatus,
    reload,
    setTimer: (fn, ms) => {
      if (ms === WATCH_EVERY) {
        seq++;
        watchers.set(seq, fn);
        return seq;
      }
      timers.push(fn);
      return -timers.length;
    },
    clearTimer: (id) => {
      watchers.delete(id as number);
    },
    now: () => clock,
  });
  store.start();
  const advance = (ms: number) => {
    clock += ms;
    for (const [id, fn] of [...watchers]) {
      watchers.delete(id);
      fn();
    }
  };
  return { store, sources, timers, reload, advance };
}

const snap = (over: Record<string, unknown> = {}) => ({
  type: "init",
  boot_id: "a",
  version: "1.0.0",
  status: "connected",
  speed_test_running: false,
  ...over,
});

describe("live store", () => {
  test("goes live and stores snapshots", () => {
    const { store, sources } = setup();
    sources[0]?.open();
    sources[0]?.send(snap());
    expect(store.getState().connection).toBe("live");
    expect(store.getState().snapshot?.status).toBe("connected");
  });

  test("reloads the page when the app restarts with a new version", () => {
    const { sources, reload } = setup();
    sources[0]?.send(snap());
    sources[0]?.send(snap({ boot_id: "b", version: "1.0.0" }));
    expect(reload).not.toHaveBeenCalled();
    sources[0]?.send(snap({ boot_id: "c", version: "1.1.0" }));
    expect(reload).toHaveBeenCalledOnce();
  });

  test("a closed stream that probes 401 becomes unauthorized without retrying", async () => {
    const { store, sources, timers } = setup(401);
    sources[0]?.open();
    sources[0]?.fail(true);
    await vi.waitFor(() => expect(store.getState().connection).toBe("unauthorized"));
    expect(timers).toHaveLength(0);
  });

  test("a closed stream that probes 503 is busy and retries", async () => {
    const { store, sources, timers } = setup(503);
    sources[0]?.open();
    sources[0]?.fail(true);
    await vi.waitFor(() => expect(store.getState().connection).toBe("busy"));
    expect(timers).toHaveLength(1);
    timers[0]?.();
    expect(sources).toHaveLength(2);
  });

  test("a silent stream is dropped and retried", async () => {
    const { store, sources, timers, advance } = setup();
    sources[0]?.open();
    advance(30_000);
    expect(store.getState().connection).toBe("live");
    advance(15_000);
    expect(sources[0]?.closed).toBe(true);
    await vi.waitFor(() => expect(timers).toHaveLength(1));
    timers[0]?.();
    expect(sources).toHaveLength(2);
  });

  test("ping frames keep the stream alive without touching the snapshot", () => {
    const { store, sources, advance } = setup();
    sources[0]?.open();
    sources[0]?.send(snap());
    for (let i = 0; i < 4; i++) {
      advance(20_000);
      sources[0]?.send({ type: "ping" });
    }
    advance(5_000);
    expect(sources[0]?.closed).toBe(false);
    expect(store.getState().snapshot?.type).toBe("init");
  });

  test("reopening after a drop asks for fresh notifications", () => {
    const { store, sources } = setup();
    sources[0]?.open();
    expect(store.getState().notice.revision).toBe(0);
    sources[0]?.fail(false);
    sources[0]?.open();
    expect(store.getState().notice.revision).toBe(1);
  });

  test("a transient drop shows reconnecting", () => {
    const { store, sources } = setup();
    sources[0]?.open();
    sources[0]?.fail(false);
    expect(store.getState().connection).toBe("reconnecting");
  });

  test("speed progress merges and keeps the download result when upload starts", () => {
    const { store, sources } = setup();
    store.beginSpeedTest();
    sources[0]?.send({ type: "speed_test_progress", phase: "download", done: false, current_mbps: 42 });
    sources[0]?.send({ type: "speed_test_progress", phase: "ping", done: false, latency_ms: 12 });
    sources[0]?.send({ type: "speed_test_progress", phase: "upload", done: false, current_mbps: 8 });
    const s = store.getState().speed;
    expect(s?.download_mbps).toBe(42);
    expect(s?.latency_ms).toBe(12);
    expect(s?.current_mbps).toBe(8);
  });

  test("a stuck test is closed when the server says nothing is running", () => {
    const { store, sources } = setup();
    store.beginSpeedTest();
    sources[0]?.send(snap({ type: "tick", speed_test_running: false }));
    expect(store.getState().speed?.done).toBe(false);
    sources[0]?.send(snap({ type: "speed_state", speed_test_running: false }));
    expect(store.getState().speed?.done).toBe(true);
  });

  test("a new phase does not inherit the previous phase's speed", () => {
    const { store, sources } = setup();
    store.beginSpeedTest();
    sources[0]?.send({ type: "speed_test_progress", phase: "download", done: false, current_mbps: 90 });
    sources[0]?.send({ type: "speed_test_progress", phase: "upload", done: false });
    expect(store.getState().speed?.current_mbps).toBe(0);
    expect(store.getState().speed?.download_mbps).toBe(90);
  });

  test("unauthorized cancels a pending reconnect", async () => {
    const { store, sources, timers } = setup(503);
    sources[0]?.open();
    sources[0]?.fail(true);
    await vi.waitFor(() => expect(timers).toHaveLength(1));
    store.markUnauthorized();
    timers[0]?.();
    expect(sources).toHaveLength(1);
    expect(store.getState().connection).toBe("unauthorized");
  });

  test("sound frames bump the sound revision", () => {
    const { store, sources } = setup();
    sources[0]?.send(snap({ type: "sound" }));
    expect(store.getState().soundRevision).toBe(1);
  });

  test("malformed frames are ignored", () => {
    const { store, sources } = setup();
    sources[0]?.onmessage?.(new MessageEvent("message", { data: "{not json" }));
    expect(store.getState().snapshot).toBeNull();
  });
});
