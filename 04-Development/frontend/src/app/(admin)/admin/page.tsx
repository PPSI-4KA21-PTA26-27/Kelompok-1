"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Card, PageHead, btnLine, stackTable, theadCls } from "@/components/admin/ui";
import { CUST_STATUS } from "@/components/order/status";
import { errorMessage } from "@/lib/api";
import { cn } from "@/lib/cn";
import { formatDate, formatRupiah } from "@/lib/format";
import { adminService } from "@/services";
import type { ApiSummary } from "@/types/api";

// L-18 Dashboard Ringkasan (fitur tambahan D-19): ringkasan kerja harian
// staff dan owner dari GET /admin/summary.
const SHORTCUTS = [
  { href: "/admin/pembayaran", label: "Validasi Pembayaran" },
  { href: "/admin/pesanan", label: "Pesanan Masuk" },
  { href: "/admin/produk/baru", label: "Tambah Produk" },
  { href: "/admin/stok", label: "Perbarui Stok" },
];

export default function AdminDashboard() {
  const [s, setS] = useState<ApiSummary | null>(null);

  useEffect(() => {
    adminService
      .summary()
      .then(setS)
      .catch((e) => toast.error(errorMessage(e)));
  }, []);

  const toShip = (s?.status_counts.paid ?? 0) + (s?.status_counts.processing ?? 0);
  const cards: [string, number | undefined][] = [
    ["Menunggu verifikasi", s?.proofs_to_review],
    ["Perlu dikirim", s ? toShip : undefined],
    ["Pesanan hari ini", s?.orders_today],
    ["Stok habis", s?.out_of_stock],
  ];

  return (
    <div>
      <PageHead title="Dashboard Ringkasan" sub="Ringkasan kerja harian dari data terkini." />

      <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {cards.map(([k, v]) => (
          <Card key={k}>
            <p className="font-semibold text-ink-3">{k}</p>
            <p className="mt-1 text-3xl font-extrabold">{v ?? "-"}</p>
          </Card>
        ))}
      </div>

      <Card title="Pintasan" className="mt-7">
        <div className="flex flex-wrap gap-3">
          {SHORTCUTS.map((l) => (
            <Link key={l.href} href={l.href} className={btnLine + " text-sm"}>
              {l.label}
            </Link>
          ))}
        </div>
      </Card>

      <Card title="Pesanan terbaru" className="mt-7 sm:p-8">
        {s && s.recent.length === 0 ? (
          <p className="py-6 text-center text-sm text-muted">Belum ada pesanan.</p>
        ) : (
          <div className="overflow-x-auto rounded-xl border border-line-strong">
            <table className={"w-full text-xs sm:min-w-160 " + stackTable}>
              <thead className={theadCls}>
                <tr>
                  {["ID Pesanan", "Tanggal", "Penerima", "Total", "Status"].map((h) => (
                    <th key={h} className="px-4 py-4 font-semibold">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {s?.recent.map((o) => {
                  const st = CUST_STATUS[o.order_status];
                  return (
                    <tr key={o.id} className="border-t border-line-strong">
                      <td className="px-4 py-3">
                        <Link href={"/admin/pesanan/" + o.id} className="font-bold underline">
                          #{o.order_number}
                        </Link>
                      </td>
                      <td data-l="Tanggal" className="px-4 py-3">
                        {formatDate(o.created_at)}
                      </td>
                      <td data-l="Penerima" className="px-4 py-3 font-semibold">
                        {o.recipient_name}
                      </td>
                      <td data-l="Total" className="px-4 py-3">
                        {formatRupiah(o.payment_amount ?? o.total_amount)}
                      </td>
                      <td data-l="Status" className="px-4 py-3">
                        <span className={cn("rounded-full px-2 py-0.5 text-[10px] font-bold", st?.cls)}>
                          {st?.label ?? o.order_status}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}
