"use client";

import { useEffect, useRef, useState } from "react";
import toast from "react-hot-toast";
import { Image as ImageIcon, Trash2, Upload } from "lucide-react";
import { Card, PageHead, Toggle, btnDanger, btnLime, btnLine, inputCls, labelCls } from "@/components/admin/ui";
import { errorMessage } from "@/lib/api";
import { cdn } from "@/lib/img";
import { adminService } from "@/services";
import type { ApiBanner, BannerInput, BannerPlacement } from "@/types/api";

const PLACEMENT: Record<BannerPlacement, string> = { hero: "Beranda (hero)", main: "Beranda (utama)", catalog: "Katalog" };

function checkFile(file: File) {
  if (!/^image\/(jpeg|png|webp)$/.test(file.type)) return "Format gambar harus JPG, PNG, atau WebP.";
  if (file.size > 5 * 1024 * 1024) return "Ukuran gambar maksimal 5 MB.";
  return "";
}

export default function BannerPage() {
  const [list, setList] = useState<ApiBanner[]>([]);
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState("");
  const [link, setLink] = useState("");
  const [placement, setPlacement] = useState<BannerPlacement>("hero");
  const [busy, setBusy] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);

  useEffect(() => {
    adminService
      .banners()
      .then((b) => setList([...b].sort((x, y) => x.sort_order - y.sort_order)))
      .catch((e) => toast.error(errorMessage(e)));
  }, []);

  const pickFile = (f: File | null) => {
    if (preview) URL.revokeObjectURL(preview);
    setFile(f);
    setPreview(f ? URL.createObjectURL(f) : "");
  };

  const patch = async (b: ApiBanner, input: Partial<BannerInput>) => {
    try {
      const u = await adminService.updateBanner(b.id, input);
      setList((ls) => ls.map((x) => (x.id === b.id ? u : x)));
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  const replaceImage = async (b: ApiBanner, f?: File) => {
    if (!f) return;
    const bad = checkFile(f);
    if (bad) return toast.error(bad);
    try {
      const { url } = await adminService.uploadImage(f);
      await patch(b, { image_url: url });
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  const remove = async (b: ApiBanner) => {
    if (!confirm("Hapus banner ini?")) return;
    try {
      await adminService.deleteBanner(b.id);
      setList((ls) => ls.filter((x) => x.id !== b.id));
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  const reset = () => {
    pickFile(null);
    setLink("");
  };

  const create = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return toast.error("Pilih gambar banner.");
    const bad = checkFile(file);
    if (bad) return toast.error(bad);
    const l = link.trim();
    if (l && !l.startsWith("/") && !/^https?:\/\//.test(l)) return toast.error("Tautan harus diawali / atau http(s)://");
    setBusy(true);
    try {
      const { url } = await adminService.uploadImage(file);
      const b = await adminService.createBanner({
        placement,
        title: null,
        subtitle: null,
        cta_label: null,
        image_url: url,
        mobile_image_url: null,
        link_url: l || null,
        sort_order: list.reduce((m, x) => Math.max(m, x.sort_order), 0) + 1,
        is_active: true,
        starts_at: null,
        ends_at: null,
      });
      setList((ls) => [...ls, b]);
      reset();
      toast.success("Banner disimpan.");
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div>
      <PageHead
        title="Kelola Banner"
        sub="Banner tampil di Beranda sesuai urutan."
        action={
          <button onClick={() => formRef.current?.querySelector("input")?.focus()} className={btnLime + " px-5 py-3"}>
            + Tambah Banner
          </button>
        }
      />
      <div className="grid gap-6 sm:grid-cols-[1fr_220px] xl:grid-cols-[1fr_400px]">
        <div className="min-w-0 space-y-6">
          {list.length === 0 && <Card><p className="text-sm text-muted">Belum ada banner.</p></Card>}
          {list.map((b) => (
            <Card key={b.id} className="flex gap-4 xl:gap-5">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={cdn(b.image_url, 600)} alt={b.title ?? "Banner"} className="aspect-[244/140] w-32 self-start rounded-xl bg-[#d9d9d9] object-cover sm:w-36 xl:w-60" />
              <div className="flex flex-1 flex-col">
                <div className="flex items-start justify-between gap-3">
                  <div className="text-sm">
                    <b className="text-base">{b.title || PLACEMENT[b.placement]}</b>
                    <p>{b.link_url || "Tanpa tautan"}</p>
                    <p>Urutan {b.sort_order} · {PLACEMENT[b.placement]}</p>
                  </div>
                  <Toggle checked={b.is_active} onChange={(v) => patch(b, { is_active: v })} label={"Aktifkan banner urutan " + b.sort_order} />
                </div>
                <div className="mt-auto flex flex-wrap justify-end gap-2 pt-4">
                  <label className={btnLine + " flex cursor-pointer items-center gap-2"}>
                    <ImageIcon className="size-4" aria-hidden /> Ganti gambar
                    <input type="file" accept="image/jpeg,image/png,image/webp" className="sr-only" onChange={(e) => replaceImage(b, e.target.files?.[0])} />
                  </label>
                  <button onClick={() => remove(b)} className={btnDanger + " flex items-center gap-2"}>
                    <Trash2 className="size-4" aria-hidden /> Hapus
                  </button>
                </div>
              </div>
            </Card>
          ))}
        </div>

        <Card title="Banner baru" className="self-start">
          <form ref={formRef} onSubmit={create} className="space-y-4">
            <label className="grid min-h-32 cursor-pointer place-items-center overflow-hidden rounded-lg border border-dashed border-ink-3 bg-[#e9e9e9] text-center">
              {preview ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={preview} alt="Pratinjau banner" className="max-h-48 w-full object-cover" />
              ) : (
                <span className="py-6 text-sm font-semibold">
                  <Upload className="mx-auto mb-2 size-7 text-muted" aria-hidden />
                  Pilih gambar
                  <span className="block text-[10px] font-normal text-muted">JPG atau PNG</span>
                </span>
              )}
              <input type="file" accept="image/jpeg,image/png,image/webp" className="sr-only" onChange={(e) => pickFile(e.target.files?.[0] ?? null)} />
            </label>
            <div>
              <label htmlFor="placement" className={labelCls}>
                Posisi
              </label>
              <select id="placement" value={placement} onChange={(e) => setPlacement(e.target.value as BannerPlacement)} className={inputCls}>
                {Object.entries(PLACEMENT).map(([v, l]) => (
                  <option key={v} value={v}>
                    {l}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="link" className={labelCls}>
                Tautan tujuan
              </label>
              <input id="link" value={link} onChange={(e) => setLink(e.target.value)} placeholder="/katalog" className={inputCls} />
              <p className="mt-2 text-[10px] text-muted">Boleh dikosongkan.</p>
            </div>
            <div className="flex justify-end gap-6">
              <button type="button" onClick={reset} className="font-semibold underline">
                Batal
              </button>
              <button disabled={busy} className={btnLime + " px-5 py-3"}>
                Simpan Banner
              </button>
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
}
