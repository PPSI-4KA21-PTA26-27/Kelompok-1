"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { Wrench } from "lucide-react";
import { settingService } from "@/services";

export default function MaintenancePage() {
  const [msg, setMsg] = useState("");
  useEffect(() => {
    settingService.get().then((s) => setMsg(s.maintenance_message)).catch(() => {});
  }, []);
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-black px-6 text-center text-white">
      <Image src="/Logo.png" alt="Hii Stranger" width={180} height={160} priority />
      <h1 className="mt-8 max-w-[640px] text-3xl font-extrabold sm:text-5xl">Website sedang dalam pemeliharaan</h1>
      <p className="mt-6 max-w-[580px] text-base text-faint sm:text-xl">{msg || "Kami sedang memperbarui toko. Silakan kembali dalam beberapa saat lagi"}</p>
      <p className="mt-10 flex items-center gap-2 rounded-full bg-ink-2 px-5 py-3 font-bold text-lime">
        <Wrench className="size-4" /> Pemeliharaan sedang berlangsung
      </p>
      <p className="absolute bottom-16 text-faint">
        Admin dan owner tetap dapat{" "}
        <Link href="/masuk" className="font-bold text-white underline underline-offset-4">
          masuk
        </Link>
        .
      </p>
    </div>
  );
}
