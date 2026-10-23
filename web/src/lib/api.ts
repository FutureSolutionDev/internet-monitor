import type {
  Availability,
  ConfigSaveResponse,
  ConfigView,
  LogEvent,
  MonthlyReport,
  SpeedTestEvent,
  StartupState,
  TestTargetsResponse,
  UpdateInfo,
} from "./types";

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly field?: string;
  readonly body: Record<string, unknown>;

  constructor(status: number, body: Record<string, unknown>) {
    super(typeof body.error === "string" ? body.error : `HTTP ${status}`);
    this.status = status;
    this.code = typeof body.code === "string" ? body.code : "http_error";
    this.field = typeof body.field === "string" ? body.field : undefined;
    this.body = body;
  }
}

type UnauthorizedListener = () => void;
const unauthorizedListeners = new Set<UnauthorizedListener>();

export function onUnauthorized(fn: UnauthorizedListener): () => void {
  unauthorizedListeners.add(fn);
  return () => unauthorizedListeners.delete(fn);
}

interface RequestOptions {
  method?: string;
  json?: unknown;
  form?: FormData;
  headers?: Record<string, string>;
  signal?: AbortSignal;
}

async function parseBody(res: Response): Promise<Record<string, unknown>> {
  const text = await res.text();
  if (!text) return {};
  try {
    const parsed: unknown = JSON.parse(text);
    return parsed && typeof parsed === "object" ? (parsed as Record<string, unknown>) : { value: parsed };
  } catch {
    return { error: text.trim() };
  }
}

export async function request<T>(
  path: string,
  opts: RequestOptions = {},
): Promise<{ data: T; res: Response }> {
  const headers: Record<string, string> = { Accept: "application/json", ...opts.headers };
  let body: BodyInit | undefined;
  if (opts.json !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(opts.json);
  } else if (opts.form) {
    body = opts.form;
  }
  const res = await fetch(path, {
    method: opts.method ?? (body ? "POST" : "GET"),
    headers,
    body,
    credentials: "same-origin",
    cache: "no-store",
    signal: opts.signal,
  });
  if (res.status === 401) {
    for (const fn of unauthorizedListeners) fn();
  }
  if (!res.ok) throw new ApiError(res.status, await parseBody(res));
  const data = (await parseBody(res)) as T;
  return { data, res };
}

async function get<T>(path: string, signal?: AbortSignal): Promise<T> {
  return (await request<T>(path, { signal })).data;
}

async function post<T>(path: string, json?: unknown): Promise<T> {
  return (await request<T>(path, { method: "POST", json })).data;
}

export interface LoadedConfig {
  config: ConfigView;
  revision: number;
}

function revisionOf(res: Response, fallback = 0): number {
  const n = Number(res.headers.get("X-Config-Revision"));
  return Number.isFinite(n) && n > 0 ? n : fallback;
}

export const api = {
  async config(): Promise<LoadedConfig> {
    const { data, res } = await request<ConfigView>("/api/config");
    return { config: data, revision: revisionOf(res) };
  },
  async saveConfig(changes: Record<string, unknown>, revision: number): Promise<ConfigSaveResponse> {
    const { data } = await request<ConfigSaveResponse>("/api/config", {
      method: "POST",
      json: changes,
      headers: { "If-Match": `"${revision}"` },
    });
    return data;
  },
  setLanguage: (lang: "ar" | "en") => post<{ ok: boolean }>(`/api/language?lang=${lang}`),
  logs: (date: string, limit = 5000, signal?: AbortSignal) =>
    get<LogEvent[]>(`/api/logs?date=${encodeURIComponent(date)}&limit=${limit}`, signal),
  logDates: () => get<string[]>("/api/log-dates"),
  availability: () => get<Availability>("/api/availability"),
  report: (month: string, signal?: AbortSignal) =>
    get<MonthlyReport>(`/api/report?month=${encodeURIComponent(month)}`, signal),
  speedHistory: (date: string, signal?: AbortSignal) =>
    get<SpeedTestEvent[]>(`/api/speed-test/history?date=${encodeURIComponent(date)}&limit=50`, signal),
  startSpeedTest: () => post<{ ok: boolean }>("/api/speed-test/start"),
  cancelSpeedTest: () => post<{ ok: boolean }>("/api/speed-test/cancel"),
  testTargets: (targets: { ping_targets: string[]; http_targets: string[]; dns_targets: string[] }) =>
    post<TestTargetsResponse>("/api/test-targets", { ...targets, notify: false }),
  testNotification: (lang: "ar" | "en") => post<{ ok: boolean }>(`/api/test-notification?lang=${lang}`),
  testWebhook: (url?: string) => post<{ ok: boolean }>("/api/test-webhook", url ? { url } : {}),
  playSound: () => post<{ ok: boolean }>("/api/play-sound"),
  async soundInfo(): Promise<{ custom: boolean; duration: number }> {
    const res = await fetch("/notification-sound", { method: "HEAD", cache: "no-store" });
    if (res.status === 401) for (const fn of unauthorizedListeners) fn();
    if (!res.ok) throw new ApiError(res.status, {});
    const duration = Number(res.headers.get("X-Sound-Duration"));
    return {
      custom: res.headers.get("X-Custom-Sound") === "1",
      duration: Number.isFinite(duration) ? duration : 0,
    };
  },
  async uploadSound(file: File): Promise<void> {
    const form = new FormData();
    form.append("sound", file);
    await request("/notification-sound", { method: "POST", form });
  },
  resetSound: () => request("/notification-sound", { method: "DELETE" }),
  startup: () => get<StartupState>("/api/startup"),
  setStartup: (enabled: boolean) => post<StartupState>("/api/startup", { enabled }),
  checkUpdate: () => post<UpdateInfo>("/api/update/check"),
  applyUpdate: () => post<{ ok: boolean; status: string }>("/api/update"),
};
