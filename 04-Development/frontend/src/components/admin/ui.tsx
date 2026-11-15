import Link from "next/link";
import { cn } from "@/lib/cn";
import { cdn } from "@/lib/img";

/** Kelas bersama back-office (desain admin). */
export const inputCls =
  "h-[34px] w-full rounded-lg border border-line-strong bg-white px-3 text-sm focus:outline-2 focus:outline-ink disabled:bg-stone";
export const labelCls = "mb-2 block text-sm font-semibold";
export const btnLime = "rounded-lg bg-lime px-4 py-2 text-sm font-bold text-ink disabled:opacity-50";
export const btnLine = "rounded-lg border border-ink bg-white px-4 py-2 text-xs font-bold disabled:opacity-50";
export const btnDanger = "rounded-lg border border-danger bg-white px-4 py-2 text-xs font-bold text-danger disabled:opacity-50";
// Mobile: tabel jadi kartu; label baris dari atribut data-l pada <td>.
export const stackTable =
  "max-sm:block max-sm:min-w-0 max-sm:[&_thead]:hidden max-sm:[&_tbody]:block max-sm:[&_tr]:m-3 max-sm:[&_tr]:block max-sm:[&_tr]:rounded-xl max-sm:[&_tr]:border max-sm:[&_tr]:border-line-strong max-sm:[&_td]:flex max-sm:[&_td]:items-center max-sm:[&_td]:justify-between max-sm:[&_td]:gap-3 max-sm:[&_td]:px-4 max-sm:[&_td]:py-1.5 max-sm:[&_td]:before:text-muted max-sm:[&_td]:before:content-[attr(data-l)]";
export const theadCls = "bg-[#b3b3ac] text-left text-xs font-semibold text-ink-3";

export function PageHead({
  title,
  sub,
  crumb,
  action,
}: {
  title: string;
  sub?: React.ReactNode;
  crumb?: { href: string; label: string };
  action?: React.ReactNode;
}) {
  return (
    <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
      <div>
        {crumb && (
          <p className="mb-2 text-sm text-muted sm:text-lg">
            <Link href={crumb.href}>{crumb.label}</Link> <span className="mx-1 text-sm">&gt;</span>{" "}
            <span className="text-ink">{title}</span>
          </p>
        )}
        <h1 className="text-3xl font-extrabold sm:text-4xl md:text-5xl">{title}</h1>
        {sub && <p className="mt-1 text-sm font-semibold text-muted sm:text-lg">{sub}</p>}
      </div>
      {action}
    </div>
  );
}

export function Card({ title, action, className, children }: { title?: string; action?: React.ReactNode; className?: string; children: React.ReactNode }) {
  return (
    <section className={cn("rounded-xl border border-line-strong bg-white p-5", className)}>
      {(title || action) && (
        <div className="mb-4 flex items-start justify-between gap-3">
          {title && <h2 className="text-lg font-bold">{title}</h2>}
          {action}
        </div>
      )}
      {children}
    </section>
  );
}

/** Saklar on/off yang dapat diakses (role="switch"). */
export function Toggle({ checked, onChange, label, disabled }: { checked: boolean; onChange: (v: boolean) => void; label: string; disabled?: boolean }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={cn("relative h-6 w-12 shrink-0 rounded-full transition disabled:opacity-50", checked ? "bg-ink" : "bg-faint")}
    >
      <span className={cn("absolute top-1 size-4 rounded-full transition-all", checked ? "left-7 bg-lime" : "left-1 bg-white")} />
    </button>
  );
}

export function Thumb({ src, alt, className }: { src?: string | null; alt: string; className?: string }) {
  return src ? (
    // eslint-disable-next-line @next/next/no-img-element
    <img src={cdn(src, 120)} alt={alt} loading="lazy" className={cn("size-12 shrink-0 rounded-lg object-cover", className)} />
  ) : (
    <div className={cn("size-12 shrink-0 rounded-lg bg-[#d9d9d9]", className)} aria-hidden />
  );
}
