"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Pencil, Plus, Search } from "lucide-react";
import {
  Card,
  PageHead,
  Thumb,
  Toggle,
  btnLime,
  inputCls,
  stackTable,
  theadCls,
} from "@/components/admin/ui";
import { EmptyState, Pagination } from "@/components/ui/misc";
import { errorMessage } from "@/lib/api";
import { formatRupiah } from "@/lib/format";
import { adminService } from "@/services";
import type { ApiCategory, ApiProduct } from "@/types/api";

const LIMIT = 20;

export default function AdminProdukPage() {
  const [tab, setTab] = useState<"produk" | "kategori">("produk");
  const [q, setQ] = useState("");
  const [cat, setCat] = useState("");
  const [page, setPage] = useState(1);
  const [rows, setRows] = useState<ApiProduct[] | null>(null);
  const [total, setTotal] = useState(0);
  const [cats, setCats] = useState<ApiCategory[]>([]);
  const [newCat, setNewCat] = useState("");
  const [edit, setEdit] = useState<{ id: string; name: string } | null>(null);

  useEffect(() => {
    adminService
      .categories()
      .then((r) => setCats(r.data))
      .catch((e) => toast.error(errorMessage(e)));
  }, []);

  useEffect(() => {
    const t = setTimeout(() => {
      adminService
        .products({
          q: q.trim() || undefined,
          category_id: cat || undefined,
          page,
          limit: LIMIT,
        })
        .then((r) => {
          setRows(r.data);
          setTotal(r.meta.total);
        })
        .catch((e) => {
          setRows([]);
          toast.error(errorMessage(e));
        });
    }, 300);
    return () => clearTimeout(t);
  }, [q, cat, page]);

  const toggleProduct = async (p: ApiProduct, on: boolean) => {
    try {
      const u = await adminService.updateProduct(p.id, {
        status: on ? "active" : "inactive",
      });
      setRows(
        (rs) =>
          rs?.map((r) => (r.id === p.id ? { ...r, status: u.status } : r)) ??
          null,
      );
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  const saveCat = async (
    id: string,
    input: { name?: string; is_active?: boolean },
  ) => {
    try {
      const u = await adminService.updateCategory(id, input);
      setCats((cs) => cs.map((c) => (c.id === id ? u : c)));
      setEdit(null);
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  const addCat = async (e: React.FormEvent) => {
    e.preventDefault();
    const name = newCat.trim();
    if (!name) return toast.error("Nama kategori wajib diisi.");
    try {
      const c = await adminService.createCategory({ name, is_active: true });
      setCats((cs) => [...cs, c]);
      setNewCat("");
    } catch (err) {
      toast.error(errorMessage(err));
    }
  };

  const catName = (id: string | null) =>
    cats.find((c) => c.id === id)?.name ?? "-";

  return (
    <div>
      <PageHead
        title="Produk & Kategori"
        sub={`${total} produk · ${cats.length} kategori`}
        action={
          tab === "produk" && (
          <Link
            href="/admin/produk/baru"
            className={btnLime + " flex items-center gap-2 px-5 py-3 text-lg"}
          >
            <Plus className="size-5" aria-hidden /> Tambah Produk
          </Link>
          )
        }
      />
      <div role="tablist" className="mb-6 flex gap-2 border-b border-line-strong">
        {(
          [
            ["produk", "Produk", total],
            ["kategori", "Kategori", cats.length],
          ] as const
        ).map(([k, label, n]) => (
          <button
            key={k}
            role="tab"
            aria-selected={tab === k}
            onClick={() => setTab(k)}
            className={
              "-mb-px border-b-2 px-4 py-3 text-base font-bold " +
              (tab === k ? "border-ink text-ink" : "border-transparent text-muted hover:text-ink")
            }
          >
            {label} <span className="ml-1 rounded-full bg-stone-2 px-2 py-0.5 text-xs">{n}</span>
          </button>
        ))}
      </div>
      {tab === "produk" && (
        <div className="min-w-0">
          <div className="mb-5 flex flex-wrap gap-3">
            <label className="relative min-w-60 flex-1">
              <span className="sr-only">Cari produk</span>
              <Search
                className="absolute left-4 top-1/2 size-4 -translate-y-1/2 text-faint"
                aria-hidden
              />
              <input
                value={q}
                onChange={(e) => {
                  setQ(e.target.value);
                  setPage(1);
                }}
                placeholder="Cari nama produk"
                className={inputCls + " h-12 pl-11 text-base"}
              />
            </label>
            <select
              aria-label="Filter kategori"
              value={cat}
              onChange={(e) => {
                setCat(e.target.value);
                setPage(1);
              }}
              className={
                inputCls +
                " h-12 max-w-56 rounded-2xl px-4 text-base font-semibold"
              }
            >
              <option value="">Semua kategori</option>
              {cats.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>
          <div className="overflow-x-auto rounded-xl border border-line-strong bg-white">
            <table className={"w-full text-xs " + stackTable}>
              <thead className={theadCls}>
                <tr>
                  {[
                    "Produk",
                    "Kategori",
                    "Harga",
                    "Stok",
                    "Jenis",
                    "Status",
                  ].map((h) => (
                    <th key={h} className="px-5 py-4 font-semibold">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows?.map((p) => (
                  <tr
                    key={p.id}
                    className="border-t border-line-strong align-top"
                  >
                    <td className="px-5 py-5">
                      <div className="flex gap-3">
                        <Thumb src={p.primary_image_url} alt={p.name} />
                        <p className="max-w-36 font-bold">{p.name}</p>
                      </div>
                      <Link
                        href={"/admin/produk/" + p.id}
                        className="mt-3 inline-flex items-center gap-1 rounded-full border border-ink px-3 py-1 font-bold"
                      >
                        <Pencil className="size-3" aria-hidden /> Ubah
                      </Link>
                    </td>
                    <td data-l="Kategori" className="px-5 py-5">{catName(p.category_id)}</td>
                    <td data-l="Harga" className="px-5 py-5">{formatRupiah(p.price)}</td>
                    <td data-l="Stok" className="px-5 py-5">{p.total_stock ?? 0}</td>
                    <td data-l="Jenis" className="px-5 py-5">
                      <span className="rounded-full bg-ink px-2.5 py-1 font-bold text-white">
                        {p.availability_type === "preorder"
                          ? "Pre-Order"
                          : "Ready Stock"}
                      </span>
                    </td>
                    <td data-l="Status" className="px-5 py-5">
                      <Toggle
                        checked={p.status === "active"}
                        onChange={(v) => toggleProduct(p, v)}
                        label={"Tampilkan " + p.name}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {rows?.length === 0 && (
              <EmptyState title="Belum ada produk" className="py-10" />
            )}
          </div>
          <div className="mt-4">
            <Pagination
              page={page}
              limit={LIMIT}
              total={total}
              onPage={setPage}
            />
          </div>
        </div>
      )}
      {tab === "kategori" && (
        <Card title="Kategori produk" className="max-w-xl">
          <ul className="space-y-3">
            {cats.map((c) => (
              <li key={c.id} className="flex items-center gap-3">
                {edit?.id === c.id ? (
                  <form
                    className="flex flex-1 gap-2"
                    onSubmit={(e) => {
                      e.preventDefault();
                      if (edit.name.trim())
                        saveCat(c.id, { name: edit.name.trim() });
                    }}
                  >
                    <input
                      autoFocus
                      aria-label="Nama kategori"
                      value={edit.name}
                      onChange={(e) =>
                        setEdit({ id: c.id, name: e.target.value })
                      }
                      className={inputCls}
                    />
                    <button className="text-sm font-bold underline">
                      Simpan
                    </button>
                  </form>
                ) : (
                  <p className="flex-1 font-semibold">{c.name}</p>
                )}
                <Toggle
                  checked={c.is_active}
                  onChange={(v) => saveCat(c.id, { is_active: v })}
                  label={"Aktifkan kategori " + c.name}
                />
                <button
                  type="button"
                  aria-label={"Ubah nama " + c.name}
                  onClick={() => setEdit({ id: c.id, name: c.name })}
                >
                  <Pencil className="size-4" aria-hidden />
                </button>
              </li>
            ))}
          </ul>
          <form
            onSubmit={addCat}
            className="mt-4 border-t border-line-strong pt-4"
          >
            <label htmlFor="new-cat" className="mb-2 block font-semibold">
              Kategori baru
            </label>
            <input
              id="new-cat"
              value={newCat}
              onChange={(e) => setNewCat(e.target.value)}
              className={inputCls + " h-10"}
            />
            <button className="mt-3 flex w-full items-center justify-center gap-2 rounded-lg border border-ink py-2 font-semibold">
              <Plus className="size-4" aria-hidden /> Tambah Kategori
            </button>
          </form>
        </Card>
      )}
    </div>
  );
}
