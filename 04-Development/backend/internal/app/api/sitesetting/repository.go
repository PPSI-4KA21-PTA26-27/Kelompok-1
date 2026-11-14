package sitesetting

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tabel site_settings hanya berisi satu baris (dijaga indeks unik
// uq_site_settings_singleton); id-nya UUID sesuai SDD sehingga baris dibaca
// tanpa menyebut id. maintenance_message boleh NULL di basis data.
const (
	settingsColumns = `maintenance_mode, COALESCE(maintenance_message, ''), qris_image_url, qris_merchant_name, qris_nmid, payment_window_hours`
	defaultMessage  = "Toko sedang dalam perawatan. Silakan kembali beberapa saat lagi."
	defaultWindow   = 24
)

// Repository mengakses tabel site_settings.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Get mengembalikan pengaturan; nilai default bila baris belum ada.
func (r *Repository) Get(ctx context.Context) (Settings, error) {
	s, err := scan(r.db.QueryRow(ctx, `SELECT `+settingsColumns+` FROM site_settings LIMIT 1`))
	if errors.Is(err, pgx.ErrNoRows) {
		return Settings{MaintenanceMessage: defaultMessage, PaymentWindowHours: defaultWindow}, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("sitesetting: get: %w", err)
	}
	return s, nil
}

// Save menyimpan seluruh pengaturan: mengubah baris yang ada, atau membuatnya
// bila belum ada. Pesan maintenance kosong disimpan sebagai NULL.
func (r *Repository) Save(ctx context.Context, s Settings) (Settings, error) {
	out, err := scan(r.db.QueryRow(ctx,
		`WITH upd AS (
		     UPDATE site_settings
		        SET maintenance_mode = $1, maintenance_message = NULLIF($2::text, ''),
		            qris_image_url = $3, qris_merchant_name = $4, qris_nmid = $5,
		            payment_window_hours = $6
		  RETURNING `+settingsColumns+`
		 ), ins AS (
		     INSERT INTO site_settings (maintenance_mode, maintenance_message, qris_image_url,
		                                qris_merchant_name, qris_nmid, payment_window_hours)
		     SELECT $1::boolean, NULLIF($2::text, ''), $3::text, $4::text, $5::text, $6::int
		      WHERE NOT EXISTS (SELECT 1 FROM upd)
		  RETURNING `+settingsColumns+`
		 )
		 SELECT * FROM upd UNION ALL SELECT * FROM ins`,
		s.MaintenanceMode, s.MaintenanceMessage, s.QRISImageURL,
		s.QRISMerchantName, s.QRISNMID, s.PaymentWindowHours))
	if err != nil {
		return Settings{}, fmt.Errorf("sitesetting: save: %w", err)
	}
	return out, nil
}

func scan(row pgx.Row) (Settings, error) {
	var s Settings
	err := row.Scan(&s.MaintenanceMode, &s.MaintenanceMessage, &s.QRISImageURL,
		&s.QRISMerchantName, &s.QRISNMID, &s.PaymentWindowHours)
	return s, err
}
