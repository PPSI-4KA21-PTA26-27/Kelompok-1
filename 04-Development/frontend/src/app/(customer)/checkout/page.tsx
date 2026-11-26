"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Info, Truck } from "lucide-react";
import { useSelector } from "react-redux";
import { useAuth } from "@/app/context/AuthContext";
import Steps from "@/components/checkout/Steps";
import { cn } from "@/lib/cn";
import { formatRupiah, splitRTRW } from "@/lib/format";
import { cdn } from "@/lib/img";
import { selectCart } from "@/redux/features/cart-slice";
import { orderService, regionService } from "@/services";
import type { ApiRegion, ApiShippingQuote } from "@/types/api";

const input = "h-10 w-full rounded-lg border bg-white px-3 focus:outline-2 focus:outline-ink";
const card = "rounded-3xl border border-line bg-white p-7";

function useRegions(load: () => Promise<ApiRegion[]>, dep: number, on = true) {
  const [list, setList] = useState<ApiRegion[]>([]);
  useEffect(() => {
    if (!on) return;
    let off = false;
    load().then((d) => !off && setList(d)).catch(() => !off && setList([]));
    return () => {
      off = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dep]);
  return on ? list : [];
}

function Sel({ label, value, onChange, items, error }: { label: string; value: number; onChange: (v: number) => void; items: ApiRegion[]; error?: string }) {
  return (
    <label className="block">
      <span className="mb-2 block font-semibold">
        {label} <span className="text-danger">*</span>
      </span>
      <select
        value={value || ""}
        onChange={(e) => onChange(Number(e.target.value))}
        disabled={!items.length}
        className={cn(input, error ? "border-danger" : "border-line-strong", "disabled:opacity-60")}
      >
        <option value="">Pilih {label.toLowerCase()}</option>
        {items.map((i) => (
          <option key={i.id} value={i.id}>
            {i.name}
          </option>
        ))}
      </select>
      {error && <span className="mt-1 block text-xs text-danger">{error}</span>}
    </label>
  );
}

export default function CheckoutPage() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const { items, subtotal, loaded } = useSelector(selectCart);
  const [f, setF] = useState({ name: "", phone: "", address: "", rt: "", rw: "" });
  const [prov, setProv] = useState(0);
  const [city, setCity] = useState(0);
  const [dist, setDist] = useState(0);
  const [sub, setSub] = useState(0);
  const [errs, setErrs] = useState<Record<string, string>>({});
  const [apiErr, setApiErr] = useState("");
  const [busy, setBusy] = useState(false);
  const [quote, setQuote] = useState<ApiShippingQuote | null>(null);

  const provs = useRegions(() => regionService.provinces(), 0);
  const cities = useRegions(() => regionService.cities(prov), prov, !!prov);
  const dists = useRegions(() => regionService.districts(city), city, !!city);
  const subs = useRegions(() => regionService.subdistricts(dist), dist, !!dist);
  const ids = items.map((i) => i.id);

  useEffect(() => {
    if (!loading && !user) router.replace("/masuk?next=/checkout");
    else if (loaded && !items.length) router.replace("/keranjang");
  }, [loading, user, loaded, items.length, router]);

  // Isi otomatis dari pesanan terakhir (wilayah tetap dipilih ulang karena backend hanya menyimpan nama wilayah).
  useEffect(() => {
    if (!user) return;
    let off = false;
    orderService
      .listMine({ limit: 1 })
      .then((r) => (r.data[0] ? orderService.get(r.data[0].id) : null))
      .then((o) => {
        if (off) return;
        const base = { name: user.name, phone: user.phone ?? "", address: "", rt: "", rw: "" };
        setF((cur) =>
          cur.name || cur.phone || cur.address
            ? cur
            : o
              ? { ...base, name: o.recipient_name, phone: o.recipient_phone, ...splitRTRW(o.shipping_address) }
              : base,
        );
      })
      .catch(() => {});
    return () => {
      off = true;
    };
  }, [user]);

  useEffect(() => {
    let off = false;
    setQuote(null);
    const q = sub && ids.length ? regionService.shippingCost(sub, ids) : null;
    if (q) q.then((r) => !off && setQuote(r)).catch(() => !off && setQuote(null));
    return () => {
      off = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sub, ids.join(",")]);

  const shipping = Number(quote?.cost ?? 0);
  const total = Number(subtotal) + shipping;
  const set = (k: keyof typeof f) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => setF({ ...f, [k]: e.target.value });

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const need = "harus diisi";
    const er: Record<string, string> = {};
    if (!f.name.trim()) er.name = `Nama penerima ${need}`;
    if (!f.phone.trim()) er.phone = `Nomor telepon ${need}`;
    if (!prov) er.prov = `Provinsi ${need}`;
    if (!city) er.city = `Kota / Kabupaten ${need}`;
    if (!dist) er.dist = `Kecamatan ${need}`;
    if (!sub) er.sub = `Kelurahan ${need}`;
    if (!f.address.trim()) er.address = `Alamat lengkap ${need}`;
    if (!f.rt.trim()) er.rt = `RT ${need}`;
    else if (!/^\d{1,3}$/.test(f.rt.trim())) er.rt = "RT berupa angka, maksimal 3 digit";
    if (!f.rw.trim()) er.rw = `RW ${need}`;
    else if (!/^\d{1,3}$/.test(f.rw.trim())) er.rw = "RW berupa angka, maksimal 3 digit";
    setErrs(er);
    if (Object.keys(er).length) return;
    setBusy(true);
    setApiErr("");
    try {
      const o = await orderService.create({
        recipient_name: f.name.trim(),
        recipient_phone: f.phone.trim(),
        shipping_address: f.address.trim(),
        rt: f.rt.trim(),
        rw: f.rw.trim(),
        province_id: prov,
        city_id: city,
        district_id: dist,
        subdistrict_id: sub,
        cart_item_ids: ids,
      });
      router.push(`/pembayaran/${o.id}`);
    } catch (x) {
      setApiErr(x instanceof Error ? x.message : "Gagal membuat pesanan");
      setBusy(false);
    }
  }

  const text = (k: keyof typeof f, label: string, type = "text") => (
    <label className="block">
      <span className="mb-2 block font-semibold">
        {label} <span className="text-danger">*</span>
      </span>
      <input type={type} value={f[k]} onChange={set(k)} className={cn(input, errs[k] ? "border-danger" : "border-line-strong")} />
      {errs[k] && <span className="mt-1 block text-xs text-danger">{errs[k]}</span>}
    </label>
  );

  return (
    <div>
      <Steps current={2} />
      <h1 className="mt-4 text-4xl font-extrabold lg:mt-10 lg:text-5xl">Formulir Pemesanan</h1>
      <p className="mt-2 text-xl text-muted">Lengkapi data penerima dan alamat pengiriman.</p>

      <form noValidate onSubmit={submit} className="mt-6 grid items-start gap-4 sm:grid-cols-[1fr_260px] lg:mt-8 lg:grid-cols-[1fr_424px] lg:gap-8">
        <div className="space-y-7">
          <section className={card}>
            <h2 className="text-xl font-semibold">Data Penerima</h2>
            <div className="mt-6 grid gap-4 sm:grid-cols-2">
              {text("name", "Nama penerima")}
              {text("phone", "Nomor telepon", "tel")}
            </div>
          </section>

          <section className={card}>
            <h2 className="text-xl font-semibold">Alamat Pengiriman</h2>
            <div className="mt-6 grid gap-4 sm:grid-cols-2">
              <Sel label="Provinsi" value={prov} onChange={(v) => (setProv(v), setCity(0), setDist(0), setSub(0))} items={provs} error={errs.prov} />
              <Sel label="Kota / Kabupaten" value={city} onChange={(v) => (setCity(v), setDist(0), setSub(0))} items={cities} error={errs.city} />
              <Sel label="Kecamatan" value={dist} onChange={(v) => (setDist(v), setSub(0))} items={dists} error={errs.dist} />
              <Sel label="Kelurahan" value={sub} onChange={setSub} items={subs} error={errs.sub} />
              <div>
                <span className="mb-2 block font-semibold">Kode pos</span>
                <input readOnly value={subs.find((s) => s.id === sub)?.zip_code ?? ""} className={cn(input, "border-line-strong")} />
                <span className="mt-1 block text-xs text-muted">Terisi otomatis dari kelurahan</span>
              </div>
              <div className="grid grid-cols-2 gap-3">
                {(["rt", "rw"] as const).map((k) => (
                  <label key={k} className="block">
                    <span className="mb-2 block font-semibold">
                      {k.toUpperCase()} <span className="text-danger">*</span>
                    </span>
                    <input
                      inputMode="numeric"
                      maxLength={3}
                      required
                      aria-required="true"
                      aria-invalid={!!errs[k]}
                      placeholder="001"
                      value={f[k]}
                      onChange={(e) => setF({ ...f, [k]: e.target.value.replace(/\D/g, "") })}
                      className={cn(input, errs[k] ? "border-danger" : "border-line-strong")}
                    />
                    {errs[k] ? (
                      <span className="mt-1 block text-xs text-danger">{errs[k]}</span>
                    ) : (
                      <span className="mt-1 block text-xs text-muted">Angka, maks. 3 digit</span>
                    )}
                  </label>
                ))}
              </div>
            </div>
            <label className="mt-4 block">
              <span className="mb-2 block font-semibold">
                Alamat lengkap <span className="text-danger">*</span>
              </span>
              <textarea
                value={f.address}
                onChange={set("address")}
                rows={4}
                className={cn("w-full rounded-lg border bg-white p-3 focus:outline-2 focus:outline-ink", errs.address ? "border-danger" : "border-line-strong")}
              />
              {errs.address && <span className="mt-1 block text-xs text-danger">{errs.address}</span>}
            </label>
            <p className="mt-4 flex items-center gap-2 rounded-lg bg-canvas px-4 py-3 text-xs">
              <Info className="size-3.5 shrink-0" /> Pengiriman saat ini hanya ke alamat di Indonesia.
            </p>
          </section>

          <section className={card}>
            <h2 className="text-xl font-semibold">Pengiriman</h2>
            <div className="mt-5 flex items-center gap-4">
              <Truck className="size-8" />
              <div className="flex-1">
                <p className="font-semibold">{quote ? `${quote.courier_name} ${quote.service}` : "Kurir"}</p>
                <p className="text-muted">{quote ? `Estimasi ${quote.etd}` : "Dihitung dari kelurahan tujuan"}</p>
              </div>
              {quote && <p className="font-semibold">{formatRupiah(quote.cost)}</p>}
            </div>
          </section>
        </div>

        <aside className={card}>
          <h2 className="text-xl font-semibold">Ringkasan Pesanan</h2>
          <ul className="mt-5 space-y-4">
            {items.map((i) => (
              <li key={i.id} className="flex items-center gap-4">
                <div className="size-12.5 shrink-0 overflow-hidden rounded-lg bg-[#d9d9d9]">
                  {i.primary_image_url && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={cdn(i.primary_image_url, 100)} alt="" className="size-full object-cover" />
                  )}
                </div>
                <div className="flex-1 text-xs">
                  <p className="font-bold">{i.product_name}</p>
                  <p className="text-muted">{[i.size, i.color].filter(Boolean).join(" · ")} x {i.quantity}</p>
                </div>
                <p className="text-xs font-bold">{formatRupiah(i.subtotal)}</p>
              </li>
            ))}
          </ul>
          <dl className="mt-5 space-y-3 border-t border-line pt-5 text-muted">
            <div className="flex justify-between">
              <dt>Subtotal</dt>
              <dd>{formatRupiah(subtotal)}</dd>
            </div>
            <div className="flex justify-between">
              <dt>Ongkos kirim</dt>
              <dd>{quote ? formatRupiah(shipping) : "-"}</dd>
            </div>
          </dl>
          <div className="mt-4 flex justify-between border-t border-line pt-4 text-lg font-bold">
            <span>Total</span>
            <span className="text-xl">{formatRupiah(total)}</span>
          </div>
          {apiErr && (
            <p role="alert" className="mt-3 text-xs text-danger">
              {apiErr}
            </p>
          )}
          <button disabled={busy || !quote} className="mt-4 h-12 w-full rounded-xl bg-lime text-sm font-bold disabled:opacity-50">
            {busy ? "Memproses..." : "Buat Pesanan"}
          </button>
          <p className="mt-3 text-[10px] text-muted">Pembayaran melalui QRIS dan diverifikasi manual oleh Admin.</p>
        </aside>
      </form>
    </div>
  );
}
