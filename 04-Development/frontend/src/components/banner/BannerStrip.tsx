"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { cn } from "@/lib/cn";
import { cdn } from "@/lib/img";
import { isSafeLink } from "@/lib/safeLink";
import { bannerService } from "@/services";
import type { ApiBanner, BannerPlacement } from "@/types/api";

/**
 * Banner yang sedang tayang pada satu penempatan, urut sort_order (L-06, L-26).
 * Bila tidak ada banner, `fallback` yang tampil (atau tidak ada apa-apa).
 */
export default function BannerStrip({
  placement,
  fallback = null,
  className,
}: {
  placement: BannerPlacement;
  fallback?: React.ReactNode;
  className?: string;
}) {
  // null = sedang memuat
  const [list, setList] = useState<ApiBanner[] | null>(null);

  useEffect(() => {
    bannerService
      .live(placement)
      .then(setList)
      .catch(() => setList([]));
  }, [placement]);

  if (list === null) return fallback ? <div aria-busy className={cn("skeleton rounded-2xl", className)} /> : null;
  if (list.length === 0) return <>{fallback}</>;

  return (
    <div
      role="region"
      aria-label="Banner"
      className={cn("flex snap-x snap-mandatory gap-4 overflow-x-auto rounded-2xl", className)}
    >
      {list.map((b) => {
        const alt = b.title ?? "Banner Hii Stranger";
        const img = (
          <picture>
            {b.mobile_image_url && <source media="(max-width: 639px)" srcSet={cdn(b.mobile_image_url, 700)} />}
            <img src={cdn(b.image_url, 1400)} alt={alt} className="size-full object-cover" />
          </picture>
        );
        const box = "block h-full w-full shrink-0 snap-start overflow-hidden rounded-2xl bg-[#d9d9d9]";
        // tautan tak aman (mis. javascript:) tidak pernah dijadikan href
        if (!b.link_url || !isSafeLink(b.link_url)) return <div key={b.id} className={box}>{img}</div>;
        return b.link_url.startsWith("/") ? (
          <Link key={b.id} href={b.link_url} className={box}>
            {img}
          </Link>
        ) : (
          <a key={b.id} href={b.link_url} rel="noopener noreferrer" className={box}>
            {img}
          </a>
        );
      })}
    </div>
  );
}
