"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { CircleCheck, MailCheck, TriangleAlert } from "lucide-react";
import { useAuth } from "@/app/context/AuthContext";
import AuthShell from "@/components/auth/AuthShell";
import { errorMessage } from "@/lib/api";
import { authService } from "@/services";

const btn = "flex h-[60px] w-full items-center justify-center rounded-xl font-bold disabled:opacity-60";

export default function VerifikasiEmailPage() {
  const { user, loading, logout, setUser } = useAuth();
  const router = useRouter();
  // token di URL = pengguna datang dari tautan email; tanpa token = menunggu verifikasi.
  const [state, setState] = useState<"wait" | "ok" | "fail">("wait");
  const [msg, setMsg] = useState("");
  const [sent, setSent] = useState(false);
  const [busy, setBusy] = useState(false);
  const [wait, setWait] = useState(0);

  // jeda kirim ulang supaya tombol tidak dipencet berulang (backend juga membatasi)
  useEffect(() => {
    if (!wait) return;
    const t = setTimeout(() => setWait(wait - 1), 1000);
    return () => clearTimeout(t);
  }, [wait]);

  useEffect(() => {
    const token = new URLSearchParams(location.search).get("token");
    if (!token) return;
    authService
      .verifyEmail(token)
      .then((u) => {
        setState("ok");
        if (user?.id === u.id) setUser(u);
      })
      .catch((e) => {
        setState("fail");
        setMsg(errorMessage(e));
      });
    // hanya sekali saat halaman dibuka; token sekali pakai.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (loading || state !== "wait" || new URLSearchParams(location.search).has("token")) return;
    if (!user) router.replace("/masuk?next=/verifikasi-email");
    else if (user.role !== "customer" || user.email_verified_at) router.replace(user.role === "customer" ? "/" : "/admin");
  }, [loading, user, state, router]);

  const resend = async () => {
    setBusy(true);
    setMsg("");
    try {
      await authService.resendVerification();
      setSent(true);
      setWait(60);
    } catch (e) {
      setMsg(errorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  if (state === "ok" || state === "fail") {
    const ok = state === "ok";
    const Icon = ok ? CircleCheck : TriangleAlert;
    return (
      <AuthShell headline={ok ? <>Email terverifikasi.</> : <>Tautan tidak berlaku.</>}>
        <Icon className={"size-12 " + (ok ? "text-lime-ink" : "text-danger")} aria-hidden />
        <h1 className="mt-4 text-4xl font-extrabold">{ok ? "Verifikasi berhasil" : "Verifikasi gagal"}</h1>
        <p className="mt-2 text-muted">
          {ok ? "Akunmu sudah aktif. Sekarang kamu bisa belanja dan checkout." : msg + " Minta tautan baru dari halaman verifikasi."}
        </p>
        <Link href={ok ? (user ? "/" : "/masuk") : "/verifikasi-email"} onClick={() => setState("wait")} className={btn + " mt-6 bg-lime text-ink"}>
          {ok ? (user ? "Mulai belanja" : "Masuk") : "Kirim ulang tautan"}
        </Link>
      </AuthShell>
    );
  }

  if (!user) {
    return (
      <AuthShell headline={<>Sebentar.</>}>
        <p className="text-lg text-muted">Memuat...</p>
      </AuthShell>
    );
  }

  return (
    <AuthShell headline={<>Satu langkah<br />lagi.</>}>
      <MailCheck className="size-12 text-lime-ink" aria-hidden />
      <h1 className="mt-4 text-4xl font-extrabold">Cek email kamu</h1>
      <p className="mt-2 text-muted">
        Kami mengirim tautan verifikasi ke <b className="text-ink">{user.email}</b>. Buka tautannya untuk mengaktifkan akun sebelum lanjut
        belanja.
      </p>
      <p className="mt-4 flex gap-3 rounded-xl bg-stone p-4 text-sm">
        <TriangleAlert className="mt-0.5 size-4 shrink-0" aria-hidden />
        <span>
          Tidak ada di kotak masuk? Cek folder <b>Spam</b> atau <b>Promosi</b>, lalu tandai &quot;Bukan spam&quot; supaya email berikutnya masuk
          normal. Tautan berlaku 24 jam.
        </span>
      </p>
      {msg && (
        <p role="alert" className="mt-4 rounded-xl bg-danger-soft p-4 text-sm text-danger">
          {msg}
        </p>
      )}
      {sent && (
        <p role="status" className="mt-4 rounded-xl bg-success-soft p-4 text-sm text-success">
          Tautan baru sudah dikirim.
        </p>
      )}
      <button onClick={resend} disabled={busy || wait > 0} className={btn + " mt-6 bg-lime text-ink"}>
        {busy ? "Mengirim..." : wait ? `Kirim ulang dalam ${wait} detik` : "Kirim ulang email"}
      </button>
      <button onClick={() => refresh()} className={btn + " mt-3 border border-ink"}>
        Saya sudah verifikasi
      </button>
      <p className="mt-4 text-center">
        Salah email?{" "}
        <button
          onClick={() => {
            logout();
            router.replace("/daftar");
          }}
          className="font-bold underline underline-offset-4"
        >
          Daftar ulang
        </button>
      </p>
    </AuthShell>
  );

  async function refresh() {
    setMsg("");
    try {
      const u = await authService.me();
      setUser(u);
      if (!u.email_verified_at) setMsg("Email belum terverifikasi. Buka tautan di email terlebih dahulu.");
    } catch (e) {
      setMsg(errorMessage(e));
    }
  }
}
