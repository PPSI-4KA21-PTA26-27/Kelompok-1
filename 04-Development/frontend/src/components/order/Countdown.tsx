"use client";

import { useEffect, useState } from "react";
import { Clock3 } from "lucide-react";
import { formatRemaining } from "@/lib/format";
import { cn } from "@/lib/cn";

/** Sisa waktu bayar, diperbarui tiap 30 detik. */
export default function Countdown({ due, className }: { due: string | null; className?: string }) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 30000);
    return () => clearInterval(t);
  }, []);
  if (!due) return null;
  const left = formatRemaining(new Date(due).getTime() - now);
  return (
    <span className={cn("inline-flex items-center gap-1.5 rounded-full bg-warn-soft px-3 py-1 text-xs font-bold text-warn", className)}>
      <Clock3 className="size-3.5" aria-hidden />
      {left ? `Sisa waktu bayar ${left}` : "Batas bayar lewat"}
    </span>
  );
}
