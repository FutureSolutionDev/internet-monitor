import { describe, expect, test, vi } from "vitest";
import { handleExternalClick } from "./externalLinks";

function clickOn(html: string, nativeOpenURL?: (url: string) => Promise<void>) {
  document.body.innerHTML = html;
  const el = document.body.querySelector("[data-hit]") ?? document.body.firstElementChild;
  const event = new MouseEvent("click", { bubbles: true, cancelable: true, button: 0 });
  Object.defineProperty(event, "target", { value: el });
  handleExternalClick(event, { nativeOpenURL } as unknown as Window);
  return event;
}

describe("external links", () => {
  test("hands new-tab https links to the native window", () => {
    const open = vi.fn().mockResolvedValue(undefined);
    const e = clickOn('<a href="https://wa.me/1" target="_blank"><span data-hit>x</span></a>', open);
    expect(open).toHaveBeenCalledWith("https://wa.me/1");
    expect(e.defaultPrevented).toBe(true);
  });

  test("leaves links alone in a normal browser", () => {
    const e = clickOn('<a href="https://wa.me/1" target="_blank">x</a>');
    expect(e.defaultPrevented).toBe(false);
  });

  test("ignores same-tab and mailto links", () => {
    const open = vi.fn().mockResolvedValue(undefined);
    clickOn('<a href="/faq">x</a>', open);
    clickOn('<a href="mailto:a@b.c" target="_blank">x</a>', open);
    expect(open).not.toHaveBeenCalled();
  });
});
