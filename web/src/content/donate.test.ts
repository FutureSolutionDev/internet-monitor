import { describe, expect, test } from "vitest";
import { DONATE, type DonateConfig, donateMethods } from "./donate";

const EMPTY: DonateConfig = { paypal: "", vodafoneCash: "", instapay: "", instapayLink: "" };

describe("donation methods", () => {
  test("nothing is shown until details are filled in", () => {
    expect(donateMethods(EMPTY)).toEqual([]);
  });

  test("valid details become methods in a fixed order", () => {
    const methods = donateMethods({
      paypal: "https://paypal.me/example/",
      vodafoneCash: "010 1234 5678",
      instapay: "example@instapay",
      instapayLink: "https://ipn.eg/S/example/instapay/abc",
    });
    expect(methods.map((m) => [m.id, m.handle, m.href])).toEqual([
      ["paypal", "paypal.me/example", "https://paypal.me/example/"],
      ["vodafone-cash", "01012345678", undefined],
      ["instapay", "example@instapay", "https://ipn.eg/S/example/instapay/abc"],
    ]);
  });

  test.each([
    ["http PayPal link", { paypal: "http://paypal.me/x" }],
    ["look-alike PayPal host", { paypal: "https://paypal.me.evil.example/x" }],
    ["PayPal link with a user", { paypal: "https://user@paypal.me/x" }],
    ["non-Vodafone number", { vodafoneCash: "01112345678" }],
    ["short number", { vodafoneCash: "0101234567" }],
    ["InstaPay address on another domain", { instapay: "x@gmail.com" }],
  ])("rejects %s", (_, partial) => {
    expect(donateMethods({ ...EMPTY, ...partial })).toEqual([]);
  });

  test("an InstaPay link on another host is dropped but the address stays", () => {
    const [m] = donateMethods({ ...EMPTY, instapay: "x@instapay", instapayLink: "https://evil.example/pay" });
    expect(m?.href).toBeUndefined();
  });

  test("every detail filled in donate.ts is valid", () => {
    const shown = new Set(donateMethods().map((m) => m.id));
    if (DONATE.paypal.trim()) expect(shown.has("paypal"), "paypal").toBe(true);
    if (DONATE.vodafoneCash.trim()) expect(shown.has("vodafone-cash"), "vodafoneCash").toBe(true);
    if (DONATE.instapay.trim()) expect(shown.has("instapay"), "instapay").toBe(true);
    if (DONATE.instapayLink.trim())
      expect(donateMethods().find((m) => m.id === "instapay")?.href).toBeTruthy();
  });
});
