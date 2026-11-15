import type { OrderStatus } from "@/types/api";

/** Label & warna status pesanan sisi pelanggan (sesuai desain), plus tab filter. */
export const CUST_STATUS: Record<OrderStatus, { label: string; cls: string }> = {
  pending_payment: { label: "Belum dibayar", cls: "bg-warn-soft text-warn" },
  paid: { label: "Perlu dikirim", cls: "bg-line/40 text-ink" },
  processing: { label: "Perlu dikirim", cls: "bg-line/40 text-ink" },
  shipped: { label: "Dikirim", cls: "bg-success-soft text-success" },
  completed: { label: "Selesai", cls: "bg-lime text-ink" },
  cancelled: { label: "Dibatalkan", cls: "bg-danger-soft text-danger" },
};

export const TABS: { label: string; match: OrderStatus[] | null }[] = [
  { label: "Semua", match: null },
  { label: "Belum dibayar", match: ["pending_payment"] },
  { label: "Perlu dikirim", match: ["paid", "processing"] },
  { label: "Dikirim", match: ["shipped"] },
  { label: "Selesai", match: ["completed"] },
  { label: "Dibatalkan", match: ["cancelled"] },
];

/** Transisi status yang boleh dipilih admin (sesuai PUT /admin/orders/{id}/status; paid lewat verifikasi bayar). */
export const ADMIN_NEXT: Partial<Record<OrderStatus, OrderStatus[]>> = {
  pending_payment: ["cancelled"],
  paid: ["processing", "cancelled"],
  processing: ["shipped", "cancelled"],
  shipped: ["completed"],
};
