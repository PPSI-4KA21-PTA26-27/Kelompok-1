"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { Search } from "lucide-react";
import ProductCard from "@/components/product/ProductCard";
import { useCategories } from "@/hooks/useCategories";
import { cn } from "@/lib/cn";
import { productService } from "@/services";
import type { ApiProduct, ApiMeta } from "@/types/api";

const LIMIT = 8;

function Katalog() {
  const sp = useSearchParams();
  const router = useRouter();
  const cats = useCategories();
  const q = sp.get("q") ?? "";
  const category = sp.get("category") ?? "";
  const page = Math.max(1, Number(sp.get("page")) || 1);
  const [items, setItems] = useState<ApiProduct[]>([]);
  const [meta, setMeta] = useState<ApiMeta>({ page: 1, limit: LIMIT, total: 0 });

  useEffect(() => {
    productService
      .list({ q: q || undefined, category_id: category || undefined, page, limit: LIMIT })
      .then((r) => {
        setItems(r.data);
        setMeta(r.meta);
      })
      .catch(() => setItems([]));
  }, [q, category, page]);

  const go = (next: Record<string, string>) => {
    const p = new URLSearchParams({ ...(q && { q }), ...(category && { category }), ...next });
    for (const [k, v] of [...p]) if (!v || (k === "page" && v === "1")) p.delete(k);
    router.push(`/katalog${p.size ? `?${p}` : ""}`);
  };
  const pages = Math.max(1, Math.ceil(meta.total / LIMIT));
  const total = cats.reduce((n, c) => n + c.count, 0);
  const chip = (on: boolean) =>
    cn("rounded-full border px-6 py-4 text-lg font-bold", on ? "border-ink bg-ink text-white" : "border-line bg-white");
  const box = "grid size-[60px] place-items-center rounded-xl border text-2xl font-bold disabled:opacity-40";

  return (
    <div>
      <p className="text-lg text-muted">
        Beranda <span className="mx-1 text-sm">&gt;</span> <span className="text-ink">Katalog</span>
      </p>
      <h1 className="mt-2 text-4xl font-extrabold lg:text-5xl">Katalog Produk</h1>
      <p className="mt-3 text-lg text-muted">
        Menampilkan {items.length} dari {meta.total} produk
      </p>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          go({ q: String(new FormData(e.currentTarget).get("q") ?? "").trim(), page: "" });
        }}
        key={q}
        className="relative mt-6"
      >
        <Search className="absolute left-4 top-1/2 size-5 -translate-y-1/2 text-faint" />
        <input
          name="q"
          defaultValue={q}
          placeholder="Cari produk..."
          aria-label="Cari produk"
          className="h-12 w-full rounded-xl border border-line bg-white pl-12 pr-4 text-lg"
        />
      </form>

      <div className="mt-6 flex flex-wrap gap-2 sm:gap-4">
        <button className={chip(!category)} onClick={() => go({ category: "", page: "" })}>
          Semua <span className="ml-1 text-muted">{total}</span>
        </button>
        {cats.map((c) => (
          <button key={c.id} className={chip(category === c.id)} onClick={() => go({ category: c.id, page: "" })}>
            {c.name} <span className="ml-1 text-muted">{c.count}</span>
          </button>
        ))}
      </div>

      <div className="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 lg:gap-[22px]">
        {items.map((p) => (
          <ProductCard key={p.id} p={p} category={cats.find((c) => c.id === p.category_id)?.name} />
        ))}
      </div>

      {pages > 1 && (
        <nav className="mt-8 flex justify-center gap-3" aria-label="Halaman">
          <button className={cn(box, "border-line bg-white")} disabled={page <= 1} onClick={() => go({ page: String(page - 1) })} aria-label="Sebelumnya">
            &lt;
          </button>
          {Array.from({ length: pages }, (_, i) => i + 1).map((n) => (
            <button key={n} className={cn(box, n === page ? "border-ink bg-ink text-white" : "border-line bg-white")} onClick={() => go({ page: String(n) })}>
              {n}
            </button>
          ))}
          <button className={cn(box, "border-line bg-white")} disabled={page >= pages} onClick={() => go({ page: String(page + 1) })} aria-label="Berikutnya">
            &gt;
          </button>
        </nav>
      )}
    </div>
  );
}

export default function KatalogPage() {
  return (
    <Suspense>
      <Katalog />
    </Suspense>
  );
}
