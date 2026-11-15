"use client";

import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Clock3 } from "lucide-react";
import {
  Card,
  PageHead,
  Thumb,
  btnLime,
  btnLine,
  inputCls,
  labelCls,
  stackTable,
  theadCls,
} from "@/components/admin/ui";
import { errorMessage } from "@/lib/api";
import { cn } from "@/lib/cn";
import { adminService } from "@/services";
import type { ApiProduct } from "@/types/api";

export default function KetersediaanPage() {
  return (
    <Suspense>
      <Ketersediaan />
    </Suspense>
  );
}

function Ketersediaan() {
  const initial = useSearchParams().get("id") ?? "";
  const [list, setList] = useState<ApiProduct[]>([]);
  const [id, setId] = useState(initial);
  const [po, setPo] = useState(false);
  const [days, setDays] = useState("");
  const [busy, setBusy] = useState(false);

  function apply(x?: ApiProduct) {
    setId(x?.id ?? "");
    setPo(x?.availability_type === "preorder");
    setDays(x?.po_estimate_days != null ? String(x.po_estimate_days) : "");
  }

  useEffect(() => {
    // shortcut: 100 produk pertama; tambah paginasi bila katalog lebih besar
    adminService
      .products({ limit: 100, sort: "name" })
      .then((r) => {
        setList(r.data);
        apply(r.data.find((l) => l.id === initial));
      })
      .catch((e) => toast.error(errorMessage(e)));
  }, [initial]);

  const pick = (pid: string) => apply(list.find((l) => l.id === pid));

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id) return toast.error("Pilih produk terlebih dahulu.");
    const n = Number(days);
    if (po && (!/^\d+$/.test(days) || n < 1))
      return toast.error("Estimasi pre-order minimal 1 hari.");
    setBusy(true);
    try {
      const u = await adminService.updateProduct(id, {
        availability_type: po ? "preorder" : "ready",
        po_estimate_days: po ? n : null,
      });
      setList((ls) => ls.map((l) => (l.id === id ? { ...l, ...u } : l)));
      toast.success("Ketersediaan disimpan.");
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setBusy(false);
    }
  };

  const seg = (on: boolean) =>
    cn(
      "flex-1 rounded-lg py-3 text-sm font-bold",
      on ? "bg-ink text-lime" : "text-ink",
    );

  return (
    <div>
      <PageHead
        title="Ready Stock / Pre-Order"
        sub="Tentukan jenis ketersediaan dan estimasi waktu pre-order."
      />
      <div className="grid gap-6 sm:grid-cols-[1fr_220px] xl:grid-cols-[1fr_403px]">
        <div className="min-w-0 space-y-8">
          <Card title="Pengaturan produk">
            <form onSubmit={save} className="space-y-6">
              <div>
                <label htmlFor="prod" className={labelCls}>
                  Produk *
                </label>
                <select
                  id="prod"
                  value={id}
                  onChange={(e) => pick(e.target.value)}
                  className={inputCls}
                >
                  <option value="">Pilih produk</option>
                  {list.map((l) => (
                    <option key={l.id} value={l.id}>
                      {l.name}
                    </option>
                  ))}
                </select>
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
                    value={days}
                    onChange={(e) => setDays(e.target.value)}
                    className={inputCls}
                  />
                </div>
              )}
              <div className="flex justify-end gap-6">
                <button
                  type="button"
                  onClick={() => pick(id)}
                  className="font-semibold underline"
                >
                  Batal
                </button>
                <button disabled={busy} className={btnLime + " px-5 py-3"}>
                  Simpan
                </button>
              </div>
            </form>
          </Card>

          <section>
            <h2 className="mb-3 text-lg font-bold">Semua produk</h2>
            <div className="overflow-x-auto rounded-xl border border-line-strong bg-white">
              <table className={"w-full text-xs sm:min-w-[560px] " + stackTable}>
                <thead className={theadCls}>
                  <tr>
                    {["Produk", "Jenis", "Estimasi PO", ""].map((h, i) => (
                      <th key={i} className="px-5 py-4 font-semibold">
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {list.map((l) => (
                    <tr key={l.id} className="border-t border-line-strong">
                      <td className="px-5 py-3">
                        <div className="flex items-center gap-3">
                          <Thumb
                            src={l.primary_image_url}
                            alt={l.name}
                            className="size-9"
                          />
                          <b>{l.name}</b>
                        </div>
                      </td>
                      <td data-l="Jenis" className="px-5 py-3">
                        <span className="rounded-full bg-ink px-2 py-1 font-bold text-white">
                          {l.availability_type === "preorder"
                            ? "Pre-Order"
                            : "Ready Stock"}
                        </span>
                      </td>
                      <td data-l="Estimasi PO" className="px-5 py-3">
                        {l.availability_type === "preorder" &&
                        l.po_estimate_days != null
                          ? l.po_estimate_days + " hari"
                          : "-"}
                      </td>
                      <td className="px-5 py-3 text-right">
                        <button
                          onClick={() => {
                            pick(l.id);
                            window.scrollTo({ top: 0, behavior: "smooth" });
                          }}
                          className={btnLine}
                        >
                          Atur
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <Card title="Tampilan di halaman produk" className="self-start">
          <span className="rounded-full bg-ink px-2 py-0.5 text-[10px] font-bold text-white">
            {po ? "Pre-Order" : "Ready Stock"}
          </span>
          {po && (
            <p className="mt-2 flex items-center gap-2 rounded-lg bg-warn-soft px-4 py-2 text-xs text-warn">
              <Clock3 className="size-3" aria-hidden /> Estimasi pre-order{" "}
              {days || "-"} hari.
            </p>
          )}
          <p className="mt-2 text-xs text-muted">
            Produk ready stock tidak menampilkan estimasi.
          </p>
        </Card>
      </div>
    </div>
  );
}
