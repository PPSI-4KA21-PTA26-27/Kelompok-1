// Package tracking menyusun tautan pelacakan resi untuk kurir pengiriman.
package tracking

import (
	"net/url"
	"strings"
)

// Link adalah tautan pelacakan yang bisa dibuka pelanggan.
type Link struct {
	Label string
	URL   string
}

// couriers memetakan kata kunci nama/kode kurir ke halaman lacak resmi.
// Urutan penting: kata kunci yang lebih spesifik dicek lebih dulu.
var couriers = []struct {
	key, name, url string
}{
	{"sicepat", "SiCepat", "https://www.sicepat.com/checkAwb"},
	{"anteraja", "AnterAja", "https://anteraja.id/tracking"},
	{"ninja", "Ninja Xpress", "https://www.ninjaxpress.co/id-id/tracking"},
	{"lion", "Lion Parcel", "https://lionparcel.com/track/stt"},
	{"j&t", "J&T Express", "https://jet.co.id/track"},
	{"jnt", "J&T Express", "https://jet.co.id/track"},
	{"jne", "JNE", "https://www.jne.co.id/tracking-package"},
	{"tiki", "TIKI", "https://www.tiki.id/id/track"},
	{"pos", "POS Indonesia", "https://www.posindonesia.co.id/id/tracking"},
	{"wahana", "Wahana", "https://www.wahana.com/"},
}

// UniversalURL membuka cekresi.com dengan no. resi yang sudah terisi; bekerja
// untuk hampir semua kurir di Indonesia.
func UniversalURL(resi string) string {
	return "https://cekresi.com/?noresi=" + url.QueryEscape(strings.TrimSpace(resi))
}

// Links mengembalikan tautan lacak: halaman resmi kurir (bila dikenal) lalu
// pelacak universal. Kosong bila no. resi kosong.
func Links(courier, resi string) []Link {
	resi = strings.TrimSpace(resi)
	if resi == "" {
		return nil
	}
	var out []Link
	c := strings.ToLower(courier)
	for _, k := range couriers {
		if strings.Contains(c, k.key) {
			out = append(out, Link{Label: "Lacak di situs " + k.name, URL: k.url})
			break
		}
	}
	return append(out, Link{Label: "Lacak cepat via cekresi.com", URL: UniversalURL(resi)})
}
