import { describe, expect, it } from "vitest";
import { isSafeLink } from "./safeLink";

describe("isSafeLink", () => {
  it("menerima path dalam situs dan http(s)", () => {
    for (const s of ["/", "/katalog", "/produk/1?x=2", "http://a.test", "https://a.test/b"]) expect(isSafeLink(s)).toBe(true);
  });
  it("menolak open redirect dan skema berbahaya", () => {
    for (const s of ["//evil.com", "/\\evil.com", "javascript:alert(1)", "data:text/html,x", "evil.com", ""]) expect(isSafeLink(s)).toBe(false);
  });
});
