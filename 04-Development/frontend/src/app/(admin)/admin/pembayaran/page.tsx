"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useCallback, useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Check, Info, X } from "lucide-react";
import { Card, PageHead, btnLime, btnDanger } from "@/components/admin/ui";
import { Badge, EmptyState } from "@/components/ui/misc";
import { errorMessage } from "@/lib/api";
import { cn } from "@/lib/cn";
import { formatDate, formatRupiah } from "@/lib/format";
import { adminService, paymentService } from "@/services";
import type { ApiAdminPayment } from "@/types/api";

export default function PembayaranPage() {
  return (
    <Suspense>
      <Pembayaran />
    </Suspense>
  );
}

function Pembayaran() {
  const [list, setList] = useState<ApiAdminPayment[] | null>(null);
  const initial = useSearchParams().get("id");
  const [sel, setSel] = useState<string | null>(initial);
  const [proof, setProof] = useState<{ id: string; url: string; image: boolean } | null>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(() => {
    // shortcut: 100 antrean pertama; tambah paginasi bila antrean lebih panjang
    adminService
      .payments({ status: "menunggu", limit: 100 })
      .then((r) => setList(r.data.filter((p) => p.has_proof)))
      .catch((e) => {
        setList([]);
        toast.error(errorMessage(e));
      });
  }, []);
  useEffect(load, [load]);

  const cur = list?.find((p) => p.id === sel) ?? null;

  useEffect(() => {
    if (!cur) return;
    let url = "";
    paymentService
      .proof(cur.order_id)
      .then((b) => {
        url = URL.createObjectURL(b);
        setProof({ id: cur.id, url, image: b.type.startsWith("image/") });
      })
      .catch((e) => toast.error(errorMessage(e)));
    return () => {
      if (url) URL.revokeObjectURL(url);
    };
  }, [cur]);

  const review = async (status: "valid" | "ditolak") => {
    if (!cur) return;
    if (status === "ditolak" && !confirm("Tolak pembayaran ini?")) return;
    setBusy(true);
    try {
      await adminService.reviewPayment(cur.id, status);
      toast.success(status === "valid" ? "Pembayaran divalidasi." : "Pembayaran ditolak.");
      setSel(null);
      load();
    } catch (e) {
      toast.error(errorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const total = cur ? Number(cur.order_total) : 0;
  const shownProof = proof?.id === cur?.id ? proof : null;

  return (
    <div>
      <PageHead title="Validasi Pembayaran" sub={(list?.length ?? 0) + " pesanan menunggu verifikasi"} />
      <div className="grid gap-8 xl:grid-cols-[397px_1fr]">
        <Card title="Menunggu verifikasi" className="self-start">
          {list?.length === 0 && <p className="text-sm text-muted">Tidak ada bukti yang perlu diperiksa.</p>}
          <ul className="space-y-3">
            {list?.map((p) => (
              <li key={p.id}>
                <button
                  onClick={() => setSel(p.id)}
                  aria-current={p.id === sel ? "true" : undefined}
                  className={cn("w-full rounded-lg border px-4 py-3 text-left text-sm", p.id === sel ? "border-ink bg-lime/30" : "border-line-strong")}
                >
                  <span className="flex justify-between font-bold">
                    <span>#{p.order_number}</span>
                    <span>{formatRupiah(p.amount)}</span>
                  </span>
                  <span className="text-muted">
                    {p.customer_name} · {formatDate(p.paid_at ?? p.created_at, true)}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </Card>

        {cur ? (
          <Card
            title={"Pesanan #" + cur.order_number}
            action={
              <Badge tone="warn" className="px-2 py-0.5 text-[10px]">
                Menunggu verifikasi
              </Badge>
            }
          >
            <p className="-mt-3 mb-5 text-sm text-muted">
              {cur.customer_name} · bukti diunggah {formatDate(cur.paid_at ?? cur.created_at, true)} ·{" "}
              <Link href={"/admin/pesanan/" + cur.order_id} className="underline">
                lihat pesanan
              </Link>
            </p>
            <div className="grid gap-6 sm:grid-cols-2">
              <figure>
                <figcaption className="mb-2 text-center text-sm text-muted">Bukti unggahan pelanggan</figcaption>
                <div className="grid min-h-48 place-items-center overflow-hidden rounded-lg bg-white shadow">
                  {!shownProof ? (
                    <span className="text-xs text-muted">Memuat bukti...</span>
                  ) : shownProof.image ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={shownProof.url} alt={"Bukti bayar pesanan #" + cur.order_number} className="max-h-[480px] w-full object-contain" />
                  ) : (
                    <a href={shownProof.url} target="_blank" rel="noreferrer" className="text-sm font-bold underline">
                      Buka berkas bukti
                    </a>
                  )}
                </div>
              </figure>
              <dl className="space-y-3 text-sm">
                {[
                  ["Total pesanan", formatRupiah(total)],
                  ["Kode unik", String(cur.unique_code)],
                  ["Nominal wajib dibayar", formatRupiah(cur.amount)],
                ].map(([k, v]) => (
                  <div key={k} className="flex justify-between">
                    <dt className="text-muted">{k}</dt>
                    <dd className="font-bold">{v}</dd>
                  </div>
                ))}
                <p className="flex gap-2 rounded-lg bg-stone px-3 py-2 text-xs text-ink-3">
                  <Info className="size-3 shrink-0" aria-hidden /> Cocokkan nominal dan kode unik dengan notifikasi pembayaran toko.
                </p>
              </dl>
            </div>
            <div className="mt-6 flex flex-wrap items-center gap-3">
              <button onClick={() => review("valid")} disabled={busy} className={btnLime + " flex items-center gap-2"}>
                <Check className="size-4" aria-hidden /> Validasi Pembayaran
              </button>
              <button onClick={() => review("ditolak")} disabled={busy} className={btnDanger + " flex items-center gap-2 text-sm"}>
                <X className="size-4" aria-hidden /> Tolak Pembayaran
              </button>
              <button onClick={() => setSel(null)} className="px-3 text-sm font-bold underline">
                Tunda
              </button>
            </div>
            <p className="mt-3 text-xs text-muted">Pembayaran valid mengubah status pesanan menjadi Perlu dikirim.</p>
          </Card>
        ) : (
          <Card>
            <EmptyState title="Pilih pembayaran" description="Pilih pesanan di sebelah kiri untuk memeriksa bukti bayar." />
          </Card>
        )}
      </div>
    </div>
  );
}
