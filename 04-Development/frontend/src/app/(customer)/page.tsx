"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import ProductCard from "@/components/product/ProductCard";
import { useCategories } from "@/hooks/useCategories";
import { cdn } from "@/lib/img";
import { bannerService, productService } from "@/services";
import type { ApiBanner, ApiProduct } from "@/types/api";

export default function Home() {
  const cats = useCategories();
  // undefined = memuat, null = tidak ada banner aktif
  const [banner, setBanner] = useState<ApiBanner | null>();
  const [items, setItems] = useState<ApiProduct[]>([]);

  useEffect(() => {
    bannerService.live("hero").then((b) => setBanner(b[0] ?? null)).catch(() => setBanner(null));
    productService.list({ sort: "newest", limit: 6 }).then((r) => setItems(r.data)).catch(() => {});
  }, []);

  return (
    <div className="space-y-6 lg:space-y-10">
      {banner === null ? (
        <div className="flex h-[354px] flex-col items-start justify-end gap-4 rounded-2xl bg-ink p-8 text-white sm:h-[290px] lg:p-12">
          <Image src="/Logo.png" alt="" width={68} height={60} className="h-14 w-auto" />
          <h1 className="max-w-xl text-3xl font-extrabold lg:text-5xl">Streetwear lokal, siap dipakai.</h1>
          <p className="font-bold text-lime">cz we are not really strangers</p>
          <Link href="/katalog" className="rounded-full bg-lime px-6 py-3 text-sm font-bold text-ink">
            Lihat katalog
          </Link>
        </div>
      ) : (
        <div className="h-[354px] overflow-hidden rounded-2xl bg-[#d9d9d9] sm:h-[290px]">
          {banner && (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={cdn(banner.image_url, 1400)} alt={banner.title ?? "Banner"} className="size-full object-cover" />
          )}
        </div>
      )}

      <div className="grid grid-cols-3 gap-3 sm:gap-4 lg:gap-8">
        {cats.map((c) => (
          <Link
            key={c.id}
            href={`/katalog?category=${c.id}`}
            className="flex flex-col items-start justify-between gap-1 rounded-2xl border border-line bg-white px-3 py-3 text-sm font-bold sm:flex-row sm:items-center sm:px-5 sm:py-5 sm:text-base lg:px-9 lg:py-6 lg:text-xl"
          >
            {c.name}
            <span className="font-semibold text-muted">{c.count} produk</span>
          </Link>
        ))}
      </div>

      <section>
        <div className="mb-4 flex items-end justify-between">
          <h2 className="text-2xl font-bold">Produk terbaru</h2>
          <Link href="/katalog" className="font-semibold underline underline-offset-4">
            Lihat semua
          </Link>
        </div>
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 lg:gap-6">
          {items.map((p, i) => (
            // tablet/handphone 6 produk, desktop 4 (sesuai desain)
            <div key={p.id} className={i >= 4 ? "lg:hidden" : undefined}>
              <ProductCard p={p} category={cats.find((c) => c.id === p.category_id)?.name} />
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
