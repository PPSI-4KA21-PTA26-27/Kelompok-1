"use client";
import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { tokenStore } from "@/lib/api";
import { authService } from "@/services";
import type { ApiUser } from "@/types/api";

type AuthState = {
  user: ApiUser | null;
  loading: boolean;
  isAdmin: boolean;
  /** Admin atau staff: akun back-office, tidak boleh berbelanja. */
  isStaff: boolean;
  /** Pelanggan yang emailnya sudah terverifikasi (boleh checkout). */
  canOrder: boolean;
  login: (email: string, password: string) => Promise<ApiUser>;
  register: (input: { name: string; email: string; password: string; phone?: string | null }) => Promise<ApiUser>;
  logout: () => void;
  refreshUser: () => Promise<void>;
  setUser: (u: ApiUser) => void;
};

const AuthContext = createContext<AuthState | null>(null);

/** Pengguna dari token tersimpan; null bila tidak ada token atau token ditolak. */
async function fetchMe(): Promise<ApiUser | null> {
  if (!tokenStore.get()) return null;
  try {
    return await authService.me();
  } catch {
    tokenStore.clear();
    return null;
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<ApiUser | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    setUser(await fetchMe());
  }, []);

  useEffect(() => {
    fetchMe()
      .then(setUser)
      .finally(() => setLoading(false));
    const onExpired = () => setUser(null);
    window.addEventListener("auth:expired", onExpired);
    return () => window.removeEventListener("auth:expired", onExpired);
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await authService.login(email, password);
    tokenStore.set(res.access_token);
    setUser(res.user);
    return res.user;
  }, []);

  const register = useCallback(async (input: { name: string; email: string; password: string; phone?: string | null }) => {
    const res = await authService.register(input);
    tokenStore.set(res.access_token);
    setUser(res.user);
    return res.user;
  }, []);

  const logout = useCallback(() => {
    tokenStore.clear();
    setUser(null);
  }, []);

  const value = useMemo<AuthState>(() => {
    const isAdmin = user?.role === "owner";
    return {
      user,
      loading,
      isAdmin,
      isStaff: isAdmin || user?.role === "staff",
      canOrder: user?.role === "customer" && !!user.email_verified_at,
      login,
      register,
      logout,
      refreshUser,
      setUser,
    };
  }, [user, loading, login, register, logout, refreshUser]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth harus dipakai di dalam AuthProvider");
  return ctx;
}
