"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { TriangleAlert } from "lucide-react";
import { useAuth } from "@/app/context/AuthContext";
import AuthShell from "@/components/auth/AuthShell";
import Field from "@/components/auth/Field";
import { ApiError } from "@/lib/api";

export default function DaftarPage() {
  const { register } = useAuth();
  const router = useRouter();
  const [err, setErr] = useState("");
  const [fields, setFields] = useState<Record<string, string>>({});
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    setBusy(true);
    setErr("");
    setFields({});
    try {
      await register({
        name: String(f.get("name")),
        email: String(f.get("email")),
        password: String(f.get("password")),
        phone: String(f.get("phone")),
      });
      router.replace("/");
    } catch (x) {
      if (x instanceof ApiError)
        setFields(
          Object.fromEntries(x.details.map((d) => [d.field, d.message])),
        );
      setErr(x instanceof Error ? x.message : "Terjadi kesalahan");
    } finally {
      setBusy(false);
    }
  }

  return (
    <AuthShell
      headline={
        <>
          Buat akun,
          <br />
          lanjut pesan.
        </>
      }
    >
      <h1 className="text-5xl font-extrabold">Daftar</h1>
      <p className="mt-1 text-muted">Isi data berikut untuk membuat akun.</p>
      {err && (
        <p
          role="alert"
          className="mt-4 flex gap-3 rounded-xl bg-danger-soft p-4 text-danger"
        >
          <TriangleAlert className="mt-1 size-4 shrink-0" /> {err}
        </p>
      )}
      <form onSubmit={submit} className="mt-6 space-y-4">
        <Field
          label="Nama Lengkap"
          name="name"
          autoComplete="name"
          required
          error={fields.name}
        />
        <Field
          label="Email"
          name="email"
          type="email"
          autoComplete="email"
          required
          error={fields.email}
        />
        <Field
          label="Kata Sandi"
          name="password"
          type="password"
          autoComplete="new-password"
          required
          error={fields.password}
        />
        <Field
          label="No. Telepon"
          name="phone"
          type="tel"
          autoComplete="tel"
          required
          error={fields.phone}
        />
        <button
          disabled={busy}
          className="mt-2 h-15 w-full rounded-xl bg-lime font-bold text-ink disabled:opacity-60"
        >
          {busy ? "Memproses..." : "Daftar"}
        </button>
      </form>
      <p className="mt-4 text-center text-lg">
        Sudah punya akun?{" "}
        <Link href="/masuk" className="font-bold underline underline-offset-4">
          Masuk
        </Link>
      </p>
    </AuthShell>
  );
}
