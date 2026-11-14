"use client";
import React, { useEffect, useRef } from "react";
import { usePathname, useSearchParams } from "next/navigation";
import { cn } from "@/lib/cn";

/** Loader merek: tile "H!" dengan cincin lime berputar + titik memantul. */
export function BrandLoader({
  label = "Memuat",
  className,
  fullscreen,
}: {
  label?: string;
  className?: string;
  fullscreen?: boolean;
}) {
  return (
    <div
      role="status"
      aria-live="polite"
      className={cn(
        "grid place-items-center animate-[fade-in_.3s_ease-out]",
        fullscreen ? "min-h-screen" : "min-h-[50vh]",
        className,
      )}
    >
      <div className="flex flex-col items-center gap-4">
        <span className="relative grid size-16 place-items-center">
          <span className="hs-ring absolute inset-0 rounded-full" aria-hidden />
          <span className="hs-tile relative grid size-12 place-items-center rounded-2xl bg-ink font-display text-lg font-extrabold text-lime">
            H!
          </span>
        </span>
        <span className="flex items-center gap-1.5 text-sm font-medium text-muted">
          {label}
          <span className="hs-dots inline-flex gap-0.5" aria-hidden>
            <i />
            <i />
            <i />
          </span>
        </span>
      </div>
    </div>
  );
}

/** Spinner kecil untuk tombol / inline. */
export function Spinner({ className }: { className?: string }) {
  return (
    <span
      aria-hidden
      className={cn(
        "inline-block size-4 animate-spin rounded-full border-2 border-current border-r-transparent",
        className,
      )}
    />
  );
}

/**
 * Bar progres tipis di atas layar saat berpindah halaman. Dimulai saat
 * tautan internal diklik, selesai ketika path/query berubah. Manipulasi DOM
 * langsung (tanpa state) supaya tidak memicu render ulang.
 */
export function RouteProgress() {
  const bar = useRef<HTMLDivElement>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pathname = usePathname();
  const search = useSearchParams();
  const key = pathname + "?" + search.toString();

  useEffect(() => {
    const el = bar.current;
    if (!el) return;
    if (el.dataset.state === "loading") {
      el.dataset.state = "done";
      if (timer.current) clearTimeout(timer.current);
      timer.current = setTimeout(() => {
        el.dataset.state = "idle";
      }, 450);
    }
  }, [key]);

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (e.defaultPrevented || e.button !== 0) return;
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
      const a = (e.target as HTMLElement | null)?.closest?.("a");
      if (!a || a.target === "_blank" || a.hasAttribute("download")) return;
      const href = a.getAttribute("href");
      if (!href || href.startsWith("#") || /^(mailto|tel):/.test(href)) return;
      const url = new URL(a.href, location.href);
      if (url.origin !== location.origin) return;
      if (url.pathname === location.pathname && url.search === location.search) return;
      const el = bar.current;
      if (!el) return;
      if (timer.current) clearTimeout(timer.current);
      el.dataset.state = "idle";
      // paksa reflow supaya animasi mulai dari 0
      void el.offsetWidth;
      el.dataset.state = "loading";
    };
    document.addEventListener("click", onClick, true);
    return () => document.removeEventListener("click", onClick, true);
  }, []);

  return <div ref={bar} data-state="idle" className="hs-progress" aria-hidden />;
}
