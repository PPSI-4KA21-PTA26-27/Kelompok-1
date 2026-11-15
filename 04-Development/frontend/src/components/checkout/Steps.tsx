import { cn } from "@/lib/cn";

const STEPS = ["Keranjang", "Formulir Pemesanan", "Pembayaran QRIS", "Status Pesanan"];

export default function Steps({ current }: { current: 1 | 2 | 3 | 4 }) {
  return (
    <>
      <div className="sm:hidden">
        <div className="flex gap-1">
          {STEPS.map((s, i) => (
            <div key={s} className={cn("h-1 flex-1 rounded-full", i + 1 <= current ? "bg-lime" : "bg-line")} />
          ))}
        </div>
        <p className="mt-1 text-xs">Langkah {current} dari 4</p>
        <p className="font-semibold">{STEPS[current - 1]}</p>
      </div>
    <ol className="hidden grid-cols-4 gap-4 sm:grid lg:gap-8">
      {STEPS.map((s, i) => (
        <li key={s} aria-current={i + 1 === current ? "step" : undefined}>
          <div className={cn("h-1 rounded-full", i + 1 <= current ? "bg-lime" : "bg-line")} />
          <p className={cn("mt-2 text-xs", i + 1 <= current ? "text-ink" : "text-muted")}>Langkah {i + 1} dari 4</p>
          <p className={cn("text-lg font-semibold", i + 1 <= current ? "text-ink" : "text-muted")}>{s}</p>
        </li>
      ))}
    </ol>
    </>
  );
}
