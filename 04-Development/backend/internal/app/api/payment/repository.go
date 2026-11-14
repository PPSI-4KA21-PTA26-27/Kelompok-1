package payment

import (
	"context"
	"errors"
	"fmt"
	"hii-stranger1/backend/internal/pkg/database"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// proofTypeSQL menurunkan tipe MIME bukti dari ekstensi proof_url (alamat
// Cloudinary selalu berakhiran format berkas).
const proofTypeSQL = `CASE
	WHEN p.proof_url IS NULL THEN NULL
	WHEN p.proof_url ILIKE '%.pdf' THEN 'application/pdf'
	WHEN p.proof_url ILIKE '%.png' THEN 'image/png'
	WHEN p.proof_url ILIKE '%.webp' THEN 'image/webp'
	ELSE 'image/jpeg' END`

// paymentColumns tidak memuat proof_url; alamat bukti hanya dibaca GetProof.
const paymentColumns = `p.id, p.order_id, p.method, p.amount::text, p.unique_code, ` + proofTypeSQL + `,
	p.proof_url IS NOT NULL, p.status, p.paid_at, p.reviewed_by, p.reviewed_at, p.created_at`

// Repository mengakses tabel payments (dan membaca/mengubah orders seperlunya).
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
		return "", fmt.Errorf("payment: order owner: %w", err)
	}
	return userID, nil
}

// GetByOrder mengambil pembayaran milik order. ErrNotFound bila belum ada.
// Dipakai juga oleh package order untuk detail order.
func (r *Repository) GetByOrder(ctx context.Context, orderID string) (Payment, error) {
	p, err := scanPayment(r.db.QueryRow(ctx,
		`SELECT `+paymentColumns+` FROM payments p WHERE p.order_id = $1 ORDER BY p.id LIMIT 1`,
		orderID))
	if err != nil {
		return Payment{}, wrap("get by order", err)
	}
	return p, nil
}

// Create membuat pembayaran untuk order milik userID dalam satu transaksi.
// amount = total order + uniqueCode. Baris order dikunci agar dua request
// bersamaan tidak membuat dua pembayaran.
func (r *Repository) Create(ctx context.Context, orderID, userID, method string, uniqueCode int) (Payment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Payment{}, wrap("create begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var owner, status, total string
	err = tx.QueryRow(ctx,
		`SELECT user_id, order_status, total_amount::text FROM orders WHERE id = $1 FOR UPDATE`,
		orderID).Scan(&owner, &status, &total)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && owner != userID) {
		return Payment{}, ErrOrderNotFound
	}
	if err != nil {
		return Payment{}, wrap("create lock order", err)
	}
	if status != orderStatusPendingPayment {
		return Payment{}, ErrOrderNotPayable
	}

	var exists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM payments WHERE order_id = $1)`, orderID).Scan(&exists); err != nil {
		return Payment{}, wrap("create check existing", err)
	}
	if exists {
		return Payment{}, ErrAlreadyExists
	}

	p, err := scanPayment(tx.QueryRow(ctx,
		`INSERT INTO payments AS p (order_id, method, amount, unique_code, status)
		 VALUES ($1, $2, $3::numeric + $4::int, $4::int, $5)
		 RETURNING `+paymentColumns,
		orderID, method, total, uniqueCode, StatusPending))
	if err != nil {
		return Payment{}, wrap("create insert", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Payment{}, wrap("create commit", err)
	}
	return p, nil
}

// SaveProof menyimpan (atau mengganti) bukti pembayaran milik order. Status
// kembali pending dan hasil verifikasi sebelumnya dihapus. Ditolak bila
// pembayaran sudah approved atau order sudah tidak menunggu pembayaran.
// CheckProofUpload memeriksa lebih dulu apakah bukti boleh diunggah (tanpa
// mengunci), supaya berkas tidak dikirim ke Cloudinary untuk pembayaran yang
// pasti ditolak. SaveProof tetap memeriksa ulang di dalam transaksi.
func (r *Repository) CheckProofUpload(ctx context.Context, orderID string) error {
	var status, orderStatus string
	err := r.db.QueryRow(ctx,
		`SELECT p.status, o.order_status
		   FROM payments p
		   JOIN orders o ON o.id = p.order_id
		  WHERE p.order_id = $1
		  ORDER BY p.id
		  LIMIT 1`,
		orderID).Scan(&status, &orderStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return wrap("check proof upload", err)
	}
	if status == StatusApproved {
		return ErrAlreadyApproved
	}
	if orderStatus != orderStatusPendingPayment {
		return ErrOrderNotPayable
	}
	return nil
}

// SaveProof menyimpan alamat bukti (proofURL di Cloudinary) milik order.
func (r *Repository) SaveProof(ctx context.Context, orderID, proofURL string) (Payment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Payment{}, wrap("save proof begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id, status, orderStatus string
	err = tx.QueryRow(ctx,
		`SELECT p.id, p.status, o.order_status
		   FROM payments p
		   JOIN orders o ON o.id = p.order_id
		  WHERE p.order_id = $1
		  ORDER BY p.id
		  LIMIT 1
		    FOR UPDATE OF p`,
		orderID).Scan(&id, &status, &orderStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	if err != nil {
		return Payment{}, wrap("save proof lock", err)
	}
	if status == StatusApproved {
		return Payment{}, ErrAlreadyApproved
	}
	if orderStatus != orderStatusPendingPayment {
		return Payment{}, ErrOrderNotPayable
	}

	p, err := scanPayment(tx.QueryRow(ctx,
		`UPDATE payments AS p
		    SET proof_url = $2, status = $3, paid_at = now(),
		        reviewed_by = NULL, reviewed_at = NULL
		  WHERE p.id = $1
		  RETURNING `+paymentColumns,
		id, proofURL, StatusPending))
	if err != nil {
		return Payment{}, wrap("save proof update", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Payment{}, wrap("save proof commit", err)
	}
	return p, nil
}

// GetProof mengambil alamat bukti pembayaran order di Cloudinary beserta
// tipenya. ErrNotFound bila belum ada.
func (r *Repository) GetProof(ctx context.Context, orderID string) (proofRef, error) {
	var ref proofRef
	err := r.db.QueryRow(ctx,
		`SELECT p.proof_url, `+proofTypeSQL+` FROM payments p
		  WHERE p.order_id = $1 AND p.proof_url IS NOT NULL
		  ORDER BY p.id LIMIT 1`,
		orderID).Scan(&ref.URL, &ref.Type)
	if errors.Is(err, pgx.ErrNoRows) {
		return proofRef{}, ErrNotFound
	}
	if err != nil {
		return proofRef{}, wrap("get proof", err)
	}
	return ref, nil
}

// List mengambil satu halaman pembayaran (antrean verifikasi admin). Yang
// menunggu verifikasi (pending dengan bukti) diletakkan paling atas.
func (r *Repository) List(ctx context.Context, f Filter, limit, offset int) ([]AdminItem, int, error) {
	where, args := f.where()

	// Pencarian menyentuh kolom orders dan users, jadi hitungan ikut di-join.
	countFrom := "payments p"
	if f.Q != "" {
		countFrom += " JOIN orders o ON o.id = p.order_id JOIN users u ON u.id = o.user_id"
	}
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM "+countFrom+where, args...).Scan(&total); err != nil {
		return nil, 0, wrap("count", err)
	}

	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx,
		fmt.Sprintf(`SELECT %s, o.order_number, o.order_status, o.total_amount::text, o.payment_due_at, u.name, u.email
		   FROM payments p
		   JOIN orders o ON o.id = p.order_id
		   JOIN users u ON u.id = o.user_id%s
		 ORDER BY (p.status = 'pending' AND p.proof_url IS NOT NULL) DESC,
		          p.paid_at DESC NULLS LAST, p.created_at DESC, p.id
		 LIMIT $%d OFFSET $%d`, paymentColumns, where, len(args)-1, len(args)),
		args...)
	if err != nil {
		return nil, 0, wrap("list", err)
	}
	defer rows.Close()

	out := make([]AdminItem, 0, limit)
	for rows.Next() {
		var it AdminItem
		if err := rows.Scan(append(paymentDest(&it.Payment),
			&it.OrderNumber, &it.OrderStatus, &it.OrderTotal, &it.PaymentDueAt, &it.CustomerName, &it.CustomerEmail)...); err != nil {
			return nil, 0, wrap("list scan", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, wrap("list rows", err)
	}
	return out, total, nil
}

// where menyusun klausa WHERE daftar pembayaran beserta argumennya. Alias:
// p = payments, o = orders, u = users.
func (f Filter) where() (string, []any) {
	var conds []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if f.Status != "" {
		add("p.status = $%d", f.Status)
	}
	if f.OrderID != "" {
		add("p.order_id = $%d", f.OrderID)
	}
	if f.Q != "" {
		add(`(o.order_number ILIKE $%[1]d ESCAPE '\' OR p.id::text ILIKE $%[1]d ESCAPE '\'`+
			` OR u.name ILIKE $%[1]d ESCAPE '\' OR u.email ILIKE $%[1]d ESCAPE '\')`, "%"+database.EscapeLike(f.Q)+"%")
	}
	if len(conds) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

func (r *Repository) Review(ctx context.Context, paymentID, adminID, status string) (Payment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Payment{}, wrap("review begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		current, orderID, orderStatus string
		hasProof                      bool
	)
	err = tx.QueryRow(ctx,
		`SELECT p.status, p.proof_url IS NOT NULL, p.order_id, o.order_status
		   FROM payments p
		   JOIN orders o ON o.id = p.order_id
		  WHERE p.id = $1
		    FOR UPDATE OF p, o`,
		paymentID).Scan(&current, &hasProof, &orderID, &orderStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	if err != nil {
		return Payment{}, wrap("review lock", err)
	}
	if !hasProof {
		return Payment{}, ErrNoProof
	}
	if current != StatusPending {
		return Payment{}, ErrAlreadyReviewed
	}

	if status == StatusApproved {
		if orderStatus != orderStatusPendingPayment {
			return Payment{}, ErrOrderNotPayable
		}
		if _, err := tx.Exec(ctx,
			`UPDATE orders SET order_status = $2, updated_at = now() WHERE id = $1`, orderID, orderStatusPaid); err != nil {
			return Payment{}, wrap("review update order", err)
		}
	}

	p, err := scanPayment(tx.QueryRow(ctx,
		`UPDATE payments AS p
		    SET status = $2, reviewed_by = $3, reviewed_at = now()
		  WHERE p.id = $1
		  RETURNING `+paymentColumns,
		paymentID, status, adminID))
	if err != nil {
		return Payment{}, wrap("review update payment", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Payment{}, wrap("review commit", err)
	}
	return p, nil
}

// contactOf mengambil email & nama pemilik order serta nomor ordernya.
func (r *Repository) contactOf(ctx context.Context, orderID string) (contact, error) {
	var c contact
	err := r.db.QueryRow(ctx,
		`SELECT u.email, u.name, o.order_number
		   FROM orders o JOIN users u ON u.id = o.user_id
		  WHERE o.id = $1`, orderID).Scan(&c.Email, &c.Name, &c.OrderNumber)
	if err != nil {
		return contact{}, fmt.Errorf("payment: contact: %w", err)
	}
	return c, nil
}

// paymentDest adalah tujuan Scan dengan urutan yang sama dengan paymentColumns.
func paymentDest(p *Payment) []any {
	return []any{&p.ID, &p.OrderID, &p.Method, &p.Amount, &p.UniqueCode, &p.ProofType,
		&p.HasProof, &p.Status, &p.PaidAt, &p.ReviewedBy, &p.ReviewedAt, &p.CreatedAt}
}

func scanPayment(row pgx.Row) (Payment, error) {
	var p Payment
	err := row.Scan(paymentDest(&p)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	return p, err
}

// wrap menambahkan konteks operasi tetapi membiarkan error domain tetap utuh.
func wrap(op string, err error) error {
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrOrderNotFound) {
		return err
	}
	return fmt.Errorf("payment: %s: %w", op, err)
}
