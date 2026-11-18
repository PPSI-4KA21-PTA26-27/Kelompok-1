"use client";
import React, { useState, useSyncExternalStore } from "react";
import Link from "next/link";
import { AlertTriangle, Check, ChevronLeft, ChevronRight, Copy, PackageOpen, RotateCw } from "lucide-react";
import toast from "react-hot-toast";
import { cn } from "@/lib/cn";
import { formatRupiah } from "@/lib/format";

export type Tone = "neutral" | "dark" | "lime" | "success" | "warn" | "danger" | "info";

const tones: Record<Tone, string> = {
  neutral: "bg-stone text-ink-3 border-line",
  dark: "bg-ink text-lime border-ink",
  lime: "bg-lime text-ink border-lime",
  success: "bg-success-soft text-success border-transparent",
  warn: "bg-warn-soft text-warn border-transparent",
  danger: "bg-danger-soft text-danger border-transparent",
  info: "bg-info-soft text-info border-transparent",
};

/** Label status berbentuk pill (garment-tag). */
export function Badge({
  tone = "neutral",
  children,
  className,
  dot,
}: {
  tone?: Tone;
  children: React.ReactNode;
  className?: string;
  dot?: boolean;
}) {
  return (
    <span className={cn("caps inline-flex items-center gap-1.5 whitespace-nowrap rounded-full border px-2.5 py-1", tones[tone], className)}>
      {dot && <span className="size-1.5 rounded-full bg-current" aria-hidden />}
      {children}
    </span>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <div className={cn("skeleton rounded-xl", className)} aria-hidden />;
}

export function EmptyState({
  title,
  description,
  action,
  icon,
  className,
}: {
  title: string;
  description?: React.ReactNode;
  action?: React.ReactNode;
  icon?: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col items-center px-6 py-14 text-center", className)}>
      <div className="mb-4 grid size-14 place-items-center rounded-2xl border border-line bg-stone text-muted">
        {icon ?? <PackageOpen className="size-6" />}
      </div>
      <p className="font-display text-lg font-bold uppercase tracking-tight">{title}</p>
      {description && <p className="mt-1.5 max-w-sm text-sm text-muted">{description}</p>}
      {action && <div className="mt-5">{action}</div>}
    </div>
  );
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="flex flex-col items-start gap-3 rounded-2xl border border-danger/20 bg-danger-soft p-5 sm:flex-row sm:items-center">
      <AlertTriangle className="size-5 shrink-0 text-danger" aria-hidden />
      <p className="flex-1 text-sm text-ink">{message}</p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="caps inline-flex h-9 items-center gap-2 rounded-xl bg-ink px-4 text-canvas"
        >
          <RotateCw className="size-3.5" /> Coba lagi
        </button>
      )}
    </div>
  );
}

/** Harga Rupiah (tabular) dengan harga coret bila ada diskon. */
export function Price({
  price,
  discountPercent = 0,
  size = "md",
  className,
}: {
  price: number | string;
  discountPercent?: number;
  size?: "sm" | "md" | "lg";
  className?: string;
}) {
  const base = Number(price) || 0;
  const final = Math.round(base * (1 - (discountPercent || 0) / 100));
  const text = { sm: "text-[15px]", md: "text-lg", lg: "text-2xl md:text-3xl" }[size];
  return (
    <span className={cn("inline-flex flex-wrap items-baseline gap-x-2", className)}>
      <span className={cn("font-display font-bold tabular tracking-tight", text)}>{formatRupiah(final)}</span>
      {discountPercent > 0 && <span className="text-[13px] text-muted line-through tabular">{formatRupiah(base)}</span>}
    </span>
  );
}

/** Tombol salin ke clipboard (no. resi, nominal, no. order). */
export function CopyButton({ value, label = "Salin", className }: { value: string; label?: string; className?: string }) {
  const [done, setDone] = useState(false);
  return (
    <button
      type="button"
      onClick={async () => {
        try {
          await navigator.clipboard.writeText(value);
          setDone(true);
          toast.success("Disalin");
          setTimeout(() => setDone(false), 1500);
        } catch {
          toast.error("Tidak bisa menyalin. Salin manual.");
        }
      }}
      className={cn(
        "inline-flex h-8 items-center gap-1.5 rounded-lg border border-line bg-white px-2.5 text-xs font-semibold hover:border-ink",
        className
      )}
    >
      {done ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
      {label}
    </button>
  );
}

export function Pagination({
  page,
  limit,
  total,
  onPage,
}: {
  page: number;
  limit: number;
  total: number;
  onPage: (p: number) => void;
}) {
  const pages = Math.max(1, Math.ceil(total / Math.max(1, limit)));
  if (pages <= 1) return null;
  const nums: (number | "…")[] = [];
  for (let i = 1; i <= pages; i++) {
    if (i === 1 || i === pages || Math.abs(i - page) <= 1) nums.push(i);
    else if (nums[nums.length - 1] !== "…") nums.push("…");
  }
  const btn = "grid h-10 min-w-10 place-items-center rounded-xl border px-3 text-sm font-semibold transition-colors";
  return (
    <nav className="flex items-center justify-center gap-1.5" aria-label="Halaman">
      <button
        type="button"
        className={cn(btn, "border-line bg-white disabled:opacity-40")}
        disabled={page <= 1}
        onClick={() => onPage(page - 1)}
        aria-label="Sebelumnya"
      >
        <ChevronLeft className="size-4" />
      </button>
      {nums.map((n, i) =>
        n === "…" ? (
          <span key={"e" + i} className="px-1 text-muted">
            …
          </span>
        ) : (
          <button
            key={n}
            type="button"
            aria-current={n === page ? "page" : undefined}
            onClick={() => onPage(n)}
            className={cn(btn, n === page ? "border-ink bg-ink text-lime" : "border-line bg-white hover:border-ink")}
          >
            {n}
          </button>
        )
      )}
      <button
        type="button"
        className={cn(btn, "border-line bg-white disabled:opacity-40")}
        disabled={page >= pages}
        onClick={() => onPage(page + 1)}
        aria-label="Berikutnya"
      >
        <ChevronRight className="size-4" />
      </button>
    </nav>
  );
}

/** Tab berbentuk segmen pill. */
export function Segmented<V extends string>({
  value,
  onChange,
  options,
  className,
}: {
  value: V;
  onChange: (v: V) => void;
  options: { value: V; label: string; count?: number }[];
  className?: string;
}) {
  return (
    <div role="tablist" className={cn("no-scrollbar flex gap-1.5 overflow-x-auto", className)}>
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          role="tab"
          aria-selected={o.value === value}
          onClick={() => onChange(o.value)}
          className={cn(
            "inline-flex h-9 shrink-0 items-center gap-2 rounded-full border px-4 text-[13px] font-semibold transition-colors",
            o.value === value ? "border-ink bg-ink text-lime" : "border-line bg-stone text-ink-3 hover:border-line-strong"
          )}
        >
          {o.label}
          {o.count !== undefined && (
            <span className={cn("rounded-full px-1.5 text-[11px] tabular", o.value === value ? "bg-lime text-ink" : "bg-white text-muted")}>
              {o.count}
            </span>
          )}
        </button>
      ))}
    </div>
  );
}

// Jam bersama (detik-an) untuk useCountdown; null di server supaya hydration cocok.
const clockSubs = new Set<() => void>();
let clockNow: number | null = null;
let clockTimer: ReturnType<typeof setInterval> | null = null;
function subscribeClock(cb: () => void) {
  clockSubs.add(cb);
  clockTimer ??= setInterval(() => {
    clockNow = Date.now();
    clockSubs.forEach((f) => f());
  }, 1000);
  return () => {
    clockSubs.delete(cb);
    if (!clockSubs.size && clockTimer) {
      clearInterval(clockTimer);
      clockTimer = null;
      clockNow = null;
    }
  };
}
const readClock = () => (clockNow ??= Date.now());

/** Hitung mundur ke batas waktu (mis. batas bayar). */
export function useCountdown(target?: string | null) {
  const now = useSyncExternalStore<number | null>(subscribeClock, readClock, () => null);
  if (!target || now === null) return null;
  const ms = new Date(target).getTime() - now;
  if (Number.isNaN(ms)) return null;
  const s = Math.max(0, Math.floor(ms / 1000));
  return {
    expired: ms <= 0,
    text: [Math.floor(s / 3600), Math.floor((s % 3600) / 60), s % 60].map((n) => String(n).padStart(2, "0")).join(":"),
  };
}

/** Breadcrumb: konteks navigasi saja; identitas halaman (nomor pesanan, nama produk) cukup di h1. */
export function Breadcrumb({ items }: { items: { href?: string; label: string }[] }) {
  return (
    <nav aria-label="Breadcrumb" className="text-sm text-muted lg:text-lg">
      <ol className="flex flex-wrap items-center gap-x-2">
        {items.map((it, i) => (
          <li key={it.label} className="flex items-center gap-x-2">
            {i > 0 && <span aria-hidden className="text-sm">&gt;</span>}
            {it.href ? (
              <Link href={it.href} className="py-1 hover:text-ink">{it.label}</Link>
            ) : (
              <span aria-current="page" className="text-ink">{it.label}</span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}

/** Judul seksi editorial: label caps + judul display. */
export function SectionHeading({
  eyebrow,
  title,
  action,
  className,
}: {
  eyebrow?: string;
  title: string;
  action?: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("mb-5 flex items-end justify-between gap-4", className)}>
      <div className="min-w-0">
        {eyebrow && <p className="caps mb-1 text-muted">{eyebrow}</p>}
        <h2 className="font-display text-[26px] font-extrabold uppercase leading-[1.05] tracking-tight md:text-[34px]">{title}</h2>
      </div>
      {action && <div className="shrink-0 whitespace-nowrap">{action}</div>}
    </div>
  );
}
