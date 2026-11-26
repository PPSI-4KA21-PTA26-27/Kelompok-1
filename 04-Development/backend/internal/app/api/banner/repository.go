package banner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// bannerColumns membaca kode penempatan dari banner_placements
// (banners.placement_id) sehingga JSON API tetap memakai placement dan
// image_url. Kolom gambar desktop di basis data bernama desktop_image_url.
const bannerColumns = `id, (SELECT bp.code FROM banner_placements bp WHERE bp.id = banners.placement_id) AS placement,
	title, subtitle, cta_label, desktop_image_url, mobile_image_url, link_url,
	sort_order, is_active, starts_at, ends_at, created_at`

// placementID mengubah kode penempatan menjadi id di banner_placements.
const placementID = `(SELECT id FROM banner_placements WHERE code = $%d)`

// Repository mengakses tabel banners.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// List mengembalikan banner. Tanpa All hanya banner aktif yang sedang
// dalam jadwal tayang.
func (r *Repository) List(ctx context.Context, f ListFilter) ([]Banner, error) {
	var conds []string
	var args []any
	if !f.All {
		conds = append(conds, "is_active", "(starts_at IS NULL OR starts_at <= now())", "(ends_at IS NULL OR ends_at > now())")
	}
	if f.Placement != "" {
		args = append(args, f.Placement)
		conds = append(conds, fmt.Sprintf("placement_id = "+placementID, len(args)))
	}
	query := "SELECT " + bannerColumns + " FROM banners"
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	rows, err := r.db.Query(ctx, query+" ORDER BY placement, sort_order, created_at DESC, id", args...)
	if err != nil {
		return nil, fmt.Errorf("banner: list: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	out := make([]Banner, 0)
	for rows.Next() {
		b, err := scanBanner(rows)
		if err != nil {
			return nil, fmt.Errorf("banner: list scan: %w", err)
		}
		b.computeStatus(now)
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("banner: list rows: %w", err)
	}
	return out, nil
}

// Create menyimpan banner baru.
func (r *Repository) Create(ctx context.Context, b Banner) (Banner, error) {
	created, err := scanBanner(r.db.QueryRow(ctx,
		`INSERT INTO banners (placement_id, title, subtitle, cta_label, desktop_image_url, mobile_image_url,
		                      link_url, sort_order, is_active, starts_at, ends_at)
		 VALUES ((SELECT id FROM banner_placements WHERE code = $1), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING `+bannerColumns,
		b.Placement, b.Title, b.Subtitle, b.CTALabel, b.ImageURL, b.MobileImageURL,
		b.LinkURL, b.SortOrder, b.IsActive, b.StartsAt, b.EndsAt))
	if err != nil {
		return Banner{}, fmt.Errorf("banner: create: %w", err)
	}
	created.computeStatus(time.Now())
	return created, nil
}

// Get mengambil satu banner.
func (r *Repository) Get(ctx context.Context, id string) (Banner, error) {
	b, err := scanBanner(r.db.QueryRow(ctx, "SELECT "+bannerColumns+" FROM banners WHERE id = $1", id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Banner{}, err
		}
		return Banner{}, fmt.Errorf("banner: get: %w", err)
	}
	b.computeStatus(time.Now())
	return b, nil
}

// Save menimpa seluruh kolom banner yang bisa diubah.
func (r *Repository) Save(ctx context.Context, b Banner) (Banner, error) {
	out, err := scanBanner(r.db.QueryRow(ctx,
		`UPDATE banners
		    SET placement_id = (SELECT id FROM banner_placements WHERE code = $2), title = $3, subtitle = $4, cta_label = $5, desktop_image_url = $6,
		        mobile_image_url = $7, link_url = $8, sort_order = $9, is_active = $10,
		        starts_at = $11, ends_at = $12
		  WHERE id = $1
		  RETURNING `+bannerColumns,
		b.ID, b.Placement, b.Title, b.Subtitle, b.CTALabel, b.ImageURL, b.MobileImageURL,
		b.LinkURL, b.SortOrder, b.IsActive, b.StartsAt, b.EndsAt))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Banner{}, err
		}
		return Banner{}, fmt.Errorf("banner: save: %w", err)
	}
	out.computeStatus(time.Now())
	return out, nil
}

// Delete menghapus banner.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM banners WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("banner: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanBanner(row pgx.Row) (Banner, error) {
	var b Banner
	err := row.Scan(&b.ID, &b.Placement, &b.Title, &b.Subtitle, &b.CTALabel, &b.ImageURL,
		&b.MobileImageURL, &b.LinkURL, &b.SortOrder, &b.IsActive, &b.StartsAt, &b.EndsAt, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Banner{}, ErrNotFound
	}
	return b, err
}
