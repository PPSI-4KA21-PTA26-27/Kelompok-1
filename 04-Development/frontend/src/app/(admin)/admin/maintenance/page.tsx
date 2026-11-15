"use client";

import { useEffect, useState } from "react";
import toast from "react-hot-toast";
import { Card, PageHead, Toggle, labelCls } from "@/components/admin/ui";
import { errorMessage } from "@/lib/api";
import { adminService, settingService } from "@/services";
import type { ApiSiteSettings } from "@/types/api";

export default function MaintenancePage() {
  const [s, setS] = useState<ApiSiteSettings | null>(null);
  const [on, setOn] = useState(false);
  const [msg, setMsg] = useState("");
  const [busy, setBusy] = useState(false);

  const sync = (x: ApiSiteSettings) => {
    setS(x);
    setOn(x.maintenance_mode);
    setMsg(x.maintenance_message);
  };

  useEffect(() => {
    settingService.get().then(sync).catch((e) => toast.error(errorMessage(e)));
  }, []);

  if (!s) return null;
  const dirty = on !== s.maintenance_mode || msg !== s.maintenance_message;

  const save = async () => {
    if (!msg.trim()) return toast.error("Pesan untuk pelanggan wajib diisi.");
    setBusy(true);
    try {
      sync(await adminService.updateSettings({ maintenance_mode: on, maintenance_message: msg.trim() }));
      toast.success("Pengaturan disimpan.");
    } catch (e) {
      // gagal simpan: status sebelumnya dipertahankan
      sync(s);
      toast.error(errorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div>
      <PageHead title="Pengaturan Maintenance" sub="Atur akses pelanggan saat website diperbarui." />
      <div className="grid gap-8 sm:grid-cols-[1fr_230px] xl:grid-cols-[1fr_414px]">
        <div className="min-w-0 space-y-7">
          <Card title="Mode maintenance" action={<Toggle checked={on} onChange={setOn} label="Mode maintenance" />} className="p-8">
            <p className="-mt-2 border-b border-ink pb-5 text-muted">
              Status saat ini: <b className="text-ink">{s.maintenance_mode ? "Aktif" : "Nonaktif"}</b>.{" "}
              {s.maintenance_mode ? "Pelanggan tidak dapat mengakses website." : "Pelanggan dapat mengakses website."}
            </p>
            <label htmlFor="msg" className={labelCls + " mt-5"}>
              Pesan untuk pelanggan
            </label>
            <textarea
              id="msg"
              rows={4}
              maxLength={500}
              value={msg}
              onChange={(e) => setMsg(e.target.value)}
              className="w-full rounded-xl border border-line-strong p-3 focus:outline-2 focus:outline-ink"
            />
            <p className="mt-2 text-[10px] text-muted">Pesan ditampilkan di halaman maintenance.</p>
          </Card>

          {dirty && (
            <section className="rounded-xl border border-ink bg-white p-6" aria-live="polite">
              <p className="rounded-lg bg-warn-soft px-11 py-2 text-xs text-warn">
                {on !== s.maintenance_mode ? (
                  on ? (
                    <>
                      <b>Aktifkan mode maintenance?</b> Pelanggan tidak dapat membuka website. Admin dan owner tetap dapat masuk.
                    </>
                  ) : (
                    <>
                      <b>Nonaktifkan mode maintenance?</b> Pelanggan dapat membuka website kembali.
                    </>
                  )
                ) : (
                  <b>Simpan pesan maintenance?</b>
                )}
              </p>
              <div className="mt-4 flex justify-end gap-6">
                <button onClick={() => sync(s)} className="font-semibold underline">
                  Batal
                </button>
                <button onClick={save} disabled={busy} className="rounded-lg bg-ink px-4 py-3 font-bold text-white disabled:opacity-50">
                  {on !== s.maintenance_mode ? (on ? "Ya, Aktifkan" : "Ya, Nonaktifkan") : "Simpan"}
                </button>
              </div>
            </section>
          )}
        </div>

        <Card title="Pratinjau untuk pelanggan" className="self-start p-7">
          <div className="grid min-h-40 place-items-center rounded-xl bg-[#d9d9d9] p-6 text-center text-sm">
            {msg.trim() || "Website sedang dalam perbaikan."}
          </div>
          <p className="mt-5 rounded-xl bg-canvas px-12 py-3 text-xs text-muted">Jika penyimpanan gagal, status sebelumnya dipertahankan.</p>
        </Card>
      </div>
    </div>
  );
}
