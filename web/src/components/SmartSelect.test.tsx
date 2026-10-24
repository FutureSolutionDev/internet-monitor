import { act, cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeAll, describe, expect, test, vi } from "vitest";
import { type SmartOption, SmartSelect } from "./SmartSelect";

const sources = import.meta.glob("../**/*.tsx", { query: "?raw", import: "default", eager: true }) as Record<
  string,
  string
>;

const fruits: SmartOption[] = ["Apple", "Banana", "Cherry", "Date", "Elder", "Fig", "Grape", "Honeydew"].map(
  (label) => ({ value: label.toLowerCase(), label }),
);

async function settle() {
  await act(async () => {
    await new Promise((r) => setTimeout(r, 0));
  });
}

beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn();
});

afterEach(cleanup);

describe("SmartSelect", () => {
  test("the app never renders a native select", () => {
    for (const [path, code] of Object.entries(sources)) {
      if (path.endsWith(".test.tsx")) continue;
      expect(/<select[\s>]/.test(code), path).toBe(false);
    }
  });

  test("opens, moves with the keyboard and commits", async () => {
    const onChange = vi.fn();
    render(<SmartSelect label="Fruit" value="apple" onChange={onChange} options={fruits.slice(0, 3)} />);
    const trigger = screen.getByRole("button", { name: /^Fruit:/ });
    fireEvent.click(trigger);
    await settle();
    const list = screen.getByRole("listbox", { name: "Fruit" });
    expect(document.activeElement).toBe(list);
    expect(screen.getByRole("option", { name: "Apple" }).getAttribute("aria-selected")).toBe("true");

    fireEvent.keyDown(list, { key: "ArrowDown" });
    fireEvent.keyDown(list, { key: "ArrowDown" });
    fireEvent.keyDown(list, { key: "ArrowDown" });
    expect(list.getAttribute("aria-activedescendant")).toBe(
      screen.getByRole("option", { name: "Cherry" }).id,
    );
    fireEvent.keyDown(list, { key: "Enter" });
    expect(onChange).toHaveBeenCalledWith("cherry");
    expect(screen.queryByRole("listbox")).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });

  test("escape closes without changing and returns focus", async () => {
    const onChange = vi.fn();
    render(<SmartSelect label="Fruit" value="apple" onChange={onChange} options={fruits.slice(0, 3)} />);
    const trigger = screen.getByRole("button", { name: /^Fruit:/ });
    fireEvent.keyDown(trigger, { key: "ArrowDown" });
    await settle();
    fireEvent.keyDown(screen.getByRole("listbox"), { key: "Escape" });
    expect(screen.queryByRole("listbox")).toBeNull();
    expect(onChange).not.toHaveBeenCalled();
    expect(document.activeElement).toBe(trigger);
  });

  test("long lists get a search box that filters", async () => {
    const onChange = vi.fn();
    render(<SmartSelect label="Fruit" value={null} onChange={onChange} options={fruits} />);
    expect(screen.getByRole("button", { name: /^Fruit:/ }).textContent).toContain("Choose");
    fireEvent.click(screen.getByRole("button", { name: /^Fruit:/ }));
    await settle();
    const search = screen.getByRole("combobox");
    expect(document.activeElement).toBe(search);
    fireEvent.change(search, { target: { value: "GR" } });
    expect(screen.getAllByRole("option").map((o) => o.textContent)).toEqual(["Grape"]);
    fireEvent.keyDown(search, { key: "Enter" });
    expect(onChange).toHaveBeenCalledWith("grape");

    fireEvent.click(screen.getByRole("button", { name: /^Fruit:/ }));
    await settle();
    fireEvent.change(screen.getByRole("combobox"), { target: { value: "zzz" } });
    expect(screen.queryAllByRole("option")).toHaveLength(0);
    expect(screen.getByText("No matches")).toBeTruthy();
  });

  test("typeahead jumps to a matching option", async () => {
    render(<SmartSelect label="Fruit" value="apple" onChange={() => {}} options={fruits.slice(0, 5)} />);
    fireEvent.click(screen.getByRole("button", { name: /^Fruit:/ }));
    await settle();
    const list = screen.getByRole("listbox");
    fireEvent.keyDown(list, { key: "d" });
    expect(list.getAttribute("aria-activedescendant")).toBe(screen.getByRole("option", { name: "Date" }).id);
  });

  test("clicking outside closes, disabled never opens", async () => {
    const { rerender } = render(
      <SmartSelect label="Fruit" value="apple" onChange={() => {}} options={fruits.slice(0, 3)} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /^Fruit:/ }));
    await settle();
    fireEvent.pointerDown(document.body);
    expect(screen.queryByRole("listbox")).toBeNull();

    rerender(<SmartSelect label="Fruit" value="apple" onChange={() => {}} options={fruits} disabled />);
    fireEvent.click(screen.getByRole("button", { name: /^Fruit:/ }));
    await settle();
    expect(screen.queryByRole("listbox")).toBeNull();
  });
});
