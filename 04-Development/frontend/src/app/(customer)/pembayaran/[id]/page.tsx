"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";
import { CircleAlert, X } from "lucide-react";
import Steps from "@/components/checkout/Steps";
import Countdown from "@/components/order/Countdown";
import { useCartActions } from "@/hooks/useCartActions";
import { formatDate, formatRupiah } from "@/lib/format";
import { orderService, paymentService } from "@/services";
import type { ApiOrderDetail, ApiPayment } from "@/types/api";

const card = "rounded-3xl border border-line bg-white p-7";
const HOW = [
  "Buka aplikasi pembayaran yang mendukung QRIS.",
  "Pindai kode di atas.",
  "Bayar tepat sesuai nominal beserta kode unik.",
  "Unggah bukti pembayaran di samping.",
];

export default function PembayaranPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const [order, setOrder] = useState<ApiOrderDetail | null>(null);
  const [pay, setPay] = useState<ApiPayment | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);
  const pick = useRef<HTMLInputElement>(null);
  const { refresh } = useCartActions();
  // checkout menghapus item keranjang di server; badge navbar disegarkan di sini (di checkout akan memicu redirect ke /keranjang)
  useEffect(() => {
    refresh().catch(() => {});
  }, [refresh]);
  const preview = useMemo(() => (file ? URL.createObjectURL(file) : ""), [file]);
  useEffect(() => () => URL.revokeObjectURL(preview), [preview]);

  useEffect(() => {
    orderService
      .get(id)
      .then(async (o) => {
        setOrder(o);
        setPay(o.payment ?? (await paymentService.create(id)));
      })
      .catch((x) => setErr(x instanceof Error ? x.message : "Gagal memuat pesanan"));
  }, [id]);

  const choose = (f?: File) => {
    if (!f) return;
    // file baru yang ditolak mengosongkan pilihan lama
    const reject = (m: string) => (setFile(null), setErr(m));
    if (!/^image\/(jpeg|png)$/.test(f.type)) return reject("Format harus JPG atau PNG");
    // batas backend maxProofBytes = 2 MB
    if (f.size > 2 * 1024 * 1024) return reject("Ukuran gambar maksimal 2 MB");
    setErr("");
    setFile(f);
  };
  async function send() {
    if (!file) return;
    setBusy(true);
    setErr("");
    try {
      await paymentService.uploadProof(id, file);
      router.push(`/pesanan/${id}`);
    } catch (x) {
      setErr(x instanceof Error ? x.message : "Gagal mengirim bukti");
      setBusy(false);
    }
  }

  const amount = pay ? formatRupiah(pay.amount) : "";
  const sent = !!pay?.has_proof && pay.status === "menunggu";

  return (
    <div>
      <Steps current={3} />
      <h1 className="mt-4 text-4xl font-extrabold lg:mt-10 lg:text-5xl">Pembayaran QRIS</h1>
      <p className="mt-2 text-xl text-muted">
        Pesanan {order?.order_number ?? "…"} · selesaikan pembayaran dengan memindai kode QRIS.
      </p>
      {err && !pay && <p role="alert" className="mt-6 text-danger">{err}</p>}

      <div className="mt-6 grid items-start gap-4 sm:grid-cols-[1fr_245px] lg:mt-8 lg:grid-cols-[1fr_433px] lg:gap-8">
        <div className="space-y-7">
          <section className={card}>
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold">Kode QRIS Hii Stranger</h2>
              <span className="rounded-full bg-warn-soft px-2.5 py-1 text-xs font-bold text-warn">
                {pay?.status === "ditolak" ? "Bukti ditolak" : pay?.has_proof ? "Menunggu verifikasi" : "Belum dibayar"}
              </span>
            </div>
            <div className="mt-6 rounded-2xl border-2 border-ink px-3 py-6 text-center lg:px-6 lg:py-10">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src="/qris.jpg" alt="QRIS Hii Stranger" className="mx-auto h-auto w-full max-w-65 rounded-lg object-contain" />
              <p className="mt-8 text-muted">Pindai dengan aplikasi pembayaran yang mendukung QRIS</p>
              <p className="mt-8 text-muted">Nominal yang harus dibayar</p>
              <p className="mt-2 text-3xl font-extrabold lg:text-4xl">
                {amount}
              </p>
              {order && pay && (
                <p className="mt-2 text-muted">
                  Total pesanan {formatRupiah(order.total_amount)} + kode unik <mark className="rounded bg-lime px-1 font-bold text-ink">{formatRupiah(pay.unique_code)}</mark>
                </p>
              )}
              {order?.payment_due_at && !pay?.has_proof && (
                <div className="mt-4 flex flex-col items-center gap-1.5">
                  <Countdown due={order.payment_due_at} className="text-sm" />
                  <p className="text-xs text-muted">Bayar sebelum {formatDate(order.payment_due_at, true)}, lewat dari itu pesanan otomatis batal.</p>
                </div>
              )}
            </div>
          </section>

          <section className={card}>
            <h2 className="text-xl font-semibold">Cara membayar</h2>
            <ol className="mt-5 space-y-4">
              {HOW.map((h, i) => (
                <li key={h} className="flex items-center gap-4 text-sm">
                  <span className="grid size-6 place-items-center rounded-full bg-ink text-xs font-bold text-lime">{i + 1}</span>
                  {h}
                </li>
              ))}
            </ol>
          </section>
        </div>

        <div className="space-y-7">
          {order && order.order_status !== "pending_payment" ? (
            <section className={card}>
              <h2 className="text-xl font-semibold">Unggah Bukti Pembayaran</h2>
              <p className="mt-3 text-sm text-muted">Pesanan ini sudah tidak menunggu pembayaran.</p>
            </section>
          ) : (
          <section className={card}>
            <h2 className="text-xl font-semibold">Unggah Bukti Pembayaran</h2>
            {sent && (
              <p role="status" className="mt-3 rounded-lg bg-lime/30 p-3 text-sm">
                <b>Bukti sudah terkirim</b> dan menunggu verifikasi admin. Unggah lagi hanya jika berkas tadi salah.
              </p>
            )}
            {pay?.status === "ditolak" && (
              <p role="alert" className="mt-3 rounded-lg bg-danger-soft p-3 text-sm text-danger">
                Bukti sebelumnya ditolak. Unggah bukti pembayaran yang benar.
              </p>
            )}
            <button
              type="button"
              onClick={() => pick.current?.click()}
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => (e.preventDefault(), choose(e.dataTransfer.files[0]))}
              className="mt-5 grid h-46.25 w-full place-items-center rounded-2xl border-2 border-dashed border-muted text-muted"
            >
              <span>
                <b className="text-ink">Pilih berkas</b> atau tarik ke sini
                <br />
                Format JPG atau PNG, maks 2 MB
              </span>
            </button>
            <input ref={pick} type="file" accept="image/jpeg,image/png" hidden onChange={(e) => choose(e.target.files?.[0])} />
            {file && (
              <div className="mt-4 flex items-center gap-4">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={preview} alt="" className="size-12.5 rounded-lg bg-[#d9d9d9] object-cover" />
                <span className="flex-1 truncate text-xs font-semibold">{file.name}</span>
                <button aria-label="Hapus berkas" onClick={() => setFile(null)}>
                  <X className="size-4" />
                </button>
              </div>
            )}
            {err && pay && <p role="alert" className="mt-3 text-xs text-danger">{err}</p>}
            <button onClick={send} disabled={!file || busy} className="mt-4 h-12 w-full rounded-xl bg-lime text-sm font-bold disabled:opacity-50">
              {busy ? "Mengirim..." : sent ? "Ganti Bukti Pembayaran" : "Kirim Bukti Pembayaran"}
            </button>
          </section>
          )}

          <section className={card}>
            <p className="flex gap-2 rounded-lg border border-warn bg-warn-soft p-3 text-xs text-warn">
              <CircleAlert className="mt-0.5 size-3 shrink-0" />
              <span>
                <b>Pembayaran belum dianggap berhasil.</b> Admin akan memeriksa bukti Anda. Status akan berubah ketika telah diverifikasi.
              </span>
            </p>
            <Link href={`/pesanan/${id}`} className="mt-4 block rounded-lg border border-ink py-2 text-center text-xs font-bold">
              Lihat Status Pesanan
            </Link>
          </section>
        </div>
      </div>
    </div>
  );
}
