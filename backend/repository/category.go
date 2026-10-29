package repository

import (
	"context"
	"database/sql"
	"fmt"

	"hii-stranger1/backend/models"
)

// Alur: services.CatalogService -> CategoryRepository -> db.DB (PostgreSQL).
// Paket ini hanya berisi query SQL. Aturan bisnis ada di services.
//
// Kolom tabel yang dibaca (dibuat lewat migrasi, bukan di sini):
//   categories(id, name, slug, description, sort_order, is_active)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// ListActive mengambil kategori yang aktif, diurutkan sesuai "Urutan Posisi" di admin.
// Hasil selalu berupa slice (kosong kalau tidak ada data), supaya JSON-nya [] dan bukan null.
func (r *CategoryRepository) ListActive(ctx context.Context) ([]models.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, slug, COALESCE(description, ''), sort_order
		 FROM categories
		 WHERE is_active
		 ORDER BY sort_order, name`)
	if err != nil {
		return nil, fmt.Errorf("ambil kategori: %w", err)
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.SortOrder); err != nil {
			return nil, fmt.Errorf("baca kategori: %w", err)
		}
		categories = append(categories, c)
	}
	// rows.Next() bisa berhenti karena error di tengah jalan, bukan hanya karena data habis.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar kategori: %w", err)
	}
	return categories, nil
}
