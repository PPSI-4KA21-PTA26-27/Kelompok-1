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
    day: "numeric",
    month: "short",
    year: "numeric",
    ...(withTime ? { hour: "2-digit", minute: "2-digit" } : {}),
  });
}

/** Waktu relatif singkat: "5 mnt lalu", "2 jam lalu", "3 hari lalu". */
export function timeAgo(iso?: string | null): string {
  if (!iso) return "-";
  const diff = (Date.now() - new Date(iso).getTime()) / 1000;
  if (Number.isNaN(diff)) return "-";
  if (diff < 60) return "baru saja";
  if (diff < 3600) return Math.floor(diff / 60) + " mnt lalu";
  if (diff < 86400) return Math.floor(diff / 3600) + " jam lalu";
  if (diff < 86400 * 7) return Math.floor(diff / 86400) + " hari lalu";
  return formatDate(iso);
}

export const ORDER_STATUS_LABEL: Record<string, string> = {
  pending_payment: "Menunggu pembayaran",
  paid: "Dibayar",
  processing: "Dikemas",
  shipped: "Dikirim",
  completed: "Selesai",
  cancelled: "Dibatalkan",
};

export const ORDER_STATUS_TONE: Record<string, Tone> = {
  pending_payment: "warn",
  paid: "info",
  processing: "info",
  shipped: "dark",
  completed: "success",
  cancelled: "danger",
};

export const PAYMENT_STATUS_LABEL: Record<string, string> = {
  menunggu: "Menunggu",
  valid: "Disetujui",
  ditolak: "Ditolak",
};

export const SHIPPING_STATUS_LABEL: Record<string, string> = {
  pending: "Menunggu pickup",
  shipped: "Dalam perjalanan",
  delivered: "Diterima",
  returned: "Dikembalikan",
};

/** Estimasi kirim RajaOngkir ("2 day", "1-2 days") -> "2 hari". */
export function etdLabel(etd?: string | null): string {
  const v = (etd ?? "").replace(/\s*(days?|hari)\s*$/i, "").trim();
  return v ? v + " hari" : "-";
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
  return parts.join(", ") + (zip && zip !== "0" ? " " + zip : "");
}

/** Harga setelah diskon (dibulatkan seperti backend). */
export function finalPrice(price: number | string, discount = 0): number {
  return Math.round((Number(price) || 0) * (1 - (discount || 0) / 100));
}

/** ISO -> nilai <input type="datetime-local"> dalam zona waktu lokal. */
export function toLocalInput(iso?: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const off = d.getTimezoneOffset() * 60000;
  return new Date(d.getTime() - off).toISOString().slice(0, 16);
}

/** Nilai datetime-local -> ISO (RFC 3339) atau null bila kosong. */
export function fromLocalInput(v: string): string | null {
  if (!v) return null;
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? null : d.toISOString();
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

/** Sisa waktu bayar singkat: "2 j 05 mnt", "45 mnt", atau "" bila sudah lewat. */
export function formatRemaining(ms: number): string {
  if (!(ms > 0)) return "";
  const m = Math.ceil(ms / 60000);
  const h = Math.floor(m / 60);
  return h ? `${h} j ${String(m % 60).padStart(2, "0")} mnt` : `${m} mnt`;
}
