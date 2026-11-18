"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { Minus, Plus, Trash2, TriangleAlert, ChevronRight } from "lucide-react";
import { useSelector } from "react-redux";
import { useAuth } from "@/app/context/AuthContext";
import Steps from "@/components/checkout/Steps";
import { useCartActions } from "@/hooks/useCartActions";
import { formatRupiah } from "@/lib/format";
import { cdn } from "@/lib/img";
import { selectCart } from "@/redux/features/cart-slice";

export default function KeranjangPage() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const { items, subtotal, loaded } = useSelector(selectCart);
  const { updateQuantity, removeItem } = useCartActions();

  useEffect(() => {
    if (!loading && !user) router.replace("/masuk?next=/keranjang");
  }, [loading, user, router]);

  const over = items.some((i) => i.quantity > i.stock);
  const btn = "grid size-9 place-items-center disabled:opacity-40";

  return (
    <div>
      <Steps current={1} />
      <h1 className="mt-4 text-4xl font-extrabold lg:mt-12 lg:text-5xl">Keranjang</h1>
      <p className="mt-2 text-xl text-muted">{items.length} produk dalam keranjang</p>

      {loaded && items.length === 0 ? (
        <p className="mt-12 text-lg text-muted">
          Keranjang masih kosong.{" "}
          <Link href="/katalog" className="font-bold text-ink underline underline-offset-4">
            Mulai belanja
          </Link>
        </p>
      ) : (
        <div className="mt-6 grid items-start gap-4 sm:grid-cols-[1fr_245px] lg:mt-8 lg:grid-cols-[1fr_424px] lg:gap-8">
          <ul className="divide-y divide-line rounded-3xl border border-line bg-white px-3 py-2 sm:px-6 sm:py-4 lg:px-7">
            {items.map((i) => (
              <li key={i.id} className="grid grid-cols-[auto_1fr_auto] items-center gap-x-3 gap-y-3 sm:grid-cols-[auto_1fr_auto_auto_auto] py-4 sm:gap-x-3 lg:gap-x-6 lg:py-6">
                <div className="row-span-2 size-12.5 shrink-0 sm:row-span-1 sm:size-14.5 lg:size-30.5 overflow-hidden rounded-2xl bg-[#d9d9d9]">
                  {i.primary_image_url && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={cdn(i.primary_image_url, 250)} alt={i.product_name} className="size-full object-cover" />
                  )}
                </div>
                <div>
                  <p className="font-semibold lg:text-xl">{i.product_name}</p>
                  <p className="mt-1 text-sm text-muted">{[i.size, i.color].filter(Boolean).join(" · ")}</p>
                  <p className="mt-2 text-sm font-bold">{formatRupiah(i.unit_price)}</p>
                  {i.quantity > i.stock && (
                    <p role="alert" className="mt-3 flex items-center gap-2 rounded-xl bg-warn-soft px-3 py-2 text-xs text-warn">
                      <TriangleAlert className="size-3" /> Jumlah melebihi stok! Tersedia maks. {i.stock}.
                    </p>
                  )}
                </div>
                <div className="col-start-2 row-start-2 flex w-fit items-center rounded-lg border border-line sm:col-start-auto sm:row-start-auto">
                  <button aria-label="Kurangi" className={btn} disabled={i.quantity <= 1} onClick={() => updateQuantity(i.id, i.quantity - 1)}>
                    <Minus className="size-3" />
                  </button>
                  <span className="w-6 text-center font-bold">{i.quantity}</span>
                  <button aria-label="Tambah" className={btn} disabled={i.quantity >= i.stock} onClick={() => updateQuantity(i.id, i.quantity + 1)}>
                    <Plus className="size-3" />
                  </button>
                </div>
                <p className="col-start-3 row-start-2 whitespace-nowrap text-right text-sm font-bold sm:col-start-auto sm:row-start-auto sm:w-20 lg:w-28 lg:text-lg">{formatRupiah(i.subtotal)}</p>
                <button aria-label={`Hapus ${i.product_name}`} className="col-start-3 row-start-1 -m-2 self-start justify-self-end p-2 text-danger sm:col-start-auto sm:row-start-auto sm:self-center" onClick={() => removeItem(i.id)}>
                  <Trash2 className="size-5" />
                </button>
              </li>
            ))}
          </ul>

          <aside className="rounded-3xl border border-line bg-white p-7">
            <h2 className="text-lg font-semibold">Ringkasan</h2>
            <dl className="mt-6 space-y-3 text-muted">
              <div className="flex justify-between">
                <dt>Subtotal</dt>
                <dd className="font-bold text-ink">{formatRupiah(subtotal)}</dd>
              </div>
              <div className="flex justify-between gap-4">
                <dt>Ongkos kirim</dt>
                <dd>Dihitung di langkah berikutnya</dd>
              </div>
            </dl>
            <div className="mt-3 flex justify-between border-t border-line pt-3 text-lg font-bold">
              <span>Total</span>
              <span>{formatRupiah(subtotal)}</span>
            </div>
            <button
              disabled={over || items.length === 0}
              onClick={() => router.push("/checkout")}
              className="mt-4 flex h-12 w-full items-center justify-center gap-1.5 rounded-xl bg-lime text-sm font-bold disabled:opacity-50"
            >
              <ChevronRight className="size-4" /> Lanjut ke Pemesanan
            </button>
          </aside>
        </div>
      )}
    </div>
  );
}
