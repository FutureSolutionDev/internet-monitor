import { navigate, type Route } from "./router";

const FIND_FOR = 4_000;
const FIND_EVERY = 100;
const GLOW_FOR = 2_600;

let search: number | null = null;

function reducedMotion(): boolean {
  return (
    typeof window.matchMedia === "function" && window.matchMedia("(prefers-reduced-motion: reduce)").matches
  );
}

export function spotlight(el: HTMLElement) {
  el.scrollIntoView({ block: "center", behavior: reducedMotion() ? "auto" : "smooth" });
  if (!el.matches("a[href], button, input, select, textarea, [tabindex]")) el.setAttribute("tabindex", "-1");
  el.focus({ preventScroll: true });
  el.classList.remove("spotlight");
  void el.offsetWidth;
  el.classList.add("spotlight");
  window.setTimeout(() => el.classList.remove("spotlight"), GLOW_FOR);
}

export function goToTarget(route: Route, anchor?: string) {
  if (search !== null) window.clearInterval(search);
  search = null;
  navigate(route);
  if (!anchor) return;
  const started = Date.now();
  search = window.setInterval(() => {
    const el = document.getElementById(anchor);
    if (el) {
      if (search !== null) window.clearInterval(search);
      search = null;
      spotlight(el);
    } else if (Date.now() - started > FIND_FOR && search !== null) {
      window.clearInterval(search);
      search = null;
    }
  }, FIND_EVERY);
}
