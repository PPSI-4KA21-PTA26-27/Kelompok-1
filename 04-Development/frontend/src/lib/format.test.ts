import { describe, expect, it } from "vitest";
import { formatRemaining, hasRTRW, paymentLabel, splitRTRW, weeklyRevenue } from "./format";

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
