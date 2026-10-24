import { afterEach, describe, expect, test, vi } from "vitest";
import { changeThemeWithTransition } from "./themeTransition";

type Doc = { startViewTransition?: unknown };

function stubMotion(reduced: boolean) {
  window.matchMedia = vi.fn().mockReturnValue({ matches: reduced }) as unknown as typeof window.matchMedia;
}

afterEach(() => {
  delete (document as unknown as Doc).startViewTransition;
  document.documentElement.removeAttribute("data-theme");
  document.documentElement.classList.remove("theme-switching");
});

describe("theme transition", () => {
  test("switches immediately when view transitions are unavailable", () => {
    stubMotion(false);
    changeThemeWithTransition("light", { clientX: 10, clientY: 10 });
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  });

  test("skips the animation when reduced motion is preferred", () => {
    stubMotion(true);
    const start = vi.fn();
    (document as unknown as Doc).startViewTransition = start;
    changeThemeWithTransition("light", { clientX: 10, clientY: 10 });
    expect(start).not.toHaveBeenCalled();
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  });

  test("runs the theme change inside a view transition", () => {
    stubMotion(false);
    const start = vi.fn((cb: () => void) => {
      cb();
      return { ready: new Promise(() => {}), finished: Promise.resolve() };
    });
    (document as unknown as Doc).startViewTransition = start;
    changeThemeWithTransition("light", { clientX: 10, clientY: 10 });
    expect(start).toHaveBeenCalledOnce();
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  });

  test("does nothing when the theme is already active", () => {
    stubMotion(false);
    document.documentElement.setAttribute("data-theme", "dark");
    const start = vi.fn();
    (document as unknown as Doc).startViewTransition = start;
    changeThemeWithTransition("dark", { clientX: 0, clientY: 0 });
    expect(start).not.toHaveBeenCalled();
  });
});
