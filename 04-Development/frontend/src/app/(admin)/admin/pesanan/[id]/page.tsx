"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Card, PageHead, Thumb, btnLime, btnLine, inputCls, labelCls } from "@/components/admin/ui";
import { ADMIN_NEXT, CUST_STATUS } from "@/components/order/status";
import { Badge, ErrorState } from "@/components/ui/misc";
import { errorMessage } from "@/lib/api";
import { cn } from "@/lib/cn";
import { formatDate, formatRupiah, hasRTRW, joinAddress, paymentLabel } from "@/lib/format";
import { adminService, orderService } from "@/services";
import type { ApiOrderDetail, OrderStatus } from "@/types/api";

const COVERAGE = [
  { v: "covered", label: "Terjangkau" },
  { v: "not_covered", label: "Tidak terjangkau", hint: "Pengiriman ditunda, tentukan tindak lanjut." },
  { v: "needs_check", label: "Perlu pemeriksaan lebih lanjut", hint: "Pengiriman ditahan sampai jangkauan dipastikan." },
];

export default function AdminPesananDetail() {
  const { id } = useParams<{ id: string }>();
  const [o, setO] = useState<ApiOrderDetail | null>(null);
  const [err, setErr] = useState("");
  const [next, setNext] = useState<OrderStatus | "">("");
  const [resi, setResi] = useState("");
  const [busy, setBusy] = useState(false);

  const load = useCallback(() => {
    orderService
      .get(id)
      .then((d) => {
        setO(d);
        setResi(d.shipment?.tracking_number ?? "");
        setNext("");
      })
      .catch((e) => setErr(errorMessage(e)));
  }, [id]);
  useEffect(load, [load]);

  if (err) return <ErrorState message={err} onRetry={load} />;
  if (!o) return null;

  const run = async (fn: () => Promise<unknown>, ok: string) => {
    setBusy(true);
    try {
      await fn();
      toast.success(ok);
      load();
    } catch (e) {
      toast.error(errorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const saveStatus = () => {
    if (!next) return toast.error("Pilih status baru.");
    if (next === "shipped" && !resi.trim()) return toast.error("Isi nomor resi sebelum menandai Dikirim.");
    if (next === "cancelled" && !confirm("Batalkan pesanan ini? Stok akan dikembalikan.")) return;
    run(
      () => adminService.updateOrderStatus(o.id, { order_status: next, tracking_number: next === "shipped" ? resi.trim() : undefined }),
      "Status disimpan.",
    );
  };

  const shipment = (patch: { tracking_number?: string | null; coverage_status?: string }) =>
    adminService.upsertShipment(o.id, {
      tracking_number: o.shipment?.tracking_number ?? null,
      coverage_status: o.shipment?.coverage_status ?? "covered",
      shipping_status: o.shipment?.shipping_status ?? "pending",
      ...patch,
    });

  const saveResi = () => {
    if (!resi.trim()) return toast.error("Nomor resi wajib diisi.");
    run(() => shipment({ tracking_number: resi.trim() }), "Nomor resi disimpan.");
  };

  const options = ADMIN_NEXT[o.order_status] ?? [];
  const courier = [o.courier_code?.toUpperCase(), o.courier_service].filter(Boolean).join(" · ") || "-";
  const pay = o.payment;
  const payBadge = pay && paymentLabel({ payment_status: pay.status, has_proof: pay.has_proof });

  return (
    <div>
      <PageHead
        title={"Pesanan #" + o.order_number}
        crumb={{ href: "/admin/pesanan", label: "Pesanan Masuk" }}
        sub={formatDate(o.created_at) + " · " + o.recipient_name}
      />
      <div className="grid gap-6 sm:grid-cols-[1fr_262px] xl:grid-cols-[1fr_400px]">
        <div className="min-w-0 space-y-6">
          <Card title="Item pesanan">
            <ul className="space-y-4">
              {o.items.map((it) => (
                <li key={it.id} className="flex items-center gap-4 text-sm">
                  <Thumb src={it.image_url} alt={it.product_name} />
                  <div className="flex-1">
                    <b>{it.product_name}</b>
                    <p className="text-muted">
                      {it.size} x {it.quantity}
                    </p>
                  </div>
                  <b>{formatRupiah(it.subtotal)}</b>
                </li>
              ))}
            </ul>
          </Card>

          <Card title="Penerima & alamat">
            <div className="grid gap-4 text-sm sm:grid-cols-2">
              <div>
                <p className="text-xs text-muted">Penerima</p>
                <b>{o.recipient_name}</b>
                <p>{o.recipient_phone}</p>
              </div>
              <div>
                <p className="text-xs text-muted">Alamat tujuan</p>
                <p>{joinAddress(o)}</p>
                {!hasRTRW(o.shipping_address) && (
                  <p role="alert" className="mt-2 rounded-lg bg-warn-soft px-3 py-2 text-xs text-warn">
                    RT/RW belum tercatat (pesanan lama). Minta pelanggan melengkapi sebelum pengiriman.
                  </p>
                )}
              </div>
            </div>
          </Card>

          <Card title="Cek ketersediaan pengiriman">
            <p className="-mt-2 mb-3 text-sm text-muted">
              Tujuan: {o.shipping_city ?? "-"} · {courier}
            </p>
            <fieldset className="space-y-3" disabled={busy}>
              <legend className="sr-only">Status jangkauan pengiriman</legend>
              {COVERAGE.map((c) => {
                const on = o.shipment?.coverage_status === c.v;
                return (
                  <label key={c.v} className={cn("flex cursor-pointer gap-3 rounded-lg border px-3 py-2 text-sm", on ? "border-ink bg-lime/30" : "border-line-strong")}>
                    <input
                      type="radio"
                      name="coverage"
                      checked={on}
                      onChange={() => run(() => shipment({ coverage_status: c.v }), "Status jangkauan disimpan.")}
                      className="mt-0.5 accent-ink"
                    />
                    <span>
                      <b>{c.label}</b>
                      {c.hint && <span className="block text-[10px] text-muted">{c.hint}</span>}
                    </span>
                  </label>
                );
              })}
            </fieldset>
          </Card>
        </div>

        <div className="space-y-6">
          <Card title="Ubah status">
            <p className="mb-3 text-sm">
              Status saat ini:{" "}
              <span className={cn("rounded-full px-2 py-0.5 text-xs font-bold", CUST_STATUS[o.order_status]?.cls)}>{CUST_STATUS[o.order_status]?.label}</span>
            </p>
            {options.length ? (
              <>
                <label htmlFor="status" className={labelCls}>
                  Status pesanan
                </label>
                <select id="status" value={next} onChange={(e) => setNext(e.target.value as OrderStatus)} className={inputCls}>
                  <option value="">Pilih status</option>
                  {options.map((s) => (
                    <option key={s} value={s}>
                      {s === "processing" ? "Dikemas" : CUST_STATUS[s].label}
                    </option>
                  ))}
                </select>
                <button onClick={saveStatus} disabled={busy} className="mt-3 w-full rounded-lg bg-ink py-2 text-sm font-bold text-white disabled:opacity-50">
                  Simpan Status
                </button>
              </>
            ) : (
              <p className="text-sm text-muted">Tidak ada perubahan status yang tersedia.</p>
            )}
          </Card>

          <Card title="Nomor resi">
            <p className="-mt-2 mb-3 text-xs text-muted">Ekspedisi: {courier}</p>
            <label htmlFor="resi" className={labelCls}>
              Nomor resi
            </label>
            <input id="resi" value={resi} onChange={(e) => setResi(e.target.value)} className={inputCls} />
            <button onClick={saveResi} disabled={busy} className={btnLime + " mt-3 w-full"}>
              Simpan Resi
            </button>
          </Card>

          <Card title="Pembayaran">
            <dl className="space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-muted">Subtotal</dt>
                <dd>{formatRupiah(o.subtotal)}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-muted">Ongkos kirim</dt>
                <dd>{formatRupiah(o.shipping_cost)}</dd>
              </div>
              {pay && (
                <div className="flex justify-between">
                  <dt className="text-muted">Kode unik</dt>
                  <dd>{formatRupiah(pay.unique_code)}</dd>
                </div>
              )}
              <div className="flex justify-between">
                <dt className="text-muted">Total</dt>
                <dd className="font-bold">{formatRupiah(pay?.amount ?? o.total_amount)}</dd>
              </div>
              <div className="flex items-center justify-between">
                <dt className="text-muted">{pay?.method?.toUpperCase() ?? "Belum ada tagihan"}</dt>
                <dd>{payBadge && <Badge tone={payBadge.tone} className="px-2 py-0.5 text-[10px]">{payBadge.label}</Badge>}</dd>
              </div>
            </dl>
            {pay?.has_proof && (
              <Link href={"/admin/pembayaran?id=" + pay.id} className={btnLine + " mt-4 block text-center text-sm"}>
                Lihat Validasi
              </Link>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
}
