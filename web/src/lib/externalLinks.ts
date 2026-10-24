type NativeWindow = Window & { nativeOpenURL?: (url: string) => Promise<void> };

export function handleExternalClick(event: MouseEvent, win: NativeWindow = window) {
  if (event.defaultPrevented || event.button !== 0 || typeof win.nativeOpenURL !== "function") return;
  const target = event.target instanceof Element ? event.target.closest("a[target='_blank']") : null;
  if (!(target instanceof HTMLAnchorElement) || !target.href.startsWith("https://")) return;
  event.preventDefault();
  void win.nativeOpenURL(target.href);
}

export function installExternalLinks() {
  document.addEventListener("click", (e) => handleExternalClick(e));
}
