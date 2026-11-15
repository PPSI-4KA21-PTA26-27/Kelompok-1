"use client";

import Link from "next/link";
import { useState } from "react";
import AuthShell from "@/components/auth/AuthShell";
import Field from "@/components/auth/Field";
import { errorMessage } from "@/lib/api";
import { authService } from "@/services";

export default function AturUlangSandiPage() {
  const [done, setDone] = useState(false);
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    const password = String(f.get("password"));
    const token = new URLSearchParams(location.search).get("token");
    if (!token) return setErr("Tautan tidak lengkap.");
    if (password.length < 8) return setErr("Kata sandi minimal 8 karakter.");
    if (password !== f.get("confirm")) return setErr("Konfirmasi kata sandi tidak sama.");
    setBusy(true);
    setErr("");
    try {
      await authService.resetPassword(token, password);
      setDone(true);
    } catch (x) {
      setErr(errorMessage(x));
    } finally {
      setBusy(false);
    }
  }

  return (
    <AuthShell headline={<>Buat kata sandi<br />baru.</>}>
      <h1 className="text-5xl font-extrabold">Atur Ulang Sandi</h1>
      {done ? (
        <>
          <p className="mt-3 text-muted">Kata sandi berhasil diganti. Silakan masuk dengan kata sandi baru.</p>
          <Link href="/masuk" className="mt-6 flex h-[60px] w-full items-center justify-center rounded-xl bg-lime font-bold text-ink">
            Masuk
          </Link>
        </>
      ) : (
        <>
          <p className="mt-1 text-muted">Minimal 8 karakter.</p>
          {err && (
            <p role="alert" className="mt-4 rounded-xl bg-danger-soft p-4 text-danger">
              {err}{" "}
              <Link href="/lupa-sandi" className="font-bold underline">
                Minta tautan baru
              </Link>
            </p>
          )}
          <form onSubmit={submit} className="mt-6 space-y-4">
            <Field label="Kata sandi baru" name="password" type="password" icon="lock" autoComplete="new-password" required minLength={8} />
            <Field label="Ulangi kata sandi" name="confirm" type="password" icon="lock" autoComplete="new-password" required />
            <button disabled={busy} className="mt-2 h-[60px] w-full rounded-xl bg-lime font-bold text-ink disabled:opacity-60">
              {busy ? "Menyimpan..." : "Simpan kata sandi"}
            </button>
          </form>
        </>
      )}
    </AuthShell>
  );
}
