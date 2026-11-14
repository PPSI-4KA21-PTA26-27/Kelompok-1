package category

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	categoryColumns = `id, name, is_active`
	uniqueViolation = "23505"
	nameConstraint  = "uq_categories_name"
)

// Repository mengakses tabel categories.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository di atas pool koneksi.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// List mengambil satu halaman kategori beserta total seluruh data yang cocok.
// Tanpa includeInactive, hanya kategori aktif yang diambil. Diurutkan menurut
// nama lalu id supaya pagination stabil.
func (r *Repository) List(ctx context.Context, includeInactive bool, limit, offset int) ([]Category, int, error) {
	where := ` WHERE is_active`
	if includeInactive {
		where = ""
	}

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM categories`+where).Scan(&total); err != nil {
		return nil, 0, wrap("count", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+categoryColumns+` FROM categories`+where+` ORDER BY name, id LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, 0, wrap("list", err)
	}
	defer rows.Close()

	items := make([]Category, 0, limit)
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, 0, wrap("list scan", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, wrap("list rows", err)
	}
	return items, total, nil
}

// GetByID mencari kategori berdasarkan ID, tanpa melihat status aktifnya.
func (r *Repository) GetByID(ctx context.Context, id string) (Category, error) {
	c, err := scanCategory(r.db.QueryRow(ctx,
		`SELECT `+categoryColumns+` FROM categories WHERE id = $1`, id))
	if err != nil {
		return Category{}, wrap("get by id", err)
	}
	return c, nil
}

// Create menyimpan kategori baru. Mengembalikan ErrNameTaken bila nama sudah ada.
func (r *Repository) Create(ctx context.Context, name string, isActive bool) (Category, error) {
	c, err := scanCategory(r.db.QueryRow(ctx,
		`INSERT INTO categories (name, is_active) VALUES ($1, $2) RETURNING `+categoryColumns,
		name, isActive))
	if err != nil {
		if isNameTaken(err) {
			return Category{}, ErrNameTaken
		}
		return Category{}, wrap("create", err)
	}
	return c, nil
}

// Update mengubah nama dan/atau status aktif dalam satu query. Mengembalikan
// ErrNotFound bila kategori tidak ada dan ErrNameTaken bila nama bentrok.
func (r *Repository) Update(ctx context.Context, id string, ch Changes) (Category, error) {
	c, err := scanCategory(r.db.QueryRow(ctx,
		`UPDATE categories
		 SET name = COALESCE($2, name),
		     is_active = COALESCE($3, is_active)
		 WHERE id = $1
		 RETURNING `+categoryColumns,
		id, ch.Name, ch.IsActive))
	if err != nil {
		if isNameTaken(err) {
			return Category{}, ErrNameTaken
		}
		return Category{}, wrap("update", err)
	}
	return c, nil
}

// Delete menghapus kategori. Produk yang memakainya tidak ikut terhapus:
// category_id produk menjadi NULL (ON DELETE SET NULL). Mengembalikan
// ErrNotFound bila kategori tidak ada.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return wrap("delete", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanCategory(row pgx.Row) (Category, error) {
	var c Category
	err := row.Scan(&c.ID, &c.Name, &c.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrNotFound
	}
	return c, err
}

func isNameTaken(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == nameConstraint
}

// wrap menambahkan konteks operasi, tetapi membiarkan error domain tetap utuh
// agar pemanggil bisa memeriksanya dengan errors.Is.
func wrap(op string, err error) error {
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrNameTaken) {
		return err
	}
	return fmt.Errorf("category: %s: %w", op, err)
}
