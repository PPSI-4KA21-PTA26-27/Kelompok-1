"use client";

import Image from "next/image";
import { useCallback, useEffect, useState } from "react";
import { ServerOff } from "lucide-react";
import { API_URL } from "@/lib/api";

/** Ganti seluruh halaman dengan layar "server tidak merespons" saat /health tak terjangkau (jaringan/timeout). Respons HTTP apa pun = server hidup. */
export default function ServerGate({
  children,
}: {
  children: React.ReactNode;
}) {
  const [down, setDown] = useState(false);
  const [busy, setBusy] = useState(false);

  const check = useCallback(async () => {
    try {
      await fetch(API_URL + "/health", {
        cache: "no-store",
        signal: AbortSignal.timeout(5000),
      });
      setDown(false);
    } catch {
      setDown(true);
    }
    setBusy(false);
  }, []);

  useEffect(() => {
    const t0 = setTimeout(check, 0); // di luar badan effect: hindari setState sinkron
    // shortcut: ping tiap 15 dtk selama halaman terbuka; ganti ke deteksi dari error request bila beban health jadi masalah
    const t = setInterval(check, 15000);
    window.addEventListener("online", check);
    return () => {
      clearTimeout(t0);
      clearInterval(t);
      window.removeEventListener("online", check);
    };
  }, [check]);

  if (!down) return children;
  return (
    <div
      role="alert"
      className="flex min-h-screen flex-col items-center justify-center bg-black px-6 text-center text-white"
    >
      <Image
        src="/Logo.png"
        alt="Hii Stranger"
        width={160} height={160}
        priority
      />
      <h1 className="mt-8 max-w-160 text-3xl font-extrabold sm:text-5xl">
        Server tidak merespons
      </h1>
      <p className="mt-6 max-w-145 text-base text-faint sm:text-xl">
        Tidak bisa terhubung ke server. Periksa koneksi internetmu atau coba
        lagi beberapa saat lagi.
      </p>
      <button
        onClick={() => {
          setBusy(true);
          check();
        }}
        disabled={busy}
        className="mt-10 flex items-center gap-2 rounded-full bg-lime px-6 py-3 font-bold text-black disabled:opacity-60"
      >
        <ServerOff className="size-4" /> {busy ? "Memeriksa..." : "Coba lagi"}
      </button>
    </div>
  );
}
