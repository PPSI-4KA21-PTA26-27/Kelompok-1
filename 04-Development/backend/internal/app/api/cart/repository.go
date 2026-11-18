package cart

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	foreignKeyViolation = "23503"
	checkViolation      = "23514"

	// unitPrice adalah harga satuan setelah diskon produk.
	unitPrice = `ROUND(p.price * (100 - p.discount_percent) / 100, 2)`

	// viewSelect membaca item keranjang beserta data varian dan produknya.
	viewSelect = `
		SELECT ci.id, v.id, p.id, p.name, v.sku, v.size, v.color,
		       ` + unitPrice + `::text,
		       ci.quantity, v.stock,
		       (` + unitPrice + ` * ci.quantity)::text,
		       (SELECT pi.image_url
		          FROM product_images pi
		         WHERE pi.product_id = p.id
		         ORDER BY pi.is_primary DESC, pi.sort_order, pi.id
		         LIMIT 1)
		  FROM cart_items ci
		  JOIN product_variants v ON v.id = ci.variant_id
		  JOIN products p ON p.id = v.product_id`
)

// Repository mengakses tabel cart_items.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository di atas pool koneksi.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetCart mengambil seluruh isi keranjang milik user beserta subtotalnya.
func (r *Repository) GetCart(ctx context.Context, userID string) (Cart, error) {
	rows, err := r.db.Query(ctx,
		viewSelect+` WHERE ci.user_id = $1 ORDER BY p.name, v.sku, ci.id`,
		userID,
	)
	if err != nil {
		return Cart{}, wrap("get cart", err)
	}
	defer rows.Close()

	items := make([]ItemView, 0)
	total := new(big.Rat)

	for rows.Next() {
		it, err := scanView(rows)
		if err != nil {
			return Cart{}, wrap("get cart scan", err)
		}

		sub, ok := new(big.Rat).SetString(it.Subtotal)
		if !ok {
			return Cart{}, fmt.Errorf("cart: get cart: subtotal %q tidak valid", it.Subtotal)
		}
		total.Add(total, sub)

		items = append(items, it)
	}

	if err := rows.Err(); err != nil {
		return Cart{}, wrap("get cart rows", err)
	}

	return Cart{Items: items, Subtotal: total.FloatString(2)}, nil
}

// GetItem mengambil satu item keranjang milik user.
// Mengembalikan ErrNotFound bila tidak ada atau bukan milik user.
func (r *Repository) GetItem(ctx context.Context, userID, id string) (ItemView, error) {
	it, err := scanView(r.db.QueryRow(ctx,
		viewSelect+` WHERE ci.id = $1 AND ci.user_id = $2`,
		id, userID,
	))
	if err != nil {
		return ItemView{}, wrap("get item", err)
	}

	return it, nil
}

// AddItem menambahkan varian ke keranjang. Bila varian sudah ada di keranjang
// user, jumlahnya ditambah (tabel tidak punya unique constraint, jadi
// penggabungan dilakukan di sini). Total jumlah tidak boleh melebihi stok.
// Mengembalikan ID item yang dibuat atau digabung.
func (r *Repository) AddItem(
	ctx context.Context,
	userID, variantID string,
	quantity int,
) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", wrap("add begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Kunci baris varian supaya dua request bersamaan tidak membuat item ganda
	// atau sama-sama lolos cek stok.
	var stock int
	var productStatus string
	err = tx.QueryRow(ctx,
		`SELECT v.stock, p.status
		   FROM product_variants v JOIN products p ON p.id = v.product_id
		  WHERE v.id = $1
		    FOR NO KEY UPDATE OF v`,
		variantID,
	).Scan(&stock, &productStatus)
	// Produk nonaktif diperlakukan seperti varian yang tidak ada.
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && productStatus != "active") {
		return "", ErrVariantNotFound
	}
	if err != nil {
		return "", wrap("add lock variant", err)
	}

	var (
		id       string
		existing int
	)
	err = tx.QueryRow(ctx,
		`SELECT id, quantity
		   FROM cart_items
		  WHERE user_id = $1 AND variant_id = $2
		  ORDER BY id
		  LIMIT 1
		    FOR UPDATE`,
		userID, variantID,
	).Scan(&id, &existing)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		if quantity > stock {
			return "", ErrInsufficientStock
		}

		err = tx.QueryRow(ctx,
			`INSERT INTO cart_items (user_id, variant_id, quantity)
			 VALUES ($1, $2, $3)
			 RETURNING id`,
			userID, variantID, quantity,
		).Scan(&id)
		if err != nil {
			return "", wrap("add insert", err)
		}

	case err != nil:
		return "", wrap("add find existing", err)

	default:
		// Ditulis sebagai pengurangan agar tidak overflow.
		if quantity > stock-existing {
			return "", ErrInsufficientStock
		}

		_, err = tx.Exec(ctx,
			`UPDATE cart_items SET quantity = $2 WHERE id = $1`,
			id, existing+quantity,
		)
		if err != nil {
			return "", wrap("add merge", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", wrap("add commit", err)
	}

	return id, nil
}

// SetQuantity mengubah jumlah item milik user menjadi quantity.
// Mengembalikan ErrNotFound bila item tidak ada atau bukan milik user, dan
// ErrInsufficientStock bila quantity melebihi stok varian.
func (r *Repository) SetQuantity(
	ctx context.Context,
	userID, id string,
	quantity int,
) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE cart_items ci
		    SET quantity = $3
		   FROM product_variants v
		  WHERE ci.id = $1
		    AND ci.user_id = $2
		    AND v.id = ci.variant_id
		    AND v.stock >= $3`,
		id, userID, quantity,
	)
	if err != nil {
		return wrap("set quantity", err)
	}

	if tag.RowsAffected() > 0 {
		return nil
	}

	// Tidak ada baris berubah: item tidak ada, atau stok tidak cukup.
	var exists bool
	err = r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM cart_items WHERE id = $1 AND user_id = $2)`,
		id, userID,
	).Scan(&exists)
	if err != nil {
		return wrap("set quantity check", err)
	}

	if !exists {
		return ErrNotFound
	}

	return ErrInsufficientStock
}

// DeleteItem menghapus satu item milik user.
// Mengembalikan ErrNotFound bila item tidak ditemukan.
func (r *Repository) DeleteItem(ctx context.Context, userID, id string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM cart_items WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return wrap("delete item", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ClearCart menghapus seluruh item keranjang milik user.
func (r *Repository) ClearCart(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM cart_items WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return wrap("clear cart", err)
	}

	return nil
}

// scanView membaca satu baris hasil viewSelect.
func scanView(row pgx.Row) (ItemView, error) {
	var it ItemView

	err := row.Scan(
		&it.ID,
		&it.VariantID,
		&it.ProductID,
		&it.ProductName,
		&it.SKU,
		&it.Size,
		&it.Color,
		&it.UnitPrice,
		&it.Quantity,
		&it.Stock,
		&it.Subtotal,
		&it.PrimaryImageURL,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ItemView{}, ErrNotFound
	}
	if err != nil {
		return ItemView{}, err
	}

	return it, nil
}

// wrap menambahkan konteks operasi dan menerjemahkan error database
// menjadi domain error.
func wrap(op string, err error) error {
	if errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrVariantNotFound) ||
		errors.Is(err, ErrInvalidQuantity) ||
		errors.Is(err, ErrInsufficientStock) {
		return err
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case foreignKeyViolation:
			switch pgErr.ConstraintName {
			case "fk_cart_items_user":
				return ErrUserNotFound
			case "fk_cart_items_variant":
				return ErrVariantNotFound
			}

		case checkViolation:
			if pgErr.ConstraintName == "ck_cart_items_quantity" {
				return ErrInvalidQuantity
			}
		}
	}

	return fmt.Errorf("cart: %s: %w", op, err)
}
