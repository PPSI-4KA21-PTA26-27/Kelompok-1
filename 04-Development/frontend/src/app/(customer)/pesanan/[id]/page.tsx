"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Check, MapPin, Truck } from "lucide-react";
import { CUST_STATUS } from "@/components/order/status";
import { CopyButton } from "@/components/ui/misc";
import { cn } from "@/lib/cn";
import { formatDate, formatRupiah, joinAddress } from "@/lib/format";
import { cdn } from "@/lib/img";
import { orderService } from "@/services";
import type { ApiOrderDetail } from "@/types/api";

const card = "rounded-3xl border border-line bg-white p-6";
const PAY = {
  menunggu: ["Menunggu", "bg-warn-soft text-warn"],
  valid: ["Valid", "bg-lime text-ink"],
  ditolak: ["Ditolak", "bg-danger-soft text-danger"],
} as const;

export default function PesananDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [o, setO] = useState<ApiOrderDetail | null>(null);
  const [err, setErr] = useState("");

  useEffect(() => {
    orderService.get(id).then(setO).catch((x) => setErr(x instanceof Error ? x.message : "Gagal memuat pesanan"));
  }, [id]);

  if (err) return <p className="py-20 text-center text-lg text-danger">{err}</p>;
  if (!o) return <div className="h-96 animate-pulse rounded-3xl bg-[#d9d9d9]" />;

  const cancelled = o.order_status === "cancelled";
  const at = { paid: ["completed", "shipped", "processing", "paid"], ship: ["completed", "shipped"] };
  const idx = ({ pending_payment: 0, paid: 1, processing: 1, shipped: 2, completed: 3, cancelled: 0 } as const)[o.order_status];
  const steps = [
    { t: "Belum dibayar", d: formatDate(o.created_at) },
    { t: "Perlu dikirim", d: at.paid.includes(o.order_status) && o.payment?.paid_at ? `Pembayaran valid · ${formatDate(o.payment.paid_at)}` : "Menunggu pembayaran valid" },
    { t: "Dikirim", d: at.ship.includes(o.order_status) && o.shipment ? formatDate(o.shipment.updated_at) : "Belum dikirim" },
    { t: "Selesai", d: o.order_status === "completed" ? formatDate(o.updated_at) : "Menunggu pesanan diterima" },
  ];
  const pay = o.payment ? PAY[o.payment.status as keyof typeof PAY] : null;

  return (
    <div>
      <p className="text-sm text-muted lg:text-lg">
        <Link href="/">Beranda</Link> <span className="mx-1 text-sm">&gt;</span> <Link href="/pesanan">Pesanan Saya</Link> &gt;{" "}
        <span className="text-ink">Pesanan #{o.order_number}</span>
      </p>
      <div className="mt-2 flex flex-wrap items-center gap-4">
        <h1 className="text-4xl font-extrabold lg:text-5xl">Pesanan #{o.order_number}</h1>
        {cancelled && <span className={cn("rounded-full px-2.5 py-1 text-xs font-bold", CUST_STATUS.cancelled.cls)}>Dibatalkan</span>}
      </div>
      <p className="mt-2 text-muted">Dibuat {formatDate(o.created_at)}</p>

      <div className="mt-6 grid items-start gap-4 sm:grid-cols-2 lg:mt-8 lg:grid-cols-[1fr_452px] lg:gap-8">
        <div className="space-y-5">
          <section className={card}>
            <h2 className="text-xl font-semibold">Status pesanan</h2>
            <ol className="mt-5">
              {steps.map((s, i) => {
                const done = !cancelled && i < idx;
                const cur = !cancelled && i === idx;
                return (
                  <li key={s.t} className="relative flex gap-4 pb-5 last:pb-0">
                    {i < steps.length - 1 && <span className="absolute left-3 top-6 h-full w-px bg-line" />}
                    <span className={cn("relative grid size-6 shrink-0 place-items-center rounded-full", done ? "bg-ink text-lime" : cur ? "bg-lime text-ink" : "bg-[#d9d9d9]")}>
                      {(done || cur) && <Check className="size-3.5" />}
                    </span>
                    <div>
                      <p className="text-xs font-bold">{s.t}</p>
                      <p className="text-[10px] text-muted">{s.d}</p>
                    </div>
                  </li>
                );
              })}
            </ol>
            {cancelled && o.cancel_reason && <p className="mt-4 text-sm text-danger">Alasan pembatalan: {o.cancel_reason}</p>}
          </section>

          {o.shipment && (
            <section className={card}>
              <h2 className="text-xl font-semibold">Pengiriman</h2>
              <div className="mt-4 flex items-center gap-4">
                <Truck className="size-8" />
                <div className="flex-1 text-xs">
                  <p className="font-bold">{o.shipment.courier_name}</p>
                  <p className="text-muted">Nomor resi</p>
                </div>
                {o.shipment.tracking_number && (
                  <>
                    <span className="text-xs font-bold">{o.shipment.tracking_number}</span>
                    <CopyButton value={o.shipment.tracking_number} />
                  </>
                )}
              </div>
            </section>
          )}

          <section className={card}>
            <h2 className="text-xl font-semibold">Produk</h2>
            <ul className="mt-4 space-y-4">
              {o.items.map((i) => (
                <li key={i.id} className="flex items-center gap-4">
                  <div className="size-[50px] shrink-0 overflow-hidden rounded-lg bg-[#d9d9d9]">
                    {i.image_url && (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={cdn(i.image_url, 100)} alt="" className="size-full object-cover" />
                    )}
                  </div>
                  <div className="flex-1 text-xs">
                    <p className="font-bold">{i.product_name}</p>
                    <p className="text-muted">{i.size} x {i.quantity}</p>
                  </div>
                  <p className="text-xs font-bold">{formatRupiah(i.subtotal)}</p>
                </li>
              ))}
            </ul>
          </section>
        </div>

        <div className="space-y-5">
          <section className={card}>
            <h2 className="text-xl font-semibold">Rincian pembayaran</h2>
            <dl className="mt-5 space-y-3 text-muted">
              <div className="flex justify-between"><dt>Subtotal</dt><dd>{formatRupiah(o.subtotal)}</dd></div>
              <div className="flex justify-between"><dt>Ongkos kirim</dt><dd>{formatRupiah(o.shipping_cost)}</dd></div>
            </dl>
            <div className="mt-3 flex justify-between border-t border-line pt-3 text-lg font-bold">
              <span>Total</span><span>{formatRupiah(o.total_amount)}</span>
            </div>
            {pay && (
              <div className="mt-4 flex items-center justify-between text-muted">
                <span>Pembayaran QRIS</span>
                <span className={cn("rounded-full px-3 py-1 text-sm font-bold", pay[1])}>{pay[0]}</span>
              </div>
            )}
            {o.order_status === "pending_payment" && (
              <Link href={`/pembayaran/${o.id}`} className="mt-4 block rounded-lg bg-lime py-2 text-center text-xs font-bold">
                Bayar sekarang
              </Link>
            )}
          </section>
          <section className={card}>
            <h2 className="text-xl font-semibold">Alamat pengiriman</h2>
            <div className="mt-4 flex gap-2 text-xs">
              <MapPin className="mt-0.5 size-3 shrink-0" />
              <div>
                <p className="font-bold">{o.recipient_name}</p>
                <p className="mt-1 text-muted">{o.recipient_phone}</p>
                <p className="mt-2 text-muted">{joinAddress(o)}</p>
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
