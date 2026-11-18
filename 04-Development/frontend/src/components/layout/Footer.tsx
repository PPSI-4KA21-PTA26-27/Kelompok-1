"use client";

import Image from "next/image";
import { useAuth } from "@/app/context/AuthContext";

export default function Footer() {
  const { isStaff } = useAuth();
  // Owner dan staff memakai area admin; footer toko hanya untuk pelanggan.
  if (isStaff) return null;
  return (
    <footer className="bg-ink text-faint">
      <div className="mx-auto flex max-w-360 items-center gap-3 px-6 py-5 sm:px-9 lg:px-14">
        <Image src="/Logo.png" alt="" width={40} height={40} className="size-10" />
        <p className="text-sm">© 2026 Hii Stranger</p>
      </div>
    </footer>
  );
}
