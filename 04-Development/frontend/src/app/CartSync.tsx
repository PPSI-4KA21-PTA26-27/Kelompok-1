"use client";
import { useEffect } from "react";
import { useAuth } from "@/app/context/AuthContext";
import { useCartActions } from "@/hooks/useCartActions";

/** Memuat keranjang dari backend setiap kali status login berubah. */
export default function CartSync() {
  const { user, loading } = useAuth();
  const { refresh } = useCartActions();
  useEffect(() => {
    if (!loading) refresh().catch(() => {});
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user?.id, user?.role, loading]);
  return null;
}
