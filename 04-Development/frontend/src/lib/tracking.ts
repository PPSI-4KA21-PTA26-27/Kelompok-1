// Sama dengan backend/internal/pkg/tracking: halaman lacak resmi kurir +
// pelacak universal cekresi.com dengan no. resi terisi.
const COURIERS: { key: string; name: string; url: string }[] = [
  { key: "sicepat", name: "SiCepat", url: "https://www.sicepat.com/checkAwb" },
  { key: "anteraja", name: "AnterAja", url: "https://anteraja.id/tracking" },
  { key: "ninja", name: "Ninja Xpress", url: "https://www.ninjaxpress.co/id-id/tracking" },
  { key: "lion", name: "Lion Parcel", url: "https://lionparcel.com/track/stt" },
  { key: "j&t", name: "J&T Express", url: "https://jet.co.id/track" },
  { key: "jnt", name: "J&T Express", url: "https://jet.co.id/track" },
  { key: "jne", name: "JNE", url: "https://www.jne.co.id/tracking-package" },
  { key: "tiki", name: "TIKI", url: "https://www.tiki.id/id/track" },
  { key: "pos", name: "POS Indonesia", url: "https://www.posindonesia.co.id/id/tracking" },
  { key: "wahana", name: "Wahana", url: "https://www.wahana.com/" },
];

export type TrackingLink = { label: string; url: string; official: boolean };

export function trackingLinks(courier: string | null | undefined, resi: string | null | undefined): TrackingLink[] {
  const r = (resi ?? "").trim();
  if (!r) return [];
  const c = (courier ?? "").toLowerCase();
  const found = COURIERS.find((k) => c.includes(k.key));
  const out: TrackingLink[] = [];
  if (found) out.push({ label: "Situs " + found.name, url: found.url, official: true });
  out.push({ label: "Lacak via cekresi.com", url: "https://cekresi.com/?noresi=" + encodeURIComponent(r), official: false });
  return out;
}
