"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { ChevronDown, LogOut, Menu, Search, ShoppingCart, User, X } from "lucide-react";
import { useSelector } from "react-redux";
import { useAuth } from "@/app/context/AuthContext";
import { selectCartCount } from "@/redux/features/cart-slice";
import { cn } from "@/lib/cn";

const LINKS = [
  { href: "/", label: "Beranda" },
  { href: "/katalog", label: "Katalog" },
  { href: "/pesanan", label: "Pesanan Saya" },
];

const icon = "grid size-10 shrink-0 place-items-center rounded-full hover:bg-ink-2";
const item = "flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-left text-sm font-semibold hover:bg-stone";

export default function Navbar() {
  const path = usePathname();
  const router = useRouter();
  const { user, isStaff, logout } = useAuth();
  const count = useSelector(selectCartCount);
  const [q, setQ] = useState("");
  const [menu, setMenu] = useState(false);
  const [acct, setAcct] = useState(false);
  const active = (href: string) => (href === "/" ? path === "/" : path.startsWith(href));

  useEffect(() => {
    const esc = (e: KeyboardEvent) => e.key === "Escape" && (setMenu(false), setAcct(false));
    addEventListener("keydown", esc);
    return () => removeEventListener("keydown", esc);
  }, []);

  const out = () => {
    setAcct(false);
    logout();
    router.push("/");
  };

  return (
    <header className="relative z-30 bg-ink text-white">
      <div className="mx-auto flex h-16 max-w-[1440px] items-center gap-2 px-4 sm:px-6 lg:gap-4 lg:px-14">
        <button type="button" aria-label="Menu" aria-expanded={menu} onClick={() => setMenu(!menu)} className={cn(icon, "-ml-2 lg:hidden")}>
          {menu ? <X className="size-5" /> : <Menu className="size-5" />}
        </button>
        <Link href="/" aria-label="Hii Stranger" className="shrink-0">
          <Image src="/Logo.png" alt="Hii Stranger" width={68} height={60} className="h-9 w-auto" priority />
        </Link>
        <nav aria-label="Utama" className="hidden gap-1 lg:flex">
          {LINKS.map((l) => (
            <Link
              key={l.href}
              href={l.href}
              className={cn(
                "rounded-full px-4 py-2 text-sm font-semibold",
                active(l.href) ? "bg-lime text-ink" : "text-faint hover:text-white",
              )}
            >
              {l.label}
            </Link>
          ))}
        </nav>
        <form
          role="search"
          onSubmit={(e) => {
            e.preventDefault();
            router.push(q.trim() ? `/katalog?q=${encodeURIComponent(q.trim())}` : "/katalog");
          }}
          className="relative ml-auto hidden w-64 lg:block"
        >
          <Search className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted" aria-hidden />
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Cari t-shirt..."
            aria-label="Cari produk"
            className="h-10 w-full rounded-full bg-ink-2 pl-10 pr-4 text-sm text-white placeholder:text-muted focus:outline-2 focus:outline-lime"
          />
        </form>
        <Link href="/katalog" aria-label="Cari produk" className={cn(icon, "ml-auto lg:hidden")}>
          <Search className="size-5" />
        </Link>
        <Link href="/keranjang" aria-label={`Keranjang, ${count} barang`} className={cn(icon, "relative")}>
          <ShoppingCart className="size-5" />
          {count > 0 && (
            <span className="absolute right-0.5 top-0.5 grid h-4 min-w-4 place-items-center rounded-full bg-lime px-1 text-[10px] font-bold text-ink">
              {count > 99 ? "99+" : count}
            </span>
          )}
        </Link>
        {user ? (
          <div className="relative">
            <button
              type="button"
              aria-label="Menu akun"
              aria-haspopup="menu"
              aria-expanded={acct}
              onClick={() => setAcct(!acct)}
              className="flex h-10 items-center gap-2 rounded-full bg-ink-2 pl-3 pr-2.5 text-sm font-semibold hover:bg-ink-3"
            >
              <User className="size-4" aria-hidden />
              <span className="hidden max-w-24 truncate lg:inline">{user.name.split(" ")[0]}</span>
              <ChevronDown className={cn("size-4 transition-transform", acct && "rotate-180")} aria-hidden />
            </button>
            {acct && (
              <>
                <button type="button" aria-hidden tabIndex={-1} className="fixed inset-0 cursor-default" onClick={() => setAcct(false)} />
                <div role="menu" className="absolute right-0 top-12 w-60 rounded-xl border border-line-strong bg-white p-2 text-ink shadow-lg">
                  <div className="border-b border-line-strong px-3 pb-2.5 pt-1.5">
                    <p className="truncate font-bold">{user.name}</p>
                    <p className="truncate text-xs text-muted">{user.email}</p>
                  </div>
                  <div className="py-1">
                    <Link role="menuitem" href={isStaff ? "/admin" : "/akun"} onClick={() => setAcct(false)} className={item}>
                      <User className="size-4" aria-hidden /> {isStaff ? "Dasbor Admin" : "Akun Saya"}
                    </Link>
                  </div>
                  <div className="border-t border-line-strong pt-1">
                    <button role="menuitem" type="button" onClick={out} className={cn(item, "text-danger")}>
                      <LogOut className="size-4" aria-hidden /> Keluar
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        ) : (
          <Link href="/masuk" className="flex h-10 shrink-0 items-center gap-2 rounded-full bg-lime px-4 text-sm font-bold text-ink">
            <User className="size-4" aria-hidden /> Masuk
          </Link>
        )}
      </div>
      {menu && (
        <nav aria-label="Utama" className="border-t border-ink-2 px-4 pb-3 pt-2 sm:px-6 lg:hidden">
          {LINKS.map((l) => (
            <Link
              key={l.href}
              href={l.href}
              onClick={() => setMenu(false)}
              className={cn(
                "block rounded-lg px-3 py-2.5 text-sm font-semibold",
                active(l.href) ? "bg-lime text-ink" : "text-faint hover:text-white",
              )}
            >
              {l.label}
            </Link>
          ))}
        </nav>
      )}
    </header>
  );
}
