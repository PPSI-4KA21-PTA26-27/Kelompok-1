package post

import (
	"context"
	"errors"
	"fmt"
	"hii-stranger1/backend/internal/pkg/database"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	listColumns = `p.id, p.slug, p.title, p.excerpt, p.cover_image_url, p.tag, p.status, p.author_id,
		u.name, p.published_at, p.created_at, p.updated_at, length(p.content)`
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
	wordsPerMinute      = 200
	// charsPerWord adalah perkiraan kasar panjang kata bahasa Indonesia.
	charsPerWord = 6
)

// Repository mengakses tabel posts dan post_products.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (f ListFilter) where() (string, []any) {
	var conds []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, strings.ReplaceAll(cond, "$?", fmt.Sprintf("$%d", len(args))))
	}
	if f.Public {
		conds = append(conds, "p.status = 'published'", "p.published_at <= now()")
	} else if f.Status != "" {
		add("p.status = $?", f.Status)
	}
	if f.Tag != "" {
		add("lower(p.tag) = lower($?)", f.Tag)
	}
	if f.Q != "" {
		add(`p.title ILIKE $? ESCAPE '\'`, "%"+database.EscapeLike(f.Q)+"%")
	}
	if len(conds) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

// List mengembalikan artikel tanpa isi lengkap.
func (r *Repository) List(ctx context.Context, f ListFilter, limit, offset int) ([]Post, int, error) {
	where, args := f.where()
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM posts p"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("post: count: %w", err)
	}
	order := "p.updated_at DESC, p.id"
	if f.Public {
		order = "p.published_at DESC, p.id"
	}
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx,
		fmt.Sprintf("SELECT %s FROM posts p LEFT JOIN users u ON u.id = p.author_id%s ORDER BY %s LIMIT $%d OFFSET $%d",
			listColumns, where, order, len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("post: list: %w", err)
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Post, error) { return scanListPost(row) })
	if err != nil {
		return nil, 0, fmt.Errorf("post: list scan: %w", err)
	}
	if out == nil {
		out = []Post{}
	}
	return out, total, nil
}

// Get mengambil artikel berdasarkan id (admin) atau slug terbit (publik).
func (r *Repository) Get(ctx context.Context, key string, bySlug, public bool) (Detail, error) {
	cond := "p.id = $1"
	if bySlug {
		cond = "p.slug = $1"
	}
	if public {
		cond += " AND p.status = 'published' AND p.published_at <= now()"
	}
	var d Detail
	var contentLen int
	err := r.db.QueryRow(ctx,
		`SELECT p.id, p.slug, p.title, p.excerpt, p.content, p.cover_image_url, p.tag, p.status, p.author_id,
		        u.name, p.published_at, p.created_at, p.updated_at, length(p.content)
		   FROM posts p LEFT JOIN users u ON u.id = p.author_id
		  WHERE `+cond, key).Scan(&d.ID, &d.Slug, &d.Title, &d.Excerpt, &d.Content, &d.CoverImageURL, &d.Tag,
		&d.Status, &d.AuthorID, &d.AuthorName, &d.PublishedAt, &d.CreatedAt, &d.UpdatedAt, &contentLen)
	if errors.Is(err, pgx.ErrNoRows) {
		return Detail{}, ErrNotFound
	}
	if err != nil {
		return Detail{}, fmt.Errorf("post: get: %w", err)
	}
	d.ReadMinutes = readMinutes(contentLen)

	productFilter := ""
	if public {
		productFilter = " AND pr.status = 'active'"
	}
	rows, err := r.db.Query(ctx,
		`SELECT pr.id, pr.name, pr.price::text, pr.discount_percent::float8,
		        (SELECT pi.image_url FROM product_images pi WHERE pi.product_id = pr.id
		          ORDER BY pi.is_primary DESC, pi.sort_order, pi.id LIMIT 1),
		        (SELECT COALESCE(SUM(v.stock), 0)::int FROM product_variants v WHERE v.product_id = pr.id),
		        pr.status
		   FROM post_products pp JOIN products pr ON pr.id = pp.product_id
		  WHERE pp.post_id = $1`+productFilter+`
		  ORDER BY pp.sort_order, pr.name`, d.ID)
	if err != nil {
		return Detail{}, fmt.Errorf("post: products: %w", err)
	}
	d.Products, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (Product, error) {
		var p Product
		err := row.Scan(&p.ID, &p.Name, &p.Price, &p.DiscountPercent, &p.PrimaryImageURL, &p.TotalStock, &p.Status)
		return p, err
	})
	if err != nil {
		return Detail{}, fmt.Errorf("post: products scan: %w", err)
	}
	d.ProductIDs = make([]string, 0, len(d.Products))
	for _, p := range d.Products {
		d.ProductIDs = append(d.ProductIDs, p.ID)
	}
	if d.Products == nil {
		d.Products = []Product{}
	}
	return d, nil
}

// Save membuat (id kosong) atau memperbarui artikel beserta produk terkait.
func (r *Repository) Save(ctx context.Context, id, authorID string, in Input) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("post: save begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Artikel yang diterbitkan tanpa tanggal terbit memakai waktu sekarang.
	const publishedAt = `CASE WHEN $7 = 'published' THEN COALESCE($8, now()) ELSE $8 END`
	if id == "" {
		err = tx.QueryRow(ctx,
			`INSERT INTO posts (slug, title, excerpt, content, cover_image_url, tag, status, published_at, author_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, `+publishedAt+`, $9)
			 RETURNING id`,
			in.Slug, in.Title, in.Excerpt, in.Content, in.CoverImageURL, in.Tag, in.Status, in.PublishedAt, authorID).Scan(&id)
	} else {
		err = tx.QueryRow(ctx,
			`UPDATE posts
			    SET slug = $1, title = $2, excerpt = $3, content = $4, cover_image_url = $5, tag = $6,
			        status = $7, published_at = `+publishedAt+`, updated_at = now()
			  WHERE id = $9
			  RETURNING id`,
			in.Slug, in.Title, in.Excerpt, in.Content, in.CoverImageURL, in.Tag, in.Status, in.PublishedAt, id).Scan(&id)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", wrap("save", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM post_products WHERE post_id = $1`, id); err != nil {
		return "", wrap("save clear products", err)
	}
	for i, pid := range in.ProductIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO post_products (post_id, product_id, sort_order) VALUES ($1, $2, $3)`, id, pid, i); err != nil {
			return "", wrap("save product", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return "", wrap("save commit", err)
	}
	return id, nil
}

// Delete menghapus artikel.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("post: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanListPost(row pgx.Row) (Post, error) {
	var p Post
	var contentLen int
	err := row.Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.CoverImageURL, &p.Tag, &p.Status, &p.AuthorID,
		&p.AuthorName, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &contentLen)
	p.ReadMinutes = readMinutes(contentLen)
	return p, err
}

func readMinutes(chars int) int {
	m := chars / charsPerWord / wordsPerMinute
	if m < 1 {
		return 1
	}
	return m
}

func wrap(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == uniqueViolation && pgErr.ConstraintName == "uq_posts_slug":
			return ErrSlugTaken
		case pgErr.Code == foreignKeyViolation && pgErr.ConstraintName == "fk_post_products_product":
			return ErrProductNotFound
		}
	}
	return fmt.Errorf("post: %s: %w", op, err)
}
