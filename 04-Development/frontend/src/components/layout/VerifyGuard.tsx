"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/app/context/AuthContext";

/** Pelanggan yang belum verifikasi email ditahan di halaman verifikasi (SDD: verifikasi sebelum belanja). */
export default function VerifyGuard() {
  const { user } = useAuth();
  const router = useRouter();
  const path = usePathname();
  const blocked = user?.role === "customer" && !user.email_verified_at;
  useEffect(() => {
    if (blocked) router.replace("/verifikasi-email");
  }, [blocked, path, router]);
  return null;
}
