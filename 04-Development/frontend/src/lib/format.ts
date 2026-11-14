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
  pending: "Menunggu",
  approved: "Disetujui",
  rejected: "Ditolak",
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
