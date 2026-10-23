import { useSyncExternalStore } from "react";

export type Route = "overview" | "history" | "speed" | "settings" | "report";

const paths: Record<Route, string> = {
  overview: "/",
  history: "/history",
  speed: "/speed",
  settings: "/settings",
  report: "/report",
};

type Blocker = (proceed: () => void) => boolean;

const listeners = new Set<() => void>();
let blocker: Blocker | null = null;
let current = typeof window === "undefined" ? "/" : window.location.pathname + window.location.search;

function emit() {
  current = window.location.pathname + window.location.search;
  for (const fn of listeners) fn();
}

function onPopState() {
  const target = window.location.pathname + window.location.search;
  if (!blocker || target === current) {
    emit();
    return;
  }
  const proceed = () => {
    window.history.replaceState(null, "", target);
    emit();
  };
  window.history.pushState(null, "", current);
  if (!blocker(proceed)) proceed();
}

if (typeof window !== "undefined") {
  window.addEventListener("popstate", onPopState);
}

export function routeOf(pathname: string): Route {
  const p = pathname.replace(/\/+$/, "") || "/";
  for (const [route, path] of Object.entries(paths) as [Route, string][]) {
    if (path === p) return route;
  }
  return "overview";
}

export function hrefOf(route: Route, search = ""): string {
  return paths[route] + search;
}

export function setNavigationBlocker(fn: Blocker | null) {
  blocker = fn;
}

export function navigate(route: Route, search = "", replace = false) {
  const target = hrefOf(route, search);
  const go = () => {
    if (target === current) return;
    if (replace) window.history.replaceState(null, "", target);
    else window.history.pushState(null, "", target);
    emit();
    window.scrollTo({ top: 0 });
  };
  if (blocker?.(go)) return;
  go();
}

function subscribe(fn: () => void) {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}

export function useLocation() {
  const location = useSyncExternalStore(subscribe, () => current);
  const [pathname = "/", search = ""] = location.split("?");
  return { route: routeOf(pathname), params: new URLSearchParams(search) };
}
