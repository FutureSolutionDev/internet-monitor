import { describe, expect, test } from "vitest";
import { placeTip } from "./ChartTooltip";

describe("placeTip", () => {
  test("centres above the anchor", () => {
    expect(placeTip(200, 100, 400, 100, 40, 300)).toEqual({ left: 150, top: 48, flipped: false });
  });

  test("stays inside the chart horizontally", () => {
    expect(placeTip(10, 100, 400, 100, 40, 300).left).toBe(0);
    expect(placeTip(395, 100, 400, 100, 40, 300).left).toBe(300);
  });

  test("may rise above the chart while the viewport has room", () => {
    expect(placeTip(200, 10, 400, 100, 60, 300)).toMatchObject({ top: -62, flipped: false });
  });

  test("flips below when it would leave the viewport", () => {
    expect(placeTip(200, 10, 400, 100, 60, 20)).toMatchObject({ top: 22, flipped: true });
  });
});
