package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// price dan discount_percent diubah tipenya supaya bisa dibaca ke string/float64.
const (
	productColumns = `id, category_id, name, description, price::text, discount_percent::float8,
		weight_gram, availability_type, po_estimate_days, status`

	foreignKeyViolation = "23503"
	// restrictViolation muncul saat baris yang mau dihapus masih dipakai
	// tabel lain dengan ON DELETE RESTRICT (mis. order_items).
	restrictViolation = "23001"

	// primaryImage mengambil URL gambar utama (atau gambar pertama) produk p.
	// Memakai index idx_product_images_primary, jadi cukup satu lompatan index.
	primaryImage = `(SELECT pi.image_url FROM product_images pi
		WHERE pi.product_id = p.id
		ORDER BY pi.is_primary DESC, pi.sort_order, pi.id LIMIT 1)`

	// rawColumns adalah kolom produk tanpa cast (dipakai di subquery halaman),
	// pageColumns adalah kolom yang sama setelah dibungkus subquery p dan
	// di-cast seperti productColumns.
	rawColumns = `id, category_id, name, description, price, discount_percent,
		weight_gram, availability_type, po_estimate_days, status, created_at`
	pageColumns = `p.id, p.category_id, p.name, p.description, p.price::text, p.discount_percent::float8,
		p.weight_gram, p.availability_type, p.po_estimate_days, p.status`
)

// orderBy hanya berisi nilai tetap; input client tidak pernah disambung ke SQL.
var orderBy = map[string]string{
	"":       "name, id",
	"name":   "name, id",
	"-name":  "name DESC, id",
	"price":  "price, id",
	"-price": "price DESC, id",
	// newest: produk terbaru dulu (beranda "Baru Datang").
	"newest": "created_at DESC, id",
}

// stockColumns menghitung total stok & jumlah varian per produk di daftar.
const stockColumns = `(SELECT COALESCE(SUM(v.stock), 0)::int FROM product_variants v WHERE v.product_id = p.id),
		(SELECT COUNT(*)::int FROM product_variants v WHERE v.product_id = p.id)`

// Repository mengakses tabel products.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository di atas pool koneksi.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// List mengambil satu halaman produk yang cocok dengan f beserta totalnya.
// Sort harus salah satu kunci orderBy (sudah divalidasi handler).
func (r *Repository) List(ctx context.Context, f Filter, limit, offset int) ([]Product, int, error) {
	var conds []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if f.Query != "" {
		add("name ILIKE $%d", "%"+escapeLike(f.Query)+"%")
	}
	if f.CategoryID != "" {
		add("category_id = $%d", f.CategoryID)
	}
	if f.AvailabilityType != "" {
		add("availability_type = $%d", f.AvailabilityType)
	}
	if f.Status != "" {
		add("status = $%d", f.Status)
	}
	if f.MinPrice != "" {
		add("price >= $%d::numeric", f.MinPrice)
	}
	if f.MaxPrice != "" {
		add("price <= $%d::numeric", f.MaxPrice)
	}
	if f.NewDays > 0 {
		add("created_at >= now() - make_interval(days => $%d::int)", f.NewDays)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	order, ok := orderBy[f.Sort]
	if !ok {
		order = orderBy[""]
	}

	// Hitung total dan ambil satu halaman dalam SATU perjalanan ke database
	// (pgx.Batch). Halaman diambil dulu di subquery p (kolom asli, tanpa
	// cast, supaya "ORDER BY price" mengurutkan angka), baru gambar utama
	// dicari untuk baris di halaman itu saja. Dulu gambar dicari untuk semua
	// baris sebelum OFFSET, sehingga halaman belakang makin lambat.
	pageArgs := append(append([]any{}, args...), limit, offset)
	query := fmt.Sprintf(`SELECT %s, %s, %s
		  FROM (SELECT %s FROM products%s ORDER BY %s LIMIT $%d OFFSET $%d) p
		 ORDER BY %s`,
		pageColumns, primaryImage, stockColumns, rawColumns, where, order, len(pageArgs)-1, len(pageArgs), qualify(order, "p."))

	b := &pgx.Batch{}
	b.Queue("SELECT COUNT(*) FROM products"+where, args...)
	b.Queue(query, pageArgs...)
	br := r.db.SendBatch(ctx, b)
	defer func() { _ = br.Close() }()

	var total int
	if err := br.QueryRow().Scan(&total); err != nil {
		return nil, 0, wrap("count", err)
	}
	items, err := collect(br, scanListProduct)
	if err != nil {
		return nil, 0, wrap("list", err)
	}
	return items, total, nil
}

// qualify menambahkan awalan tabel ke setiap kolom pada klausa ORDER BY,
// contoh "price DESC, id" menjadi "p.price DESC, p.id".
func qualify(order, prefix string) string {
	parts := strings.Split(order, ", ")
	for i, part := range parts {
		parts[i] = prefix + part
	}
	return strings.Join(parts, ", ")
}

// GetDetail mengambil produk beserta varian dan gambarnya dalam SATU
// perjalanan ke database (pgx.Batch), bukan tiga query berurutan.
func (r *Repository) GetDetail(ctx context.Context, id string) (Product, []Variant, []Image, error) {
	b := &pgx.Batch{}
	b.Queue("SELECT "+productColumns+" FROM products WHERE id = $1", id)
	b.Queue("SELECT "+variantColumns+" FROM product_variants WHERE product_id = $1 ORDER BY sku, id", id)
	b.Queue("SELECT "+imageColumns+" FROM product_images WHERE product_id = $1 ORDER BY sort_order, id", id)
	br := r.db.SendBatch(ctx, b)
	defer func() { _ = br.Close() }()

	p, err := scanProduct(br.QueryRow())
	if err != nil {
		return Product{}, nil, nil, wrap("get detail", err)
	}
	variants, err := collect(br, scanVariant)
	if err != nil {
		return Product{}, nil, nil, fmt.Errorf("product: get detail variants: %w", err)
	}
	images, err := collect(br, scanImage)
	if err != nil {
		return Product{}, nil, nil, fmt.Errorf("product: get detail images: %w", err)
	}
	return p, variants, images, nil
}

// collect membaca hasil query berikutnya dari batch menjadi slice (tidak
// pernah nil, supaya JSON-nya [] bukan null).
func collect[T any](br pgx.BatchResults, scan func(pgx.Row) (T, error)) ([]T, error) {
	rows, err := br.Query()
	if err != nil {
		return nil, err
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (T, error) { return scan(row) })
	if out == nil {
		out = []T{}
	}
	return out, err
}

// GetByID mencari produk berdasarkan ID, apa pun statusnya.
func (r *Repository) GetByID(ctx context.Context, id string) (Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx,
		"SELECT "+productColumns+" FROM products WHERE id = $1", id))
	if err != nil {
		return Product{}, wrap("get by id", err)
	}
	return p, nil
}

// Create menyimpan produk baru. Mengembalikan ErrCategoryNotFound bila
// category_id tidak ada.
func (r *Repository) Create(ctx context.Context, p Product) (Product, error) {
	created, err := scanProduct(r.db.QueryRow(ctx,
		`INSERT INTO products
		   (category_id, name, description, price, discount_percent, weight_gram,
		    availability_type, po_estimate_days, status)
		 VALUES ($1, $2, $3, $4::numeric, $5, $6, $7, $8, $9)
		 RETURNING `+productColumns,
		p.CategoryID, p.Name, p.Description, p.Price, p.DiscountPercent, p.WeightGram,
		p.AvailabilityType, p.PoEstimateDays, p.Status))
	if err != nil {
		return Product{}, wrap("create", err)
	}
	return created, nil
}

// Update hanya mengubah field yang tidak nil, dalam satu query. Mengembalikan
// ErrNotFound bila produk tidak ada dan ErrCategoryNotFound bila category_id salah.
func (r *Repository) Update(ctx context.Context, id string, ch Changes) (Product, error) {
	args := []any{id}
	var sets []string
	set := func(col string, v any, cast string) {
		args = append(args, v)
		sets = append(sets, fmt.Sprintf("%s = $%d%s", col, len(args), cast))
	}
	if ch.CategoryID != nil {
		set("category_id", *ch.CategoryID, "")
	}
	if ch.Name != nil {
		set("name", *ch.Name, "")
	}
	if ch.Description != nil {
		set("description", *ch.Description, "")
	}
	if ch.Price != nil {
		set("price", *ch.Price, "::numeric")
	}
	if ch.DiscountPercent != nil {
		set("discount_percent", *ch.DiscountPercent, "")
	}
	if ch.WeightGram != nil {
		set("weight_gram", *ch.WeightGram, "")
	}
	if ch.AvailabilityType != nil {
		set("availability_type", *ch.AvailabilityType, "")
	}
	if ch.PoEstimateDays != nil {
		set("po_estimate_days", *ch.PoEstimateDays, "")
	}
	if ch.Status != nil {
		set("status", *ch.Status, "")
	}
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	p, err := scanProduct(r.db.QueryRow(ctx,
		"UPDATE products SET "+strings.Join(sets, ", ")+" WHERE id = $1 RETURNING "+productColumns,
		args...))
	if err != nil {
		return Product{}, wrap("update", err)
	}
	return p, nil
}

// Delete menghapus produk. Mengembalikan ErrNotFound bila tidak ada dan
// ErrInUse bila masih dipakai pesanan.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return ErrInUse
		}
		return wrap("delete", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func productDest(p *Product) []any {
	return []any{&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price,
		&p.DiscountPercent, &p.WeightGram, &p.AvailabilityType, &p.PoEstimateDays, &p.Status}
}

func scanProduct(row pgx.Row) (Product, error) {
	var p Product
	err := row.Scan(productDest(&p)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return p, err
}

// scanListProduct sama dengan scanProduct ditambah kolom gambar utama.
func scanListProduct(row pgx.Row) (Product, error) {
	var p Product
	err := row.Scan(append(productDest(&p), &p.PrimaryImageURL, &p.TotalStock, &p.VariantCount)...)
	return p, err
}

// isForeignKeyViolation melaporkan pelanggaran relasi: foreign key tidak
// ditemukan, atau baris masih direferensikan (RESTRICT) saat dihapus.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == foreignKeyViolation || pgErr.Code == restrictViolation)
}

// escapeLike membuat %, _, dan \ dibaca apa adanya oleh ILIKE.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(s)
}

// wrap menambahkan konteks operasi, tetapi membiarkan error domain tetap utuh
// agar pemanggil bisa memeriksanya dengan errors.Is. Pelanggaran foreign key
// pada INSERT/UPDATE berarti category_id tidak ada.
func wrap(op string, err error) error {
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrInUse) {
		return err
	}
	if isForeignKeyViolation(err) && (op == "create" || op == "update") {
		return ErrCategoryNotFound
	}
	return fmt.Errorf("product: %s: %w", op, err)
}
