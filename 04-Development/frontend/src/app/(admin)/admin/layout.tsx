"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import {
  CreditCard,
  Files,
  Layers,
  LogOut,
  Menu,
  Tag,
} from "lucide-react";
import { useAuth } from "@/app/context/AuthContext";
import { BrandLoader } from "@/components/ui/Loader";
import { cn } from "@/lib/cn";
import { adminService } from "@/services";

const NAV = [
  {
    href: "/admin/produk",
    label: "Produk & Kategori",
    title: "Daftar Produk dan Kategori",
    icon: Tag,
  },
  {
    href: "/admin/stok",
    label: "Stok & Varian",
    title: "Stok dan Varian Produk",
    icon: Layers,
  },
  {
    href: "/admin/pesanan",
    label: "Pesanan Masuk",
    title: "Daftar Pesanan Masuk",
    icon: Files,
  },
  {
    href: "/admin/pembayaran",
    label: "Validasi Pembayaran",
    title: "Validasi Pembayaran",
    icon: CreditCard,
  },
];

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, loading, isStaff, isAdmin, logout } = useAuth();
  const router = useRouter();
  const path = usePathname();
  const [proofs, setProofs] = useState(0);
  const [open, setOpen] = useState(false);
  const allowed = isStaff;

  useEffect(() => {
    if (loading) return;
    if (!user) router.replace("/masuk?next=" + encodeURIComponent(path));
    else if (!isStaff) router.replace("/");
  }, [loading, user, isStaff, path, router]);

  useEffect(() => {
    if (!isStaff) return;
    adminService
      .payments({ status: "menunggu", limit: 1 })
      .then((r) => setProofs(r.meta.total))
      .catch(() => {});
  }, [isStaff, path]);

  if (loading || !allowed) return <BrandLoader fullscreen />;

  const nav = NAV;
  const current = nav.find((n) => path.startsWith(n.href));
  const title = path.startsWith("/admin/produk/")
    ? "Form Tambah / Ubah Produk"
    : path.startsWith("/admin/pesanan/")
      ? "Detail Pesanan"
      : current?.title;
  const initials = (user?.name ?? "")
    .split(/\s+/)
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  const item = (n: (typeof NAV)[number]) => {
    const active = n === current;
    const Icon = n.icon;
    return (
      <Link
        key={n.href}
        href={n.href}
        onClick={() => setOpen(false)}
        aria-current={active ? "page" : undefined}
        title={n.label}
        className={cn(
          "relative flex shrink-0 items-center gap-3 rounded-xl px-5 py-2.5 font-semibold sm:max-lg:justify-center sm:max-lg:px-0",
          active ? "bg-lime text-ink" : "text-faint hover:text-white",
        )}
      >
        <Icon className="size-4 shrink-0 sm:max-lg:size-5" aria-hidden />
        <span className="flex-1 sm:max-lg:sr-only">{n.label}</span>
        {n.href === "/admin/pembayaran" && proofs > 0 && (
          <span
            className={cn(
              "rounded-full px-2 text-xs font-bold sm:max-lg:absolute sm:max-lg:right-1 sm:max-lg:top-0 sm:max-lg:px-1.5",
              active ? "bg-ink text-lime" : "bg-lime text-ink",
            )}
          >
            {proofs}
          </span>
        )}
      </Link>
    );
  };

  return (
    <div className="min-h-screen bg-canvas sm:flex">
      {open && (
        <button
          aria-label="Tutup menu"
          onClick={() => setOpen(false)}
          className="fixed inset-0 z-30 bg-black/50 sm:hidden"
        />
      )}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 w-68.25 overflow-y-auto bg-ink px-6 py-7 transition-transform sm:sticky sm:top-0 sm:z-auto sm:h-screen sm:w-22.5 sm:shrink-0 sm:translate-x-0 sm:px-3 lg:w-68.25 lg:px-6",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <Link
          href="/admin/produk"
          className="inline-block sm:max-lg:flex sm:max-lg:justify-center"
        >
          <Image
            src="/Logo.png"
            alt="Hii Stranger"
            width={64} height={64} className="size-16 sm:max-lg:size-10"
            priority
          />
        </Link>
        <nav aria-label="Menu admin" className="mt-5 flex flex-col gap-1">
          {NAV.map(item)}
          <button
            onClick={() => logout()}
            title="Keluar"
            className="flex shrink-0 items-center gap-3 rounded-xl px-5 py-2.5 font-semibold text-faint hover:text-white sm:max-lg:justify-center sm:max-lg:px-0 lg:mt-3"
          >
            <LogOut className="size-4 sm:max-lg:size-5" aria-hidden />{" "}
            <span className="sm:max-lg:sr-only">Keluar</span>
          </button>
        </nav>
      </aside>
      <div className="min-w-0 flex-1">
        <header className="flex items-center gap-4 border-b border-line bg-white px-4 py-4 sm:justify-between sm:px-6 sm:py-6 lg:px-10">
          <button
            aria-label="Buka menu"
            aria-expanded={open}
            onClick={() => setOpen(true)}
            className="sm:hidden"
          >
            <Menu className="size-7" aria-hidden />
          </button>
          <p className="flex-1 text-sm font-semibold sm:flex-none sm:truncate sm:text-lg">
            <span className="text-muted">
              {isAdmin ? "Area Owner" : "Area Staff"} /{" "}
            </span>
            {title}
          </p>
          <div className="flex items-center gap-3">
            <span className="rounded-full bg-ink px-3 py-0.5 text-xs font-semibold text-lime sm:text-base sm:text-white">
              {isAdmin ? "Owner" : "Staff"}
            </span>
            <span
              className="hidden size-12 place-items-center sm:grid rounded-full bg-lime text-lg font-bold"
              aria-label={user?.name}
            >
              {initials}
            </span>
          </div>
        </header>
        <main className="px-4 py-6 sm:px-6 sm:py-8 lg:px-10">{children}</main>
      </div>
    </div>
  );
}
