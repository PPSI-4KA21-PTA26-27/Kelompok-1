"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Trash2, TriangleAlert } from "lucide-react";
import {
  Card,
  PageHead,
  Thumb,
  btnLime,
  btnLine,
  inputCls,
  stackTable,
} from "@/components/admin/ui";
import { Badge, EmptyState } from "@/components/ui/misc";
import { errorMessage } from "@/lib/api";
import { cn } from "@/lib/cn";
import { adminService } from "@/services";
import type { ApiProduct, ApiProductDetail } from "@/types/api";

const NEW = { sku: "", size: "", color: "", stock: "0" };

export default function StokPage() {
  return (
    <Suspense>
      <Stok />
    </Suspense>
  );
}

function Stok() {
  const router = useRouter();
  const id = useSearchParams().get("id");
  const [list, setList] = useState<ApiProduct[]>([]);
  const [loaded, setP] = useState<ApiProductDetail | null>(null);
  const [draft, setDraft] = useState<Record<string, string>>({});
  const [nv, setNv] = useState<typeof NEW | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    // shortcut: 100 produk pertama; tambah pencarian/paginasi bila katalog lebih besar
    adminService
      .products({ limit: 100, sort: "name" })
      .then((r) => setList(r.data))
      .catch((e) => toast.error(errorMessage(e)));
  }, []);

  useEffect(() => {
    if (!id) return;
    adminService
      .product(id)
      .then(setP)
      .catch((e) => toast.error(errorMessage(e)));
  }, [id]);

  const p = loaded?.id === id ? loaded : null;
  const total = p?.variants.reduce((n, v) => n + v.stock, 0) ?? 0;

  const saveStock = async () => {
    if (!p) return;
    const changed = Object.entries(draft).filter(
      ([vid, s]) => String(p.variants.find((v) => v.id === vid)?.stock) !== s,
    );
    if (changed.some(([, s]) => !/^\d+$/.test(s)))
      return toast.error("Stok harus bilangan bulat ≥ 0.");
    setBusy(true);
    try {
      const done = await Promise.all(
        changed.map(([vid, s]) =>
          adminService.updateVariant(p.id, vid, { stock: Number(s) }),
        ),
      );
      setP({
        ...p,
        variants: p.variants.map((v) => done.find((d) => d.id === v.id) ?? v),
      });
      setDraft({});
      toast.success("Stok disimpan.");
    } catch (e) {
      toast.error(errorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const addVariant = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!p || !nv) return;
    if (!nv.sku.trim() || !nv.size.trim())
      return toast.error("Kode dan ukuran wajib diisi.");
    if (!/^\d+$/.test(nv.stock))
      return toast.error("Stok harus bilangan bulat ≥ 0.");
    try {
      const v = await adminService.createVariant(p.id, {
        sku: nv.sku.trim(),
        size: nv.size.trim(),
        color: nv.color.trim() || null,
        stock: Number(nv.stock),
      });
      setP({ ...p, variants: [...p.variants, v] });
      setNv(null);
    } catch (err) {
      toast.error(errorMessage(err));
    }
  };

  const delVariant = async (vid: string) => {
    if (!p || !confirm("Hapus varian ini?")) return;
    try {
      await adminService.deleteVariant(p.id, vid);
      setP({ ...p, variants: p.variants.filter((v) => v.id !== vid) });
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  const hide = async () => {
    if (!p) return;
    try {
      const u = await adminService.updateProduct(p.id, { status: "inactive" });
      setP({ ...p, status: u.status });
      toast.success("Produk disembunyikan.");
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  return (
    <div>
      <PageHead
        title="Stok & Varian"
        sub="Perbarui stok per ukuran dan warna."
      />
      <div className="grid gap-8 xl:grid-cols-[397px_1fr]">
        <Card title="Produk" className="self-start">
          <ul className="max-h-[70vh] space-y-3 overflow-y-auto">
            {list.map((x) => (
              <li key={x.id}>
                <button
                  onClick={() => {
                    setDraft({});
                    setNv(null);
                    router.replace("/admin/stok?id=" + x.id);
                  }}
                  aria-current={x.id === id ? "true" : undefined}
                  className={cn(
                    "flex w-full items-center gap-3 rounded-xl border p-3 text-left",
                    x.id === id
                      ? "border-ink bg-lime/30"
                      : "border-line-strong",
                  )}
                >
                  <Thumb src={x.primary_image_url} alt={x.name} />
                  <span className="text-sm">
                    <b className="block">{x.name}</b>
                    <span className="text-muted">
                      {(x.total_stock ?? 0) +
                        " pcs · " +
                        (x.variant_count ?? 0) +
                        " varian"}
                    </span>
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </Card>

        <div className="min-w-0 space-y-6">
          {!p ? (
            <Card>
              <EmptyState
                title="Pilih produk"
                description="Pilih produk dari daftar untuk mengelola variannya."
              />
            </Card>
          ) : (
            <>
              <Card
                title={p.name}
                action={
                  <button
                    onClick={() => setNv(nv ? null : NEW)}
                    className={btnLine}
                  >
                    + Tambah Varian
                  </button>
                }
              >
                <p className="-mt-3 mb-4 text-xs text-muted">
                  Total Stock: {total} pcs
                </p>
                <div className="overflow-x-auto rounded-xl border border-line-strong px-4">
                  <table
                    className={"w-full text-sm sm:min-w-130 " + stackTable}
                  >
                    <thead>
                      <tr className="text-left font-semibold text-ink-3">
                        {["Kode", "Ukuran", "Warna", "Stok", "Status", ""].map(
                          (h, i) => (
                            <th key={i} className="py-3 pr-3 font-semibold">
                              {h}
                            </th>
                          ),
                        )}
                      </tr>
                    </thead>
                    <tbody>
                      {p.variants.map((v) => (
                        <tr key={v.id} className="border-t border-line-strong">
                          <td className="py-3 pr-3">{v.sku}</td>
                          <td data-l="Ukuran" className="py-3 pr-3">
                            {v.size}
                          </td>
                          <td data-l="Warna" className="py-3 pr-3">
                            {v.color ?? "-"}
                          </td>
                          <td data-l="Stok" className="py-3 pr-3">
                            <input
                              aria-label={"Stok " + v.sku}
                              inputMode="numeric"
                              value={draft[v.id] ?? String(v.stock)}
                              onChange={(e) =>
                                setDraft((d) => ({
                                  ...d,
                                  [v.id]: e.target.value.trim(),
                                }))
                              }
                              className={inputCls + " w-20"}
                            />
                          </td>
                          <td data-l="Status" className="py-3 pr-3">
                            {v.stock === 0 ? (
                              <Badge tone="danger">Habis</Badge>
                            ) : v.stock <= 3 ? (
                              <Badge tone="warn">Menipis</Badge>
                            ) : (
                              <Badge tone="success">Tersedia</Badge>
                            )}
                          </td>
                          <td className="py-3">
                            <button
                              onClick={() => delVariant(v.id)}
                              aria-label={"Hapus varian " + v.sku}
                            >
                              <Trash2 className="size-4" aria-hidden />
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {p.variants.length === 0 && !nv && (
                    <p className="py-6 text-center text-sm text-muted">
                      Belum ada varian.
                    </p>
                  )}
                  {nv && (
                    <form
                      onSubmit={addVariant}
                      className="flex flex-wrap items-end gap-2 border-t border-line-strong py-3"
                    >
                      {(["sku", "size", "color", "stock"] as const).map((k) => (
                        <input
                          key={k}
                          aria-label={
                            {
                              sku: "Kode",
                              size: "Ukuran",
                              color: "Warna",
                              stock: "Stok",
                            }[k]
                          }
                          placeholder={
                            {
                              sku: "Kode",
                              size: "Ukuran",
                              color: "Warna",
                              stock: "Stok",
                            }[k]
                          }
                          value={nv[k]}
                          onChange={(e) =>
                            setNv({ ...nv, [k]: e.target.value })
                          }
                          className={inputCls + " w-28"}
                        />
                      ))}
                      <button className={btnLime}>Tambah</button>
                    </form>
                  )}
                </div>
              </Card>
              <Card>
                {total === 0 && (
                  <p className="mb-4 flex items-center gap-2 rounded-lg bg-warn-soft px-4 py-2 text-sm text-warn">
                    <TriangleAlert className="size-4 shrink-0" aria-hidden />
                    <span>
                      <b>Semua varian habis?</b> Sembunyikan produk agar tidak
                      dapat dipesan pelanggan.
                    </span>
                  </p>
                )}
                <div className="flex flex-wrap gap-2">
                  <button
                    onClick={hide}
                    disabled={total > 0 || p.status !== "active"}
                    className={btnLine + " text-sm"}
                  >
                    Sembunyikan Produk Habis
                  </button>
                  <button
                    onClick={saveStock}
                    disabled={busy || Object.keys(draft).length === 0}
                    className={btnLime}
                  >
                    Simpan Stok
                  </button>
                </div>
                <p className="mt-4 text-[10px] text-muted">
                  Stok website tidak tersinkron otomatis dengan Shopee dan
                  Tiktok Shop. Perbarui secara manual.
                </p>
              </Card>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
