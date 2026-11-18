"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";
import toast from "react-hot-toast";
import { useAuth } from "@/app/context/AuthContext";
import Field from "@/components/auth/Field";
import { Breadcrumb } from "@/components/ui/misc";
import { errorMessage } from "@/lib/api";
import { authService } from "@/services";

const card = "rounded-2xl border border-line-strong bg-white p-6 lg:p-8";
const submit = "h-12 rounded-xl bg-lime px-8 font-bold text-ink disabled:opacity-60";

export default function AkunPage() {
  const { user, loading, setUser } = useAuth();
  const router = useRouter();
  const [busy, setBusy] = useState("");

  useEffect(() => {
    if (loading) return;
    if (!user) router.replace("/masuk?next=/akun");
    else if (user.role !== "customer") router.replace("/admin");
  }, [loading, user, router]);

  if (!user || user.role !== "customer") return null;

  const run = async (key: string, fn: () => Promise<void>, ok: string) => {
    setBusy(key);
    try {
      await fn();
      toast.success(ok);
    } catch (e) {
      toast.error(errorMessage(e));
    } finally {
      setBusy("");
    }
  };

  const saveProfile = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    const name = String(f.get("name")).trim();
    if (!name) return toast.error("Nama wajib diisi.");
    const phone = String(f.get("phone")).trim() || null;
    run("profile", async () => setUser(await authService.updateMe({ name, phone })), "Profil diperbarui.");
  };

  const savePassword = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const f = new FormData(form);
    const next = String(f.get("new_password"));
    if (next.length < 8) return toast.error("Kata sandi baru minimal 8 karakter.");
    if (next !== f.get("confirm")) return toast.error("Konfirmasi kata sandi tidak sama.");
    run(
      "password",
      async () => {
        await authService.changePassword({ current_password: String(f.get("current_password")), new_password: next });
        form.reset();
      },
      "Kata sandi diganti.",
    );
  };

  return (
    <div>
      <Breadcrumb items={[{ href: "/", label: "Beranda" }, { label: "Akun Saya" }]} />
      <div className="mt-2 flex flex-wrap items-end justify-between gap-4">
        <h1 className="text-4xl font-extrabold lg:text-5xl">Akun Saya</h1>
        <Link href="/pesanan" className="font-bold underline">
          Lihat pesanan saya
        </Link>
      </div>

      <div className="mt-6 grid gap-6 lg:mt-8 lg:grid-cols-2">
        <form onSubmit={saveProfile} className={card}>
          <h2 className="text-2xl font-bold">Data diri</h2>
          <div className="mt-5 space-y-4">
            <Field label="Nama" name="name" defaultValue={user.name} required autoComplete="name" />
            <Field label="Nomor telepon" name="phone" type="tel" defaultValue={user.phone ?? ""} autoComplete="tel" />
            <div>
              <p className="mb-2 font-semibold">Email</p>
              <p className="rounded-xl bg-stone px-6 py-4 text-muted">
                {user.email}
              </p>
            </div>
          </div>
          <button disabled={busy === "profile"} className={submit + " mt-6"}>
            {busy === "profile" ? "Menyimpan..." : "Simpan"}
          </button>
        </form>

        <form onSubmit={savePassword} className={card}>
          <h2 className="text-2xl font-bold">Ganti kata sandi</h2>
          <div className="mt-5 space-y-4">
            <Field label="Kata sandi saat ini" name="current_password" type="password" icon="lock" required autoComplete="current-password" />
            <Field label="Kata sandi baru" name="new_password" type="password" icon="lock" required minLength={8} autoComplete="new-password" />
            <Field label="Ulangi kata sandi baru" name="confirm" type="password" icon="lock" required autoComplete="new-password" />
          </div>
          <button disabled={busy === "password"} className={submit + " mt-6"}>
            {busy === "password" ? "Menyimpan..." : "Ganti kata sandi"}
          </button>
        </form>
      </div>

    </div>
  );
}
