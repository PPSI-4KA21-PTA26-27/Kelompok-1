import { describe, expect, it } from "vitest";
import { formatRemaining, hasRTRW, lastMonths, paymentLabel, splitRTRW, weeklyRevenue } from "./format";

describe("paymentLabel", () => {
  it("membedakan belum ada bukti, menunggu, valid, ditolak", () => {
    expect(paymentLabel({ payment_status: null, has_proof: false }).label).toBe("Belum ada bukti");
    expect(paymentLabel({ payment_status: "menunggu", has_proof: false }).label).toBe("Belum ada bukti");
    expect(paymentLabel({ payment_status: "menunggu", has_proof: true }).label).toBe("Menunggu verifikasi");
    expect(paymentLabel({ payment_status: "valid", has_proof: true }).label).toBe("Valid");
    expect(paymentLabel({ payment_status: "ditolak", has_proof: true }).label).toBe("Ditolak");
  });
});

describe("weeklyRevenue", () => {
  it("menjumlah per 7 hari, sisa hari digabung ke minggu terakhir", () => {
    const s = Array.from({ length: 9 }, () => ({ revenue: "100.00" }));
    expect(weeklyRevenue(s)).toEqual([900]);
    expect(weeklyRevenue(Array.from({ length: 30 }, () => ({ revenue: "1" })))).toHaveLength(4);
    expect(weeklyRevenue([])).toEqual([]);
  });
});

describe("splitRTRW", () => {
  it("memisah RT/RW dari alamat tersimpan", () => {
    expect(splitRTRW("Jl. Mawar 1, RT 001/RW 023")).toEqual({ address: "Jl. Mawar 1", rt: "1", rw: "23" });
    expect(splitRTRW("Jl. Mawar 1, RT 005/RW -")).toEqual({ address: "Jl. Mawar 1", rt: "5", rw: "" });
    expect(splitRTRW("Jl. Mawar 1")).toEqual({ address: "Jl. Mawar 1", rt: "", rw: "" });
  });
});

describe("hasRTRW", () => {
  it("true hanya bila RT dan RW sama-sama terisi", () => {
    expect(hasRTRW("Jl. Mawar 1, RT 001/RW 023")).toBe(true);
    expect(hasRTRW("Jl. Mawar 1, RT 005/RW -")).toBe(false);
    expect(hasRTRW("Jl. Mawar 1")).toBe(false);
  });
});

describe("formatRemaining", () => {
  it("jam + menit, menit saja, dan lewat", () => {
    expect(formatRemaining(125 * 60000)).toBe("2 j 05 mnt");
    expect(formatRemaining(30 * 1000)).toBe("1 mnt");
    expect(formatRemaining(0)).toBe("");
  });
});

describe("lastMonths", () => {
  it("12 bulan terakhir dari bulan berjalan, lintas tahun", () => {
    const m = lastMonths(new Date("2026-02-15T05:00:00Z"));
    expect(m).toHaveLength(12);
    expect(m[0].value).toBe("2026-02");
    expect(m[1].value).toBe("2026-01");
    expect(m[2].value).toBe("2025-12");
    expect(m[11].value).toBe("2025-03");
  });
  it("memakai bulan WIB: 30 Nov 18:00 UTC = 1 Des WIB", () => {
    const m = lastMonths(new Date("2026-11-30T18:00:00Z"));
    expect(m[0].value).toBe("2026-12");
    expect(m[0].label).toBe("Desember 2026");
  });
});

describe("weeklyRevenue (bulan kalender L-29)", () => {
  it("31 hari menjadi 4 minggu; hari 29-31 masuk minggu ke-4", () => {
    const s = Array.from({ length: 31 }, () => ({ revenue: "10" }));
    expect(weeklyRevenue(s)).toEqual([70, 70, 70, 100]);
  });
});
