"use client";

import Link from "next/link";
import { useState } from "react";
import AuthShell from "@/components/auth/AuthShell";
import Field from "@/components/auth/Field";
import { errorMessage } from "@/lib/api";
import { authService } from "@/services";

export default function LupaSandiPage() {
  const [sent, setSent] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const email = String(new FormData(e.currentTarget).get("email")).trim();
    setBusy(true);
    setErr("");
    try {
      await authService.forgotPassword(email);
      setSent(email);
    } catch (x) {
      setErr(errorMessage(x));
    } finally {
      setBusy(false);
    }
  }

  return (
    <AuthShell headline={<>Lupa kata sandi?<br />Tenang.</>}>
      <h1 className="text-5xl font-extrabold">Lupa Kata Sandi</h1>
      {sent ? (
        <>
          <p className="mt-3 text-muted">
            Bila <b className="text-ink">{sent}</b> terdaftar, tautan atur ulang kata sandi sudah dikirim. Tautan hanya berlaku sebentar.
          </p>
          <p className="mt-4 rounded-xl bg-stone p-4 text-sm">Tidak ada di kotak masuk? Cek folder Spam atau Promosi.</p>
        </>
      ) : (
        <>
          <p className="mt-1 text-muted">Masukkan email akunmu, kami kirim tautan untuk membuat kata sandi baru.</p>
          {err && (
            <p role="alert" className="mt-4 rounded-xl bg-danger-soft p-4 text-danger">
              {err}
            </p>
          )}
          <form onSubmit={submit} className="mt-6 space-y-4">
            <Field label="Email" name="email" type="email" icon="user" autoComplete="email" required />
            <button disabled={busy} className="mt-2 h-[60px] w-full rounded-xl bg-lime font-bold text-ink disabled:opacity-60">
              {busy ? "Mengirim..." : "Kirim tautan"}
            </button>
          </form>
        </>
      )}
      <p className="mt-4 text-center text-lg">
        Ingat kata sandi?{" "}
        <Link href="/masuk" className="font-bold underline underline-offset-4">
          Masuk
        </Link>
      </p>
    </AuthShell>
  );
}
