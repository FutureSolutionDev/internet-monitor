import { ApiError } from "./api";
import type { MessageKey } from "./i18n";

const BY_CODE: Record<string, MessageKey> = {
  rate_limited: "err_rate_limited",
  unsupported: "err_webhook",
  not_configured: "err_webhook_missing",
  webhook_failed: "err_webhook_failed",
  in_progress: "update_in_progress",
  no_update: "update_latest",
  unavailable: "err_unavailable",
  check_failed: "update_check_failed",
};

export function errorKey(e: unknown, fallback: MessageKey): MessageKey {
  if (e instanceof ApiError && e.code && BY_CODE[e.code]) return BY_CODE[e.code] as MessageKey;
  return fallback;
}

const SPEED: Record<string, MessageKey> = {
  internal_error: "speed_err_failed",
  no_result: "speed_err_no_result",
  failed: "speed_err_failed",
  test_already_running: "speed_err_running",
};

export function speedErrorKey(code: string | undefined): MessageKey {
  return (code && SPEED[code]) || "speed_err_failed";
}
