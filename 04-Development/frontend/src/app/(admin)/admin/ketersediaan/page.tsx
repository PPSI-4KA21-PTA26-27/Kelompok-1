"use client";

import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useMemo, useRef, useState } from "react";
import toast from "react-hot-toast";
import { Search } from "lucide-react";
import {
  AvailBadge,
  Card,
  PageHead,
  Thumb,
  btnLime,
  inputCls,
  labelCls,
} from "@/components/admin/ui";
import { ApiError, errorMessage } from "@/lib/api";
import { cn } from "@/lib/cn";
import { formatRupiah } from "@/lib/format";
import { adminService } from "@/services";
import type { ApiProduct } from "@/types/api";

type Filter = "all" | "ready" | "preorder";

const FILTERS: [Filter, string][] = [
  ["all", "Semua"],
  ["ready", "Ready Stock"],
  ["preorder", "Pre-Order"],
];

const isPo = (p: ApiProduct) => p.availability_type === "preorder";

export default function KetersediaanPage() {
  return (
    <Suspense>
      <Ketersediaan />
    </Suspense>
  );
}

function Ketersediaan() {
  const initial = useSearchParams().get("id") ?? "";
  const [list, setList] = useState<ApiProduct[] | null>(null);
  const [id, setId] = useState(initial);
  const [po, setPo] = useState(false);
  const [days, setDays] = useState("");
  const [busy, setBusy] = useState(false);
  const [q, setQ] = useState("");
  const [filter, setFilter] = useState<Filter>("all");
  const editor = useRef<HTMLElement>(null);

  const sel = list?.find((l) => l.id === id);

  // Form selalu diisi dari data produk terbaru (server), bukan sisa state sebelumnya.
  function apply(x?: ApiProduct) {
    setId(x?.id ?? "");
    setPo(x ? isPo(x) : false);
    setDays(x?.po_estimate_days != null ? String(x.po_estimate_days) : "");
  }

  useEffect(() => {
    // shortcut: 100 produk pertama; tambah paginasi bila katalog lebih besar
    adminService
      .products({ limit: 100, sort: "name" })
      .then((r) => {
        setList(r.data);
        const found = r.data.find((l) => l.id === initial);
        if (initial && !found) toast.error("Produk tidak ditemukan");
        apply(found);
      })
      .catch((e) => {
        setList([]);
        toast.error(errorMessage(e));
      });
  }, [initial]);

  const counts = useMemo(() => {
    const all = list?.length ?? 0;
    const preorder = list?.filter(isPo).length ?? 0;
    return { all, preorder, ready: all - preorder };
  }, [list]);

  const shown = useMemo(() => {
    const s = q.trim().toLowerCase();
    return (list ?? []).filter(
      (l) =>
        (filter === "all" || (filter === "preorder") === isPo(l)) &&
        l.name.toLowerCase().includes(s),
    );
  }, [list, q, filter]);

  const changed =
    !!sel &&
    (po !== isPo(sel) ||
      (po && days !== String(sel.po_estimate_days ?? "")));

  const pick = (x: ApiProduct) => {
    if (changed && x.id !== id && !window.confirm("Perubahan belum disimpan. Pindah produk?")) return;
    apply(x);
    editor.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
  };

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!sel) return toast.error("Pilih produk terlebih dahulu.");
    const n = Number(days);
    if (po && (!/^\d+$/.test(days) || n < 1))
      return toast.error("Estimasi pre-order harus lebih dari 0 hari.");
    setBusy(true);
    try {
      const u = await adminService.updateProduct(sel.id, {
        availability_type: po ? "preorder" : "ready",
        po_estimate_days: po ? n : null,
      });
      const next = { ...sel, ...u };
      setList((ls) => ls?.map((l) => (l.id === next.id ? next : l)) ?? ls);
      apply(next);
      toast.success("Data berhasil disimpan");
    } catch (err) {
      // SDD L-22: penolakan validasi memakai pesan baku layar ini.
      toast.error(err instanceof ApiError && err.status === 422 ? "Pengaturan ditolak. Periksa kembali isian." : errorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  const seg = (on: boolean) =>
    cn(
      "flex-1 rounded-lg py-2.5 text-sm font-bold transition-colors",
      on ? "bg-ink text-lime" : "text-ink-3 hover:text-ink",
    );

  return (
    <div>
      <PageHead
        title="Ready Stock / Pre-Order"
        sub="Pilih produk dari daftar, lalu tentukan jenis ketersediaan dan estimasi pre-order."
      />
      <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <section
          aria-labelledby="semua-produk"
          className="min-w-0 overflow-hidden rounded-xl border border-line-strong bg-white"
        >
          <div className="space-y-4 border-b border-line-strong p-4 sm:p-5">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <h2 id="semua-produk" className="text-lg font-bold">
                Semua produk
              </h2>
              <div
                role="group"
                aria-label="Filter jenis ketersediaan"
                className="flex rounded-lg bg-stone p-1 text-xs font-bold"
              >
                {FILTERS.map(([k, label]) => (
                  <button
                    key={k}
                    type="button"
                    aria-pressed={filter === k}
                    onClick={() => setFilter(k)}
                    className={cn(
                      "whitespace-nowrap rounded-md px-3 py-1.5 transition-colors",
                      filter === k
                        ? "bg-ink text-white"
                        : "text-ink-3 hover:text-ink",
                    )}
                  >
                    {label}
                    <span className="ml-1.5 opacity-60">{counts[k]}</span>
                  </button>
                ))}
              </div>
            </div>
            <label className="relative block">
              <span className="sr-only">Cari produk</span>
              <Search
                className="absolute left-4 top-1/2 size-4 -translate-y-1/2 text-faint"
                aria-hidden
              />
              <input
                type="search"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari nama produk"
                className={inputCls + " pl-11"}
              />
            </label>
          </div>
          {list === null ? (
            <p role="status" className="p-6 text-sm text-muted">
              Memuat produk...
            </p>
          ) : shown.length === 0 ? (
            <p className="p-10 text-center text-sm text-muted">
              {list.length
                ? "Tidak ada produk yang cocok dengan pencarian/filter."
                : "Belum ada produk."}
            </p>
          ) : (
            <ul className="divide-y divide-line-strong">
              {shown.map((l) => {
                const on = l.id === id;
                return (
                  <li key={l.id}>
                    <button
                      type="button"
                      onClick={() => pick(l)}
                      aria-current={on || undefined}
                      className={cn(
                        "flex w-full items-center gap-3 px-4 py-3 text-left transition-colors sm:gap-4 sm:px-5",
                        on
                          ? "bg-lime/15 shadow-[inset_3px_0_0_var(--color-lime-deep)]"
                          : "hover:bg-stone",
                      )}
                    >
                      <Thumb src={l.primary_image_url} alt="" />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-bold">
                          {l.name}
                        </span>
                        <span className="mt-0.5 block truncate text-xs text-muted">
                          {formatRupiah(l.price)} · Stok {l.total_stock ?? 0}
                          {l.status !== "active" && " · Nonaktif"}
                        </span>
                      </span>
                      <AvailBadge
                        type={l.availability_type}
                        days={l.po_estimate_days}
                      />
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </section>

        <aside ref={editor} className="scroll-mt-6 xl:sticky xl:top-6">
          <Card title="Atur ketersediaan">
            {!sel ? (
              <p className="text-sm text-muted">
                Pilih produk dari daftar untuk mengatur jenis ketersediaannya.
              </p>
            ) : (
              <form noValidate onSubmit={save} className="space-y-5">
                <div className="flex items-center gap-3 rounded-lg bg-stone p-3">
                  <Thumb src={sel.primary_image_url} alt="" className="size-10" />
                  <div className="min-w-0">
                    <p className="truncate text-sm font-bold">{sel.name}</p>
                    <p className="text-xs text-muted">
                      Tersimpan:{" "}
                      {isPo(sel)
                        ? "Pre-Order · " + sel.po_estimate_days + " hari"
                        : "Ready Stock"}
                    </p>
                  </div>
                </div>
                <fieldset>
                  <legend className={labelCls}>Jenis ketersediaan</legend>
                  <div className="flex rounded-xl bg-stone p-1">
                    <button
                      type="button"
                      aria-pressed={!po}
                      onClick={() => setPo(false)}
                      className={seg(!po)}
                    >
                      Ready Stock
                    </button>
                    <button
                      type="button"
                      aria-pressed={po}
                      onClick={() => setPo(true)}
                      className={seg(po)}
                    >
                      Pre-Order
                    </button>
                  </div>
                </fieldset>
                {po && (
                  <div>
                    <label htmlFor="days" className={labelCls}>
                      Estimasi waktu pre-order (hari) *
                    </label>
                    <input
                      id="days"
                      type="number"
                      min={1}
                      inputMode="numeric"
                      value={days}
                      onChange={(e) => setDays(e.target.value)}
                      className={inputCls}
                    />
                  </div>
                )}
                <div>
                  <p className={labelCls}>Pratinjau di halaman produk</p>
                  <div className="rounded-lg border border-dashed border-line-strong p-3">
                    <AvailBadge
                      type={po ? "preorder" : "ready"}
                      days={po ? Number(days) || null : null}
                    />
                    <p className="mt-2 text-xs text-muted">
                      {po
                        ? "Pelanggan melihat estimasi pre-order " +
                          (days || "-") +
                          " hari sebelum checkout."
                        : "Produk ready stock tidak menampilkan estimasi."}
                    </p>
                  </div>
                </div>
                <div className="flex items-center justify-end gap-4">
                  <button
                    type="button"
                    onClick={() => apply(sel)}
                    disabled={!changed || busy}
                    className="text-sm font-semibold underline disabled:opacity-40 disabled:no-underline"
                  >
                    Batal
                  </button>
                  <button
                    disabled={busy || !changed}
                    className={btnLime + " px-5 py-3"}
                  >
                    {busy ? "Menyimpan..." : "Simpan"}
                  </button>
                </div>
              </form>
            )}
          </Card>
        </aside>
      </div>
    </div>
  );
}
