import Link from "next/link";
import { formatRupiah, finalPrice } from "@/lib/format";
import { cdn } from "@/lib/img";
import type { ApiProduct } from "@/types/api";

export default function ProductCard({ p, category }: { p: ApiProduct; category?: string }) {
  return (
    <Link href={`/produk/${p.id}`} className="group block overflow-hidden rounded-2xl bg-white">
      <div className="relative aspect-square bg-[#d9d9d9]">
        {p.primary_image_url && (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={cdn(p.primary_image_url, 600)} alt={p.name} loading="lazy" className="size-full object-cover" />
        )}
        {p.availability_type === "preorder" && (
          <span className="absolute left-2.5 top-2.5 sm:left-5 sm:top-5 rounded-full bg-ink px-2.5 py-1 text-xs font-bold text-white">
            Pre-Order
          </span>
        )}
      </div>
      <div className="space-y-2 p-3 sm:space-y-3 sm:p-4.5">
        {category && <p className="text-sm text-muted">{category}</p>}
        <h3 className="line-clamp-2 min-h-11 text-sm font-bold leading-snug text-ink">{p.name}</h3>
        <p className="text-sm font-bold text-ink">{formatRupiah(finalPrice(p.price, p.discount_percent))}</p>
      </div>
    </Link>
  );
}
