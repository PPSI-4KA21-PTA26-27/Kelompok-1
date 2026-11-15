"use client";

import { useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Card, PageHead, stackTable, theadCls } from "@/components/admin/ui";
import { errorMessage } from "@/lib/api";
import { formatRupiah, weeklyRevenue } from "@/lib/format";
import { adminService } from "@/services";
import type { ApiAnalytics } from "@/types/api";

type Days = 7 | 30 | 90;

export default function LaporanPage() {
  const [days, setDays] = useState<Days>(30);
  const [a, setA] = useState<ApiAnalytics | null>(null);

  useEffect(() => {
    adminService.analytics(days).then(setA).catch((e) => toast.error(errorMessage(e)));
  }, [days]);

  const revenue = Number(a?.current.revenue ?? 0);
  const orders = a?.current.paid_orders ?? 0;
  const weeks = weeklyRevenue(a?.series ?? []);
  const max = Math.max(1, ...weeks);

  return (
    <div>
      <PageHead
        title="Laporan Pendapatan"
        sub="Berdasarkan pesanan dengan pembayaran valid."
        action={
          <select
            aria-label="Periode laporan"
            value={days}
            onChange={(e) => setDays(Number(e.target.value) as Days)}
            className="rounded-full border border-line-strong bg-white px-4 py-2.5 font-bold"
          >
            <option value={7}>7 hari terakhir</option>
            <option value={30}>30 hari terakhir</option>
            <option value={90}>90 hari terakhir</option>
          </select>
        }
      />
      <div className="grid gap-5 sm:grid-cols-3">
        {[
          ["Total pendapatan", formatRupiah(revenue)],
          ["Jumlah pesanan", String(orders)],
          ["Rata-rata per pesanan", formatRupiah(orders ? revenue / orders : 0)],
        ].map(([k, v]) => (
          <Card key={k}>
            <p className="font-semibold text-ink-3">{k}</p>
            <p className="mt-1 text-3xl font-extrabold">{v}</p>
          </Card>
        ))}
      </div>

      <Card title="Pendapatan per minggu" className="mt-7 p-8">
        {!weeks.some(Boolean) ? (
          <p className="text-sm text-muted">Belum ada pendapatan pada periode ini.</p>
        ) : (
          <div className="overflow-x-auto">
            <ul className="mx-auto flex h-52 w-max items-end gap-10 border-b border-ink px-10">
              {weeks.map((w, i) => (
                <li key={i} className="flex h-full w-20 flex-col justify-end" title={formatRupiah(w)}>
                  <span className="sr-only">
                    Minggu {i + 1}: {formatRupiah(w)}
                  </span>
                  <span className="rounded-t-xl bg-[#d9d9d9]" style={{ height: Math.max(2, (w / max) * 100) + "%" }} aria-hidden />
                </li>
              ))}
            </ul>
            <div className="mx-auto flex w-max gap-10 px-10 pt-2" aria-hidden>
              {weeks.map((_, i) => (
                <span key={i} className="w-20 text-center text-muted">
                  Minggu {i + 1}
                </span>
              ))}
            </div>
          </div>
        )}
      </Card>

      <Card title="Ringkasan per produk" className="mt-7 p-8">
        <div className="overflow-x-auto rounded-xl border border-line-strong">
          <table className={"w-full text-xs sm:min-w-[480px] " + stackTable}>
            <thead className={theadCls}>
              <tr>
                {["Produk", "Terjual", "Pendapatan"].map((h) => (
                  <th key={h} className="px-6 py-4 font-semibold">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {a?.top_products.map((t) => (
                <tr key={t.product_id} className="border-t border-line-strong">
                  <td className="px-6 py-5 font-bold">{t.name}</td>
                  <td data-l="Terjual" className="px-6 py-5">{t.quantity} pcs</td>
                  <td data-l="Pendapatan" className="px-6 py-5">{formatRupiah(t.revenue)}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {a?.top_products.length === 0 && <p className="py-6 text-center text-sm text-muted">Belum ada penjualan.</p>}
        </div>
      </Card>
    </div>
  );
}
