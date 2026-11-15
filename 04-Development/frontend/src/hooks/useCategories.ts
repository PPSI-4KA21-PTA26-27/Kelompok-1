"use client";

import { useEffect, useState } from "react";
import { categoryService, productService } from "@/services";
import type { ApiCategory } from "@/types/api";

export type CategoryCount = ApiCategory & { count: number };

/** Kategori aktif + jumlah produk aktifnya (total dari meta, limit 1). */
export function useCategories() {
  const [cats, setCats] = useState<CategoryCount[]>([]);
  useEffect(() => {
    let off = false;
    categoryService
      .list()
      .then(async (r) => {
        const act = r.data.filter((c) => c.is_active);
        const all = await Promise.all(
          act.map((c) => productService.list({ category_id: c.id, limit: 1 }).then((p) => ({ ...c, count: p.meta.total }))),
        );
        if (!off) setCats(all);
      })
      .catch(() => {});
    return () => {
      off = true;
    };
  }, []);
  return cats;
}
