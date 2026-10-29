package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"hii-stranger1/backend/models"
)

// Alur: services.CatalogService -> ProductRepository -> db.DB (PostgreSQL).
// Paket ini hanya berisi query SQL. Aturan bisnis (diskon, label, status stok) ada di services.
//
// Kolom tabel yang dibaca (dibuat lewat migrasi, bukan di sini):
//   products(id, category_id, name, slug, description, price, original_price,
//            is_preorder, preorder_days, image_url, is_active, created_at)
//   product_variants(id, product_id, color, size, sku, stock)
//   categories(id, name, slug, is_active)
// Produk hanya tampil ke publik kalau produk DAN kategorinya sama-sama aktif.

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// productFrom adalah bagian FROM yang sama untuk query daftar, hitung, dan detail.
// Subquery menjumlahkan stok semua varian per produk; produk tanpa varian dianggap stok 0.
const productFrom = `
	FROM products p
	JOIN categories c ON c.id = p.category_id
	LEFT JOIN (
		SELECT product_id, SUM(stock) AS total_stock
		FROM product_variants
		GROUP BY product_id
	) s ON s.product_id = p.id`

// productColumns adalah kolom untuk membentuk models.ProductSummary.
// Urutannya harus sama dengan urutan Scan di scanProduct.
const productColumns = `p.id, p.name, p.slug, c.name, c.slug, p.price, p.original_price,
	p.is_preorder, p.preorder_days, COALESCE(p.image_url, ''), COALESCE(s.total_stock, 0)`

// productOrderBy adalah pilihan urutan yang diizinkan. Teks ORDER BY tidak boleh berasal dari
// input pengguna, jadi hanya nilai di peta ini yang dipakai. p.id memastikan urutan stabil
// saat dua produk punya nilai yang sama.
var productOrderBy = map[models.ProductSort]string{
	models.SortNewest:    "p.created_at DESC, p.id DESC",
	models.SortPriceAsc:  "p.price ASC, p.id DESC",
	models.SortPriceDesc: "p.price DESC, p.id DESC",
	models.SortNameAsc:   "p.name ASC, p.id DESC",
}

// rowScanner dipenuhi oleh *sql.Row dan *sql.Rows, jadi satu fungsi scan bisa dipakai keduanya.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanProduct membaca kolom productColumns, lalu kolom tambahan (extra) kalau ada.
// Kolom yang boleh NULL (harga coret, estimasi hari) dibaca lewat sql.NullInt64 lalu diubah
// menjadi pointer, supaya "tidak ada" tidak tercampur dengan angka 0.
func scanProduct(row rowScanner, extra ...any) (models.ProductSummary, error) {
	var (
		p        models.ProductSummary
		original sql.NullInt64
		days     sql.NullInt64
	)
	dest := append([]any{
		&p.ID, &p.Name, &p.Slug, &p.Category.Name, &p.Category.Slug, &p.Price, &original,
		&p.IsPreorder, &days, &p.ImageURL, &p.TotalStock,
	}, extra...)
	if err := row.Scan(dest...); err != nil {
		return models.ProductSummary{}, err
	}

	if original.Valid {
		p.OriginalPrice = &original.Int64
	}
	if days.Valid {
		d := int(days.Int64)
		p.PreorderDays = &d
	}
	return p, nil
}

// buildProductWhere menyusun klausa WHERE dari filter katalog. Teks SQL-nya tetap (tidak ada
// input pengguna di dalamnya); nilai filter dikirim terpisah lewat placeholder $1, $2, dst.
func buildProductWhere(f models.ProductFilter) (string, []any) {
	conds := []string{"p.is_active", "c.is_active"}
	var args []any

	if f.CategorySlug != "" {
		args = append(args, f.CategorySlug)
		conds = append(conds, fmt.Sprintf("c.slug = $%d", len(args)))
	}
	if f.Search != "" {
		args = append(args, escapeLike(f.Search))
		conds = append(conds, fmt.Sprintf("p.name ILIKE '%%' || $%d || '%%'", len(args)))
	}

	switch f.Availability {
	case models.AvailabilityInStock:
		conds = append(conds, "COALESCE(s.total_stock, 0) > 0")
	case models.AvailabilityPreorder:
		conds = append(conds, "p.is_preorder")
	}

	return " WHERE " + strings.Join(conds, " AND "), args
}

// escapeLike membuat karakter khusus LIKE (\, %, _) dibaca apa adanya, supaya pencarian "50%"
// tidak dianggap pola "apa saja".
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// List mengambil satu halaman produk aktif sesuai filter, beserta jumlah total yang cocok
// (untuk paginasi). Hasil selalu berupa slice, kosong kalau tidak ada yang cocok.
func (r *ProductRepository) List(ctx context.Context, f models.ProductFilter) ([]models.ProductSummary, int, error) {
	orderBy, ok := productOrderBy[f.Sort]
	if !ok {
		return nil, 0, fmt.Errorf("urutan produk tidak dikenal: %q", f.Sort)
	}
	where, args := buildProductWhere(f)

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) `+productFrom+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hitung produk: %w", err)
	}
	products := []models.ProductSummary{}
	if total == 0 {
		return products, 0, nil
	}

	// LIMIT dan OFFSET juga lewat placeholder; nomornya melanjutkan argumen filter.
	query := fmt.Sprintf(`SELECT %s %s %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		productColumns, productFrom, where, orderBy, len(args)+1, len(args)+2)
	args = append(args, f.Limit, (f.Page-1)*f.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("ambil produk: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("baca produk: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("baca daftar produk: %w", err)
	}
	return products, total, nil
}

// FindBySlug mengambil satu produk aktif beserta varian-nya. Mengembalikan ErrNotFound kalau
// slug tidak ada, produk nonaktif, atau kategorinya nonaktif.
func (r *ProductRepository) FindBySlug(ctx context.Context, slug string) (*models.ProductDetail, error) {
	var description string
	row := r.db.QueryRowContext(ctx,
		`SELECT `+productColumns+`, COALESCE(p.description, '') `+productFrom+`
		 WHERE p.slug = $1 AND p.is_active AND c.is_active`, slug)

	summary, err := scanProduct(row, &description)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("cari produk by slug: %w", err)
	}

	variants, err := r.listVariants(ctx, summary.ID)
	if err != nil {
		return nil, err
	}
	return &models.ProductDetail{ProductSummary: summary, Description: description, Variants: variants}, nil
}

// listVariants mengambil varian satu produk menurut urutan input di admin (id), karena urutan
// ukuran seperti S, M, L, XL tidak bisa diurutkan dengan benar secara alfabet.
func (r *ProductRepository) listVariants(ctx context.Context, productID int64) ([]models.ProductVariant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, COALESCE(color, ''), COALESCE(size, ''), sku, stock
		 FROM product_variants
		 WHERE product_id = $1
		 ORDER BY id`, productID)
	if err != nil {
		return nil, fmt.Errorf("ambil varian produk: %w", err)
	}
	defer rows.Close()

	variants := []models.ProductVariant{}
	for rows.Next() {
		var v models.ProductVariant
		if err := rows.Scan(&v.ID, &v.Color, &v.Size, &v.SKU, &v.Stock); err != nil {
			return nil, fmt.Errorf("baca varian produk: %w", err)
		}
		variants = append(variants, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar varian: %w", err)
	}
	return variants, nil
}
