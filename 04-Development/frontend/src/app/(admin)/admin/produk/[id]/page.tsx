"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Info, Upload, X } from "lucide-react";
import {
  AvailBadge,
  Card,
  PageHead,
  Thumb,
  Toggle,
  btnLime,
  btnLine,
  inputCls,
  labelCls,
} from "@/components/admin/ui";
import { errorMessage } from "@/lib/api";
import { finalPrice, formatRupiah } from "@/lib/format";
import { adminService } from "@/services";
import type { ApiCategory, ApiProductDetail } from "@/types/api";

const EMPTY = {
  name: "",
  category_id: "",
  weight_gram: "",
  description: "",
  price: "",
  discount_percent: "",
  active: true,
};

export default function AdminProdukForm() {
  const { id } = useParams<{ id: string }>();
  const isNew = id === "baru";
  const router = useRouter();
  const [cats, setCats] = useState<ApiCategory[]>([]);
  const [p, setP] = useState<ApiProductDetail | null>(null);
  const [f, setF] = useState(EMPTY);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    adminService
      .categories()
      .then((r) => setCats(r.data))
      .catch(() => {});
    if (isNew) return;
    adminService
      .product(id)
      .then((d) => {
        setP(d);
        setF({
          name: d.name,
          category_id: d.category_id ?? "",
          weight_gram: String(d.weight_gram),
          description: d.description ?? "",
          price: String(Number(d.price)),
          discount_percent: d.discount_percent
            ? String(d.discount_percent)
            : "",
          active: d.status === "active",
        });
      })
      .catch((e) => toast.error(errorMessage(e)));
  }, [id, isNew]);

  const set =
    (k: keyof typeof EMPTY) =>
    (
      e: React.ChangeEvent<
        HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
      >,
    ) =>
      setF((s) => ({ ...s, [k]: e.target.value }));

  const save = async () => {
    const price = Number(f.price);
    const weight = Number(f.weight_gram);
    const discount = Number(f.discount_percent || 0);
    if (!f.name.trim()) return toast.error("Nama produk wajib diisi.");
    if (!f.category_id) return toast.error("Kategori wajib dipilih.");
    if (!Number.isInteger(weight) || weight < 1)
      return toast.error("Berat minimal 1 gram.");
    if (!(price > 0)) return toast.error("Harga normal harus lebih dari 0.");
    if (!Number.isInteger(discount) || discount < 0 || discount > 100)
      return toast.error("Diskon 0 sampai 100.");
    const input = {
      name: f.name.trim(),
      category_id: f.category_id,
      weight_gram: weight,
      description: f.description.trim() || null,
      price: String(price),
      discount_percent: discount,
      status: (f.active ? "active" : "inactive") as "active" | "inactive",
    };
    setBusy(true);
    try {
      if (isNew) {
        const c = await adminService.createProduct({
          ...input,
          availability_type: "ready",
          po_estimate_days: null,
        });
        toast.success("Produk dibuat. Tambahkan gambar dan varian.");
        router.replace("/admin/produk/" + c.id);
      } else {
        await adminService.updateProduct(id, input);
        toast.success("Perubahan disimpan.");
        router.push("/admin/produk");
      }
    } catch (e) {
      toast.error(errorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const addImage = async (file?: File) => {
    if (!file || !p) return;
    if (!/^image\/(jpeg|png|webp)$/.test(file.type))
      return toast.error("Format gambar harus JPG, PNG, atau WebP.");
    if (file.size > 5 * 1024 * 1024)
      return toast.error("Ukuran gambar maksimal 5 MB.");
    setBusy(true);
    try {
      const { url } = await adminService.uploadImage(file);
      const img = await adminService.createImage(p.id, {
        image_url: url,
        sort_order: p.images.length,
        is_primary: p.images.length === 0,
      });
      setP({ ...p, images: [...p.images, img] });
    } catch (e) {
      toast.error(errorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const delImage = async (imageId: string) => {
    if (!p || !confirm("Hapus gambar ini?")) return;
    try {
      await adminService.deleteImage(p.id, imageId);
      setP({ ...p, images: p.images.filter((i) => i.id !== imageId) });
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  const title = isNew ? "Tambah Produk" : "Ubah Produk";
  const shown = finalPrice(f.price || 0, Number(f.discount_percent || 0));

  return (
    <div className="pb-24">
      <PageHead
        title={title}
        crumb={{ href: "/admin/produk", label: "Produk & Kategori" }}
        sub={
          p
            ? p.name +
              " · " +
              (cats.find((c) => c.id === p.category_id)?.name ??
                "Tanpa kategori")
            : undefined
        }
      />
      <div className="grid gap-6 sm:grid-cols-[1fr_262px] xl:grid-cols-[1fr_397px]">
        <div className="min-w-0 space-y-6">
          <Card title="Informasi produk">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="sm:col-span-2">
                <label htmlFor="name" className={labelCls}>
                  Nama produk <span className="text-danger">*</span>
                </label>
                <input
                  id="name"
                  required
                  value={f.name}
                  onChange={set("name")}
                  className={inputCls}
                />
              </div>
              <div>
                <label htmlFor="cat" className={labelCls}>
                  Kategori <span className="text-danger">*</span>
                </label>
                <select
                  id="cat"
                  value={f.category_id}
                  onChange={set("category_id")}
                  className={inputCls}
                >
                  <option value="">Pilih kategori</option>
                  {cats.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="weight" className={labelCls}>
                  Berat (gram) <span className="text-danger">*</span>
                </label>
                <input
                  id="weight"
                  type="number"
                  min={1}
                  inputMode="numeric"
                  value={f.weight_gram}
                  onChange={set("weight_gram")}
                  placeholder="g"
                  className={inputCls}
                />
              </div>
              <div className="sm:col-span-2">
                <label htmlFor="desc" className={labelCls}>
                  Deskripsi
                </label>
                <textarea
                  id="desc"
                  rows={4}
                  value={f.description}
                  onChange={set("description")}
                  className={inputCls + " h-auto py-2"}
                />
              </div>
            </div>
          </Card>

          <Card title="Harga">
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <label htmlFor="price" className={labelCls}>
                  Harga normal <span className="text-danger">*</span>
                </label>
                <input
                  id="price"
                  type="number"
                  min={1}
                  inputMode="numeric"
                  value={f.price}
                  onChange={set("price")}
                  className={inputCls}
                />
              </div>
              <div>
                <label htmlFor="disc" className={labelCls}>
                  Diskon (%)
                </label>
                <input
                  id="disc"
                  type="number"
                  min={0}
                  max={100}
                  value={f.discount_percent}
                  onChange={set("discount_percent")}
                  placeholder="%"
                  className={inputCls}
                />
              </div>
            </div>
            <p className="mt-4 flex items-center gap-2 rounded-lg bg-canvas px-3 py-2 text-xs">
              <Info className="size-4" aria-hidden /> Harga yang tampil ke
              pelanggan: <b>{formatRupiah(shown)}</b>
            </p>
          </Card>

          <Card title="Gambar Produk">
            {p ? (
              <div className="flex flex-wrap gap-3">
                {p.images.map((img) => (
                  <div key={img.id} className="relative">
                    <Thumb
                      src={img.image_url}
                      alt={p.name}
                      className="size-28 rounded-xl"
                    />
                    <button
                      type="button"
                      onClick={() => delImage(img.id)}
                      aria-label="Hapus gambar"
                      className="absolute right-1 top-1 rounded-full bg-white p-1 shadow"
                    >
                      <X className="size-3" aria-hidden />
                    </button>
                  </div>
                ))}
                <label className="grid size-28 cursor-pointer place-items-center rounded-xl border border-dashed border-faint bg-canvas text-center text-[10px] text-faint">
                  <span>
                    <Upload className="mx-auto mb-2 size-4" aria-hidden />
                    Tambah gambar
                    <br />
                    JPG atau PNG
                  </span>
                  <input
                    type="file"
                    accept="image/jpeg,image/png,image/webp"
                    className="sr-only"
                    disabled={busy}
                    onChange={(e) => addImage(e.target.files?.[0])}
                  />
                </label>
              </div>
            ) : (
              <p className="text-sm text-muted">
                Simpan produk terlebih dahulu untuk menambah gambar.
              </p>
            )}
          </Card>
        </div>

        <div className="space-y-6">
          <Card title="Tampil di website">
            <div className="flex items-center justify-between">
              <p className="text-lg">Produk aktif</p>
              <Toggle
                checked={f.active}
                onChange={(v) => setF((s) => ({ ...s, active: v }))}
                label="Produk aktif"
              />
            </div>
            <p className="mt-3 text-sm text-muted">
              Produk nonaktif tidak tampil kepada pelanggan.
            </p>
          </Card>
          {p && (
            <Card title="Ketersediaan">
              <AvailBadge type={p.availability_type} days={p.po_estimate_days} />
              <Link
                href={"/admin/ketersediaan?id=" + p.id}
                className={btnLine + " mt-3 block text-center"}
              >
                Atur Ready Stock / PO
              </Link>
              <p className="mt-4 flex justify-between border-t border-line-strong pt-4 text-sm">
                <span className="text-muted">Stok total</span>
                <b>{p.variants.reduce((n, v) => n + v.stock, 0)} pcs</b>
              </p>
              <Link
                href={"/admin/stok?id=" + p.id}
                className={btnLine + " mt-3 block text-center"}
              >
                Kelola Stok & Varian
              </Link>
            </Card>
          )}
        </div>
      </div>

      <div className="fixed inset-x-0 bottom-0 flex justify-end gap-6 border-t border-line bg-white px-4 py-4 sm:left-22.5 sm:px-6 lg:left-68.25 lg:px-10">
        <Link
          href="/admin/produk"
          className="self-center font-semibold underline"
        >
          Batal
        </Link>
        <button
          onClick={save}
          disabled={busy}
          className={btnLime + " px-5 py-3"}
        >
          {isNew ? "Simpan produk" : "Simpan perubahan"}
        </button>
      </div>
    </div>
  );
}
