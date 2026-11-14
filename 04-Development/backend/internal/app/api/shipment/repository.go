package shipment

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	foreignKeyViolation = "23503"

	// courier_name tidak disimpan di shipments (SDD); diturunkan dari kurir
	// pesanan: kode kurir (huruf besar) + layanan, mis. "JNE REG".
	shipmentColumns = `id, order_id,
		(SELECT upper(o.courier_code) || ' ' || o.courier_service FROM orders o WHERE o.id = shipments.order_id),
		tracking_number, coverage_status, shipping_status, updated_at`
)

// Repository mengakses tabel shipments.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository di atas pool koneksi.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// OrderOwner mengembalikan user_id pemilik order. ErrOrderNotFound bila
// order tidak ada.
func (r *Repository) OrderOwner(ctx context.Context, orderID string) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx, `SELECT user_id FROM orders WHERE id = $1`, orderID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrOrderNotFound
	}
	if err != nil {
		return "", fmt.Errorf("shipment: order owner: %w", err)
	}
	return userID, nil
}

// GetByOrder mengambil pengiriman milik order. ErrNotFound bila belum ada.
// Dipakai juga oleh package order untuk detail order.
func (r *Repository) GetByOrder(ctx context.Context, orderID string) (Shipment, error) {
	s, err := scanShipment(r.db.QueryRow(ctx,
		"SELECT "+shipmentColumns+" FROM shipments WHERE order_id = $1", orderID))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Shipment{}, err
		}
		return Shipment{}, fmt.Errorf("shipment: get by order: %w", err)
	}
	return s, nil
}

// queryRower dipenuhi oleh *pgxpool.Pool maupun pgx.Tx, sehingga query yang
// sama bisa dijalankan di luar atau di dalam transaksi.
type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Upsert membuat pengiriman bila belum ada, atau memperbaruinya bila sudah
// ada (ON CONFLICT order_id). created bernilai true bila baris baru dibuat.
func (r *Repository) Upsert(ctx context.Context, s Shipment) (Shipment, bool, error) {
	return upsert(ctx, r.db, s)
}

// UpsertTx sama dengan Upsert tetapi berjalan di transaksi tx milik pemanggil
// (dipakai package order saat status berubah menjadi shipped).
func (r *Repository) UpsertTx(ctx context.Context, tx pgx.Tx, s Shipment) (Shipment, bool, error) {
	return upsert(ctx, tx, s)
}

// upsert menjalankan INSERT ... ON CONFLICT. Trik "xmax = 0" bernilai true
// hanya untuk baris hasil INSERT, sehingga bisa membedakan buat vs ubah.
func upsert(ctx context.Context, q queryRower, s Shipment) (Shipment, bool, error) {
	var out Shipment
	var created bool
	err := q.QueryRow(ctx,
		`INSERT INTO shipments (order_id, tracking_number, coverage_status, shipping_status)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (order_id) DO UPDATE
		    SET tracking_number = EXCLUDED.tracking_number,
		        coverage_status = EXCLUDED.coverage_status,
		        shipping_status = EXCLUDED.shipping_status,
		        updated_at = now()
		 RETURNING `+shipmentColumns+`, (xmax = 0)`,
		s.OrderID, s.TrackingNumber, s.CoverageStatus, s.ShippingStatus,
	).Scan(&out.ID, &out.OrderID, &out.CourierName, &out.TrackingNumber,
		&out.CoverageStatus, &out.ShippingStatus, &out.UpdatedAt, &created)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation {
			return Shipment{}, false, ErrOrderNotFound
		}
		return Shipment{}, false, fmt.Errorf("shipment: upsert: %w", err)
	}
	return out, created, nil
}

// contactOf mengambil email & nama pemilik order serta nomor ordernya.
func (r *Repository) contactOf(ctx context.Context, orderID string) (contact, error) {
	var c contact
	err := r.db.QueryRow(ctx,
		`SELECT u.email, u.name, o.order_number
		   FROM orders o JOIN users u ON u.id = o.user_id
		  WHERE o.id = $1`, orderID).Scan(&c.Email, &c.Name, &c.OrderNumber)
	if err != nil {
		return contact{}, fmt.Errorf("shipment: contact: %w", err)
	}
	return c, nil
}

func scanShipment(row pgx.Row) (Shipment, error) {
	var s Shipment
	err := row.Scan(&s.ID, &s.OrderID, &s.CourierName, &s.TrackingNumber, &s.CoverageStatus, &s.ShippingStatus, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Shipment{}, ErrNotFound
	}
	return s, err
}
