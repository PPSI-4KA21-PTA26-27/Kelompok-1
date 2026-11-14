import { describe, expect, it } from "vitest";
import { cn } from "./utils";
import { cn as cnLegacy } from "./cn";

describe("cn (utils, clsx + tailwind-merge)", () => {
  it("menggabungkan dan mengabaikan nilai falsy", () => {
    expect(cn("a", false, null, undefined, "b")).toBe("a b");
  });

  it("mendukung objek dan array clsx", () => {
    expect(cn("a", { b: true, c: false }, ["d", ["e"]])).toBe("a b d e");
  });

  it("kelas tailwind yang bentrok: yang terakhir menang", () => {
    expect(cn("px-2 py-1", "px-4")).toBe("py-1 px-4");
    expect(cn("text-red-500", "text-blue-500")).toBe("text-blue-500");
  });
});

describe("cn lama (@/lib/cn)", () => {
  it("menggabungkan string dan membuang falsy tanpa merge tailwind", () => {
    expect(cnLegacy("a", false, null, undefined, "", "b")).toBe("a b");
    expect(cnLegacy("px-2", "px-4")).toBe("px-2 px-4");
  });
});
