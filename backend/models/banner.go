package models

// BannerPosition adalah slot tempat banner tampil di storefront (pilihan "Penempatan / Posisi Slot"
// di halaman admin banner).
type BannerPosition string

const (
	BannerHero         BannerPosition = "hero"         // carousel utama di beranda
	BannerAnnouncement BannerPosition = "announcement" // bar pengumuman di atas (teks berjalan)
	BannerMidFeed      BannerPosition = "mid_feed"     // promo di tengah beranda
	BannerCatalog      BannerPosition = "catalog"      // banner di atas halaman katalog
)

// Valid memeriksa apakah posisi termasuk salah satu slot yang dikenal.
func (p BannerPosition) Valid() bool {
	switch p {
	case BannerHero, BannerAnnouncement, BannerMidFeed, BannerCatalog:
		return true
	}
	return false
}

// Banner adalah satu banner yang sedang tayang. Jadwal tayang dan status aktif sudah disaring
// di repository, jadi tidak ikut dikirim ke publik.
type Banner struct {
	ID        int64          `json:"id"`
	Title     string         `json:"title"` // juga dipakai sebagai teks pengumuman dan alt gambar
	Position  BannerPosition `json:"position"`
	ImageURL  string         `json:"image_url"`
	LinkURL   string         `json:"link_url"`
	CTALabel  string         `json:"cta_label"`
	SortOrder int            `json:"sort_order"`
}
