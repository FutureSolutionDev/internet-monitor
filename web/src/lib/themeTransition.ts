import { flushSync } from "react-dom";
import { setTheme, type Theme } from "./theme";

const DURATION_MS = 1000;
let generation = 0;

export interface TransitionOrigin {
  clientX: number;
  clientY: number;
  detail?: number;
  currentTarget?: EventTarget | null;
}

function originOf(event: TransitionOrigin): { x: number; y: number } {
  const target = event.currentTarget;
  if ((event.detail === 0 || (event.clientX === 0 && event.clientY === 0)) && target instanceof Element) {
    const rect = target.getBoundingClientRect();
    return { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 };
  }
  return { x: event.clientX, y: event.clientY };
}

function prefersReducedMotion(): boolean {
  try {
    return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  } catch {
    return false;
  }
}

export function changeThemeWithTransition(next: Theme, event: TransitionOrigin) {
  const root = document.documentElement;
  if (root.getAttribute("data-theme") === next) return;
  const apply = () => flushSync(() => setTheme(next));
  const mine = ++generation;
  root.classList.add("theme-switching");
  const done = () => {
    if (mine === generation) root.classList.remove("theme-switching");
  };
  if (typeof document.startViewTransition !== "function" || prefersReducedMotion()) {
    try {
      apply();
    } finally {
      requestAnimationFrame(done);
    }
    return;
  }
  const { x, y } = originOf(event);
  const radius = Math.hypot(Math.max(x, window.innerWidth - x), Math.max(y, window.innerHeight - y));
  let transition: ViewTransition;
  try {
    transition = document.startViewTransition(apply);
  } catch {
    done();
    setTheme(next);
    return;
  }
  transition.finished.then(done, done);
  transition.ready
    .then(() => {
      root.animate(
        { clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${radius}px at ${x}px ${y}px)`] },
        { duration: DURATION_MS, easing: "ease-in-out", pseudoElement: "::view-transition-new(root)" },
      );
    })
    .catch(() => undefined);
}
