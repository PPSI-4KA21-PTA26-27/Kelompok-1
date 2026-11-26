"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import BannerStrip from "@/components/banner/BannerStrip";
import ProductCard from "@/components/product/ProductCard";
import { useCategories } from "@/hooks/useCategories";
import { productService } from "@/services";
import type { ApiProduct } from "@/types/api";

export default function Home() {
  const cats = useCategories();
  const [items, setItems] = useState<ApiProduct[]>([]);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    productService
      .list({ sort: "newest", limit: 6 })
      .then((r) => setItems(r.data))
      .catch(() => setFailed(true));
  }, []);

  return (
    <div className="space-y-6 lg:space-y-10">
      {/* Banner hero diatur Admin pada L-26; tanpa banner aktif tampil hero bawaan. */}
      <BannerStrip
        placement="hero"
        className="h-88.5 sm:h-72.5"
        fallback={
          <div className="flex min-h-88.5 flex-col items-start justify-end gap-4 rounded-2xl bg-ink p-8 text-white sm:min-h-72.5 lg:p-12">
            <Image src="/Logo.png" alt="" width={60} height={60} className="size-14" />
            <h1 className="max-w-xl text-3xl font-extrabold lg:text-5xl">Streetwear lokal, siap dipakai.</h1>
            <p className="font-bold text-lime">cz we are not really strangers</p>
            <Link href="/katalog" className="rounded-full bg-lime px-6 py-3 text-sm font-bold text-ink">
              Lihat katalog
            </Link>
          </div>
        }
      />

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
        {failed && (
          <p role="alert" className="rounded-xl bg-white p-6 text-center font-semibold text-danger">
            Data produk gagal dimuat. Muat ulang halaman.
          </p>
        )}
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 lg:gap-6">
          {items.map((p, i) => (
            // tablet/handphone 6 produk, desktop 4 (sesuai desain)
            <div key={p.id} className={i >= 4 ? "lg:hidden" : undefined}>
              <ProductCard p={p} category={cats.find((c) => c.id === p.category_id)?.name} />
            </div>
          ))}
        </div>
      </section>

      <BannerStrip placement="main" className="aspect-5/4 sm:aspect-[10/3]" />
    </div>
  );
}
