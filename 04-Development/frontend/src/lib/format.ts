import type { Tone } from "@/components/ui/misc";

/** Angka / string desimal backend -> Rupiah, mis. "150000.00" -> "Rp150.000". */
export function formatRupiah(value: number | string | null | undefined): string {
  const n = typeof value === "string" ? Number(value) : (value ?? 0);
  if (!Number.isFinite(n)) return "Rp0";
  return "Rp" + new Intl.NumberFormat("id-ID", { maximumFractionDigits: 0 }).format(Math.round(n));
}

export function formatDate(iso?: string | null, withTime = false): string {
  if (!iso) return "-";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "-";
  return d.toLocaleString("id-ID", {
    timeZone: "Asia/Jakarta",
    day: "numeric",
    month: "short",
    year: "numeric",
    ...(withTime ? { hour: "2-digit", minute: "2-digit" } : {}),
  });
}

/** Alamat kirim lengkap, melewati bagian yang kosong/"-". */
export function joinAddress(o: {
  shipping_address: string;
  shipping_kelurahan?: string | null;
  shipping_district?: string | null;
  shipping_city?: string | null;
  shipping_zip_code?: string | null;
}): string {
  const parts = [o.shipping_address, o.shipping_kelurahan, o.shipping_district, o.shipping_city]
    .map((p) => (p ?? "").trim())
    .filter((p) => p && p !== "-");
  const zip = (o.shipping_zip_code ?? "").trim();
  // NBSP agar "RT 001/RW 002" tidak terpotong antar-baris.
  return parts.join(", ").replace(/RT (\d{3})\/RW (\d{3})/, "RT\u00a0$1/RW\u00a0$2") + (zip && zip !== "0" ? " " + zip : "");
}

/** Harga setelah diskon (dibulatkan seperti backend). */
export function finalPrice(price: number | string, discount = 0): number {
  return Math.round((Number(price) || 0) * (1 - (discount || 0) / 100));
}

/** Label + nada kolom "Pembayaran" di daftar pesanan admin. */
export function paymentLabel(o: { payment_status: string | null; has_proof: boolean }): { label: string; tone: Tone } {
  if (o.payment_status === "valid") return { label: "Valid", tone: "lime" };
  if (o.payment_status === "ditolak") return { label: "Ditolak", tone: "danger" };
  if (o.payment_status === "menunggu" && o.has_proof) return { label: "Menunggu verifikasi", tone: "warn" };
  return { label: "Belum ada bukti", tone: "neutral" };
}

/** Seri harian (urut tanggal) -> total pendapatan per 7 hari: Minggu 1, Minggu 2, ...; sisa hari (<7) digabung ke minggu terakhir. */
export function weeklyRevenue(series: { revenue: string }[]): number[] {
  const weeks: number[] = [];
  const last = Math.max(0, Math.floor(series.length / 7) - 1);
  series.forEach((d, i) => {
    const w = Math.min(Math.floor(i / 7), last);
    weeks[w] = (weeks[w] ?? 0) + (Number(d.revenue) || 0);
  });
  return weeks;
}

/** Pecah alamat tersimpan "Jl. X, RT 001/RW 002" (format backend) jadi jalan + RT + RW untuk isi ulang form. */
export function splitRTRW(address: string): { address: string; rt: string; rw: string } {
  const m = address.match(/^(.*), RT (\d{3}|-)\/RW (\d{3}|-)$/);
  if (!m) return { address, rt: "", rw: "" };
  const num = (v: string) => (v === "-" ? "" : String(Number(v)));
  return { address: m[1], rt: num(m[2]), rw: num(m[3]) };
}

/** RT dan RW sama-sama tercatat di alamat (pesanan lama bisa belum punya). */
export function hasRTRW(address: string): boolean {
  const { rt, rw } = splitRTRW(address);
  return !!rt && !!rw;
}

/** Sisa waktu bayar singkat: "2 j 05 mnt", "45 mnt", atau "" bila sudah lewat. */
export function formatRemaining(ms: number): string {
  if (!(ms > 0)) return "";
  const m = Math.ceil(ms / 60000);
  const h = Math.floor(m / 60);
  return h ? `${h} j ${String(m % 60).padStart(2, "0")} mnt` : `${m} mnt`;
}

/** 12 bulan terakhir (terbaru dulu) sebagai pilihan "YYYY-MM" untuk L-29, bulan menurut WIB. */
export function lastMonths(now: Date, n = 12): { value: string; label: string }[] {
  const [y, m] = new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Jakarta", year: "numeric", month: "2-digit" })
    .format(now)
    .split("-")
    .map(Number);
  return Array.from({ length: n }, (_, i) => {
    const d = new Date(Date.UTC(y, m - 1 - i, 1));
    return {
      value: d.getUTCFullYear() + "-" + String(d.getUTCMonth() + 1).padStart(2, "0"),
      label: d.toLocaleDateString("id-ID", { month: "long", year: "numeric", timeZone: "UTC" }),
    };
  });
}
