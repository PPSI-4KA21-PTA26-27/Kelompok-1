package repository

import (
	"context"
	"database/sql"
	"fmt"

	"hii-stranger1/backend/models"
)

// Alur: services.CatalogService -> BannerRepository -> db.DB (PostgreSQL).
// Paket ini hanya berisi query SQL. Aturan bisnis ada di services.
//
// Kolom tabel yang dibaca (dibuat lewat migrasi, bukan di sini):
//   banners(id, title, position, image_url, link_url, cta_label, sort_order,
//           starts_at, ends_at, is_active)
// starts_at dan ends_at boleh NULL; NULL berarti "selalu tayang" (tanpa jadwal).

type BannerRepository struct {
	db *sql.DB
}

func NewBannerRepository(db *sql.DB) *BannerRepository {
	return &BannerRepository{db: db}
}

// ListActive mengambil banner yang saat ini tayang: aktif dan sedang dalam jadwal.
// Kalau position kosong, banner semua slot dikembalikan. Urutan mengikuti "Urutan Posisi" di admin.
// Hasil selalu berupa slice (kosong kalau tidak ada), supaya JSON-nya [] dan bukan null.
func (r *BannerRepository) ListActive(ctx context.Context, position models.BannerPosition) ([]models.Banner, error) {
	// NOW() bernilai sama di seluruh satu query. $1::text dipakai dua kali, jadi tipenya
	// ditulis eksplisit supaya PostgreSQL tidak bingung menebaknya.
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, position, COALESCE(image_url, ''), COALESCE(link_url, ''),
		        COALESCE(cta_label, ''), sort_order
		 FROM banners
		 WHERE is_active
		   AND (starts_at IS NULL OR starts_at <= NOW())
		   AND (ends_at IS NULL OR ends_at > NOW())
		   AND ($1::text = '' OR position = $1::text)
		 ORDER BY sort_order, id`, string(position))
	if err != nil {
		return nil, fmt.Errorf("ambil banner: %w", err)
	}
	defer rows.Close()

	banners := []models.Banner{}
	for rows.Next() {
		var b models.Banner
		if err := rows.Scan(&b.ID, &b.Title, &b.Position, &b.ImageURL, &b.LinkURL, &b.CTALabel, &b.SortOrder); err != nil {
			return nil, fmt.Errorf("baca banner: %w", err)
		}
		banners = append(banners, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar banner: %w", err)
	}
	return banners, nil
}
