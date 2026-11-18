"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { Clock, Minus, Plus, ShoppingCart, SquareArrowRight } from "lucide-react";
import { useCartActions } from "@/hooks/useCartActions";
import { useCategories } from "@/hooks/useCategories";
import { Breadcrumb, Skeleton } from "@/components/ui/misc";
import { cn } from "@/lib/cn";
import { finalPrice, formatRupiah } from "@/lib/format";
import { cdn } from "@/lib/img";
import { productService } from "@/services";
import type { ApiProductDetail } from "@/types/api";

export default function ProdukPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const cats = useCategories();
  const { add } = useCartActions();
  const [p, setP] = useState<ApiProductDetail | null>(null);
  const [missing, setMissing] = useState(false);
  const [size, setSize] = useState("");
  const [color, setColor] = useState("");
  const [qty, setQty] = useState(1);
  const [img, setImg] = useState(0);

  useEffect(() => {
    productService.get(id).then(setP).catch(() => setMissing(true));
  }, [id]);

  const sizes = useMemo(() => [...new Set(p?.variants.map((v) => v.size))], [p]);
  const sizeStock = (s: string) => p?.variants.filter((v) => v.size === s).reduce((n, v) => n + v.stock, 0) ?? 0;
  const activeSize = size || sizes.find((s) => sizeStock(s) > 0) || sizes[0] || "";
  const colors = useMemo(
    () => [...new Set(p?.variants.filter((v) => v.size === activeSize).map((v) => v.color).filter(Boolean))] as string[],
    [p, activeSize],
  );
  const activeColor = colors.includes(color) ? color : (colors.find((c) => p?.variants.some((v) => v.size === activeSize && v.color === c && v.stock > 0)) ?? colors[0] ?? "");
  const variant = p?.variants.find((v) => v.size === activeSize && (v.color ?? "") === activeColor);

  if (missing) return <p className="py-20 text-center text-lg text-muted">Produk tidak ditemukan.</p>;
  if (!p)
    return (
      <div role="status" className="grid gap-6 sm:grid-cols-2 sm:gap-7 lg:gap-15">
        <span className="sr-only">Memuat produk...</span>
        <Skeleton className="aspect-square rounded-2xl" />
        <div className="space-y-4">
          <Skeleton className="h-10 w-3/4" />
          <Skeleton className="h-6 w-1/3" />
          <Skeleton className="h-32" />
          <Skeleton className="h-12" />
        </div>
      </div>
    );

  const catName = cats.find((c) => c.id === p.category_id)?.name;
  const images = [...p.images].sort((a, b) => a.sort_order - b.sort_order);
  const price = finalPrice(p.price, p.discount_percent);
  const po = p.availability_type === "preorder";
  const max = variant?.stock ?? 0;
  const stockNote = sizes.map((s) => (sizeStock(s) > 0 ? `${s}:${sizeStock(s)}` : `${s} habis`)).join(" · ");
  const addToCart = async (go: string) => {
    if (variant && (await add(variant.id, qty, `/produk/${p.id}`))) router.push(go);
  };
  const opt = (on: boolean, off: boolean) =>
    cn(
      "rounded-xl border px-5 py-3 text-lg font-bold",
      on ? "border-ink bg-ink text-white" : "border-line bg-white",
      off && "border-dashed text-faint line-through",
    );

  return (
    <div>
      <Breadcrumb
        items={[
          { href: "/", label: "Beranda" },
          { href: "/katalog", label: "Katalog" },
          ...(catName ? [{ href: "/katalog?category=" + p.category_id, label: catName }] : []),
          { label: "Detail Produk" },
        ]}
      />

      <div className="mt-4 grid gap-6 sm:grid-cols-2 sm:gap-7 lg:gap-15">
        <div>
          <div className="aspect-square overflow-hidden rounded-2xl bg-[#d9d9d9]">
            {images[img] && (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={cdn(images[img].image_url, 900)} alt={p.name} className="size-full object-cover" />
            )}
          </div>
          <div className="mt-4 grid grid-cols-4 gap-3 sm:gap-4">
            {images.slice(0, 4).map((im, i) => (
              <button
                key={im.id}
                onClick={() => setImg(i)}
                aria-label={`Foto ${i + 1}`}
                className={cn("aspect-square overflow-hidden rounded-xl bg-[#d9d9d9]", i === img && "outline-2 outline-ink")}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={cdn(im.image_url, 300)} alt="" className="size-full object-cover" />
              </button>
            ))}
          </div>
        </div>

        <div>
          {catName && <p className="text-xl text-muted">{catName}</p>}
          <h1 className="text-3xl font-extrabold sm:text-4xl lg:text-5xl">{p.name}</h1>
          <div className="mt-4 flex flex-wrap items-center gap-4 lg:mt-8">
            <p className="text-2xl font-extrabold sm:text-3xl lg:text-5xl">{formatRupiah(price)}</p>
            {p.discount_percent > 0 && <s className="text-xl text-muted">{formatRupiah(p.price)}</s>}
            {po && <span className="rounded-full bg-ink px-3 py-1.5 text-sm font-bold text-white">Pre-Order</span>}
          </div>
          {po && (
            <p className="mt-4 flex items-center gap-3 rounded-xl bg-warn-soft px-4 py-3 text-warn">
              <Clock className="size-4 shrink-0" />
              <span>
                <b>Estimasi pre-order {p.po_estimate_days ?? "-"} hari.</b> Produk akan dikirim setelah masa pengerjaan selesai.
              </span>
            </p>
          )}

          <h2 className="mt-8 text-lg font-semibold">Ukuran</h2>
          <div className="mt-2 flex flex-wrap gap-3">
            {sizes.map((s) => (
              <button key={s} disabled={!sizeStock(s)} onClick={() => setSize(s)} className={opt(s === activeSize, !sizeStock(s))}>
                {s}
              </button>
            ))}
          </div>
          <p className="mt-2 text-xs text-muted">Stok ukuran {stockNote}</p>

          {colors.length > 0 && (
            <>
              <h2 className="mt-6 text-lg font-semibold">Warna</h2>
              <div className="mt-2 flex flex-wrap gap-3">
                {colors.map((c) => {
                  const out = !p.variants.some((v) => v.size === activeSize && v.color === c && v.stock > 0);
                  return (
                    <button key={c} disabled={out} onClick={() => setColor(c)} className={opt(c === activeColor, out)}>
                      {c}
                    </button>
                  );
                })}
              </div>
            </>
          )}

          <div className="mt-6 flex flex-wrap gap-3">
            <div className="flex h-14 items-center rounded-xl border border-line bg-white">
              <button aria-label="Kurangi" className="h-full px-5 disabled:opacity-40" disabled={qty <= 1} onClick={() => setQty(qty - 1)}>
                <Minus className="size-3" />
              </button>
              <span className="w-8 text-center text-lg font-bold">{qty}</span>
              <button aria-label="Tambah" className="h-full px-5 disabled:opacity-40" disabled={qty >= max} onClick={() => setQty(qty + 1)}>
                <Plus className="size-3" />
              </button>
            </div>
            <button
              disabled={!variant || max < 1}
              onClick={() => addToCart("/keranjang")}
              className="flex h-14 flex-1 items-center justify-center gap-2 whitespace-nowrap rounded-xl bg-ink px-3 text-sm sm:px-6 sm:text-lg font-bold text-lime disabled:opacity-50"
            >
              <ShoppingCart className="size-5" /> Tambah ke keranjang
            </button>
            <button
              disabled={!variant || max < 1}
              onClick={() => addToCart("/checkout")}
              className="flex h-14 flex-1 items-center justify-center gap-2 whitespace-nowrap rounded-xl bg-lime px-3 text-sm sm:px-6 sm:text-lg font-bold disabled:opacity-50"
            >
              <SquareArrowRight className="size-5" /> Beli sekarang
            </button>
          </div>

          <div className="mt-6 border-t border-line pt-5">
            <h2 className="text-xl font-semibold">Deskripsi</h2>
            <p className="mt-3 whitespace-pre-line text-xl text-muted">{p.description}</p>
          </div>
          <dl className="mt-6 grid grid-cols-2 gap-4 border-t border-line pt-5 text-xs">
            <div>
              <dt className="text-muted">Berat</dt>
              <dd className="mt-1 font-semibold">{p.weight_gram} g</dd>
            </div>
            <div>
              <dt className="text-muted">Ketersediaan</dt>
              <dd className="mt-1 font-semibold">{po ? `Pre-order · ${p.po_estimate_days ?? "-"} hari` : "Ready stock"}</dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  );
}
