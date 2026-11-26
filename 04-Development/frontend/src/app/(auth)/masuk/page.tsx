"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { TriangleAlert } from "lucide-react";
import { useAuth } from "@/app/context/AuthContext";
import AuthShell from "@/components/auth/AuthShell";
import Field from "@/components/auth/Field";
import { ApiError } from "@/lib/api";
import { isSafeLink } from "@/lib/safeLink";

export default function MasukPage() {
  const { login } = useAuth();
  const router = useRouter();
  const [err, setErr] = useState<{ bad: boolean; msg: string } | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    setBusy(true);
    setErr(null);
    try {
      const u = await login(String(f.get("email")), String(f.get("password")));
      const next = new URLSearchParams(location.search).get("next");
      // hanya path internal, cegah open redirect
      router.replace(
        u.role !== "customer"
          ? "/admin"
          : next?.startsWith("/") && isSafeLink(next)
            ? next
            : "/",
      );
    } catch (x) {
      setErr({ bad: x instanceof ApiError && x.status === 401, msg: x instanceof Error ? x.message : "Terjadi kesalahan" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <AuthShell headline={<>Masuk dan lanjutkan belanja.</>}>
      <h1 className="text-5xl font-extrabold">Masuk</h1>
      <p className="mt-1 text-muted">Masuk untuk melanjutkan pemesanan!</p>
      {err && (
        <p role="alert" className="mt-4 flex gap-3 rounded-xl bg-danger-soft p-4 text-danger">
          <TriangleAlert className="mt-1 size-4 shrink-0" />
          <span>
            {err.bad ? (
              <>
                <b>Email atau kata sandi salah!</b> Periksa kembali data Anda, lalu coba lagi.
              </>
            ) : (
              err.msg
            )}
          </span>
        </p>
      )}
      <form onSubmit={submit} className="mt-6 space-y-4">
        <Field label="Email" name="email" type="email" icon="user" autoComplete="email" required error={err?.bad ? "Login gagal! Data tidak sesuai." : undefined} />
        <Field label="Kata Sandi" name="password" type="password" icon="lock" autoComplete="current-password" required />
        <button disabled={busy} className="mt-2 h-15 w-full rounded-xl bg-lime font-bold text-ink disabled:opacity-60">
          {busy ? "Memproses..." : "Masuk"}
        </button>
      </form>
      <p className="mt-4 text-center text-lg">
        Belum punya akun?{" "}
        <Link href="/daftar" className="font-bold underline underline-offset-4">
          Daftar
        </Link>
      </p>
    </AuthShell>
  );
}
