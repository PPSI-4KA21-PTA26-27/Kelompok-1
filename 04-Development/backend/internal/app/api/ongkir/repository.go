package ongkir

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository membaca data keranjang yang dibutuhkan untuk hitung berat.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository di atas pool koneksi.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CartWeight menjumlahkan berat (gram) item keranjang user: berat produk x
// jumlah. ids kosong berarti seluruh keranjang. count adalah jumlah item
// keranjang yang ikut dihitung.
func (r *Repository) CartWeight(ctx context.Context, userID string, ids []string) (weight, count int, err error) {
	query := `SELECT COALESCE(SUM(p.weight_gram * ci.quantity), 0), COUNT(*)
	            FROM cart_items ci
	            JOIN product_variants v ON v.id = ci.variant_id
	            JOIN products p ON p.id = v.product_id
	           WHERE ci.user_id = $1`
	args := []any{userID}
	if len(ids) > 0 {
		query += ` AND ci.id = ANY($2::uuid[])`
		args = append(args, ids)
	}
	if err := r.db.QueryRow(ctx, query, args...).Scan(&weight, &count); err != nil {
		return 0, 0, fmt.Errorf("ongkir: cart weight: %w", err)
	}
	return weight, count, nil
}
