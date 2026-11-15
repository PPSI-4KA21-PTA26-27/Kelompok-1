"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuth } from "@/app/context/AuthContext";
import Countdown from "@/components/order/Countdown";
import { CUST_STATUS, TABS } from "@/components/order/status";
import { cn } from "@/lib/cn";
import { formatDate, formatRupiah } from "@/lib/format";
import { orderService } from "@/services";
import type { ApiOrderListItem } from "@/types/api";

export default function PesananPage() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const [orders, setOrders] = useState<ApiOrderListItem[] | null>(null);
  const [tab, setTab] = useState(0);

  useEffect(() => {
    if (loading) return;
    if (!user) return router.replace("/masuk?next=/pesanan");
    // shortcut: 50 pesanan terbaru, filter di klien; pakai paginasi bila lebih
    orderService.listMine({ limit: 50 }).then((r) => setOrders(r.data)).catch(() => setOrders([]));
  }, [loading, user, router]);

  const match = TABS[tab].match;
  const shown = orders?.filter((o) => !match || match.includes(o.order_status)) ?? [];
  const btn = "rounded-lg border border-ink px-3 py-2 text-xs font-bold";

  return (
    <div>
      <p className="text-sm text-muted lg:text-lg">
        <Link href="/">Beranda</Link> <span className="mx-1 text-sm">&gt;</span> <span className="text-ink">Pesanan Saya</span>
      </p>
      <h1 className="mt-2 text-4xl font-extrabold lg:text-5xl">Pesanan Saya</h1>
      <div className="mt-4 flex flex-wrap gap-2 lg:mt-6 lg:gap-3">
        {TABS.map((t, i) => (
          <button
            key={t.label}
            onClick={() => setTab(i)}
            className={cn("rounded-full border px-3 py-1.5 text-xs font-bold sm:px-5 sm:py-2.5 sm:text-sm lg:px-6 lg:py-4 lg:text-lg", i === tab ? "border-ink bg-ink text-white" : "border-line bg-white")}
          >
            {t.label}
          </button>
        ))}
      </div>

      {orders && !shown.length && <p className="mt-10 text-lg text-muted">Belum ada pesanan.</p>}
      <div className="mt-6 grid gap-4 lg:mt-8 lg:grid-cols-2 lg:gap-8">
        {shown.map((o) => {
          const s = CUST_STATUS[o.order_status];
          return (
            <article key={o.id} className="rounded-3xl border border-line bg-white p-6">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <h2 className="text-xl font-bold">Pesanan #{o.order_number}</h2>
                  <p className="mt-1 text-muted">
                    {formatDate(o.created_at)} · {o.item_count} item
                  </p>
                </div>
                <span className={cn("rounded-full px-2.5 py-1 text-xs font-bold", s.cls)}>{s.label}</span>
              </div>
              <div className="mt-4 text-right">
                <p className="text-muted">Total</p>
                <p className="font-bold">{formatRupiah(o.total_amount)}</p>
              </div>
              <div className="mt-3 flex flex-wrap items-center justify-end gap-2">
                {o.order_status === "pending_payment" && !o.has_proof && <Countdown due={o.payment_due_at} className="mr-auto" />}
                {o.order_status === "pending_payment" && (
                  <Link href={`/pembayaran/${o.id}`} className="rounded-lg bg-lime px-3 py-2 text-xs font-bold">
                    Bayar sekarang
                  </Link>
                )}
                <Link href={`/pesanan/${o.id}`} className={btn}>
                  Lihat detail
                </Link>
              </div>
            </article>
          );
        })}
      </div>
    </div>
  );
}
