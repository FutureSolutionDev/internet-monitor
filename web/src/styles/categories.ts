export const tone = {
  connected: "var(--accent-green)",
  degraded: "var(--accent-amber)",
  disconnected: "var(--accent-coral)",
  checking: "var(--accent-blue)",
  latency: "var(--accent-blue)",
  jitter: "var(--accent-violet)",
  loss: "var(--accent-coral)",
  uptime: "var(--accent-green)",
  download: "var(--accent-blue)",
  upload: "var(--accent-violet)",
  ping: "var(--accent-amber)",
} as const;

export function statusTone(status: string): string {
  switch (status) {
    case "connected":
      return tone.connected;
    case "degraded":
      return tone.degraded;
    case "disconnected":
      return tone.disconnected;
    default:
      return tone.checking;
  }
}

export function availabilityTone(pct: number | null | undefined): string {
  if (pct === null || pct === undefined) return "var(--text-muted)";
  if (pct >= 99) return tone.connected;
  if (pct >= 95) return tone.degraded;
  return tone.disconnected;
}
