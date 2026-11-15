"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/app/context/AuthContext";
import { settingService } from "@/services";

/** Arahkan pelanggan ke /maintenance saat mode pemeliharaan aktif; staff/owner tetap lewat. */
export default function MaintenanceGuard() {
  const { isStaff, loading } = useAuth();
  const router = useRouter();
  useEffect(() => {
    if (loading || isStaff) return;
    settingService.get().then((s) => s.maintenance_mode && router.replace("/maintenance")).catch(() => {});
  }, [loading, isStaff, router]);
  return null;
}
