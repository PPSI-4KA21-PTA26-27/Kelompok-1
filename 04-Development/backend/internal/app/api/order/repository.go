package order

import (
	"context"
	"errors"
	"fmt"
	"hii-stranger1/backend/internal/pkg/database"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"hii-stranger1/backend/internal/app/api/product"
)

const (
	uniqueViolation = "23505"

	itemColumns = `oi.id, oi.order_id, oi.product_id, oi.variant_id, oi.product_name, oi.size, oi.color,
		oi.unit_price::text, oi.quantity, oi.subtotal::text,
		(SELECT pi.image_url FROM product_images pi WHERE pi.product_id = oi.product_id
		  ORDER BY pi.is_primary DESC, pi.sort_order, pi.id LIMIT 1)`
	itemsQuery = "SELECT " + itemColumns + " FROM order_items oi WHERE oi.order_id = $1 ORDER BY oi.product_name, oi.size, oi.id"

	// unitPrice adalah harga setelah diskon, dibulatkan 2 desimal (sama dengan keranjang).
	unitPrice = `ROUND(p.price * (100 - p.discount_percent) / 100, 2)`

	lowStockThreshold = 3
)

// orderFields adalah kolom order dengan urutan yang sama dengan scanOrder.
var orderFields = []string{
	"id", "order_number", "user_id", "recipient_name", "recipient_phone", "shipping_address",
	"shipping_city", "shipping_kelurahan", "subtotal::text", "shipping_cost::text", "total_amount::text", "order_status",
	"shipping_district", "shipping_subdistrict_id", "shipping_zip_code",
	"courier_code", "courier_service", "shipping_etd", "shipping_weight_gram",
	"created_at", "updated_at", "payment_due_at", "cancel_reason",
}

// orderColumns dipakai pada query/RETURNING dengan alias tabel o.
var orderColumns = "o." + strings.Join(orderFields, ", o.")

// Repository mengakses tabel orders dan order_items.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type cartLine struct {
	CartItemID    string
	VariantID     string
	Quantity      int
	Stock         int
	Size          string
	Color         *string
	ProductID     string
	ProductName   string
	ProductStatus string
	UnitPrice     string
	Subtotal      string
	WeightGram    int // berat produk x jumlah
}

// Checkout membuat order + item + tagihan QRIS, mengurangi stok, dan
// menghapus item keranjang dalam satu transaksi. Hanya pelanggan aktif yang
// emailnya sudah terverifikasi yang boleh checkout (dicek di sini supaya
// tidak bergantung pada isi token).
func (r *Repository) Checkout(ctx context.Context, userID, orderNumber string, ship shipping, in CheckoutInput) (Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Order{}, wrap("checkout begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := checkCustomer(ctx, tx, userID); err != nil {
		return Order{}, err
	}

	lines, err := lockCartLines(ctx, tx, userID, in.CartItemIDs)
	if err != nil {
		return Order{}, err
	}
	if len(lines) == 0 {
		return Order{}, ErrCartEmpty
	}
	if len(in.CartItemIDs) > 0 && len(lines) != len(in.CartItemIDs) {
		return Order{}, ErrCartItemNotFound
	}

	need := map[string]int{}
	subtotal := new(big.Rat)
	cartIDs := make([]string, 0, len(lines))
	weight := 0
	for _, l := range lines {
		weight += l.WeightGram
		if l.ProductStatus != product.StatusActive {
			return Order{}, ErrProductUnavailable
		}
		need[l.VariantID] += l.Quantity
		if need[l.VariantID] > l.Stock {
			return Order{}, ErrInsufficientStock
		}
		sub, ok := new(big.Rat).SetString(l.Subtotal)
		if !ok {
			return Order{}, fmt.Errorf("order: checkout: subtotal %q tidak valid", l.Subtotal)
		}
		subtotal.Add(subtotal, sub)
		cartIDs = append(cartIDs, l.CartItemID)
	}
	if weight != ship.WeightGram {
		return Order{}, ErrCartChanged
	}

	cost := ship.Option.Cost
	shippingCost, ok := new(big.Rat).SetString(cost)
	if !ok {
		return Order{}, fmt.Errorf("order: checkout: shipping cost %q tidak valid", cost)
	}
	total := new(big.Rat).Add(subtotal, shippingCost)

	window := defaultPaymentWindow

	a := ship.Address
	o, err := scanOrder(tx.QueryRow(ctx,
		`INSERT INTO orders AS o
		   (order_number, user_id, recipient_name, recipient_phone, shipping_address,
		    shipping_city, shipping_district, shipping_kelurahan,
		    shipping_subdistrict_id, shipping_zip_code,
		    courier_code, courier_service, shipping_etd, shipping_weight_gram,
		    subtotal, shipping_cost, total_amount, order_status, payment_due_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
		         $15::numeric, $16::numeric, $17::numeric, $18, now() + make_interval(hours => $19))
		 RETURNING `+orderColumns,
		orderNumber, userID, in.RecipientName, in.RecipientPhone, in.ShippingAddress,
		a.City, a.District, a.Subdistrict,
		a.SubdistrictID, database.NullIfEmpty(a.ZipCode),
		ship.Option.CourierCode, ship.Option.Service, database.NullIfEmpty(ship.Option.ETD), ship.WeightGram,
		subtotal.FloatString(2), cost, total.FloatString(2), StatusPendingPayment, window))
	if err != nil {
		return Order{}, wrap("checkout insert order", err)
	}

	for _, l := range lines {
		_, err := tx.Exec(ctx,
			`INSERT INTO order_items
			   (order_id, product_id, variant_id, product_name, size, color, unit_price, quantity, subtotal)
			 VALUES ($1, $2, $3, $4, $5, $6, $7::numeric, $8, $9::numeric)`,
			o.ID, l.ProductID, l.VariantID, l.ProductName, l.Size, l.Color, l.UnitPrice, l.Quantity, l.Subtotal)
		if err != nil {
			return Order{}, wrap("checkout insert item", err)
		}
	}

	for variantID, qty := range need {
		if _, err := tx.Exec(ctx,
			`UPDATE product_variants SET stock = stock - $2 WHERE id = $1`, variantID, qty); err != nil {
			return Order{}, wrap("checkout reduce stock", err)
		}
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM cart_items WHERE id = ANY($1::uuid[])`, cartIDs); err != nil {
		return Order{}, wrap("checkout clear cart", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, wrap("checkout commit", err)
	}
	return o, nil
}

// CanOrder memastikan user adalah pelanggan aktif dengan email terverifikasi.
func (r *Repository) CanOrder(ctx context.Context, userID string) error {
	return checkCustomer(ctx, r.db, userID)
}

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func checkCustomer(ctx context.Context, q queryRower, userID string) error {
	var role, status string
	err := q.QueryRow(ctx,
		`SELECT (SELECT r.code FROM roles r WHERE r.id = users.role_id), status FROM users WHERE id = $1`, userID).
		Scan(&role, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotCustomer
	}
	if err != nil {
		return wrap("check customer", err)
	}
	if role != "customer" || status != "active" {
		return ErrNotCustomer
	}
	return nil
}

// lockCartLines membaca item keranjang (semua atau sesuai ids) dan mengunci
// baris keranjang & varian sampai transaksi selesai.
func lockCartLines(ctx context.Context, tx pgx.Tx, userID string, ids []string) ([]cartLine, error) {
	query := `SELECT ci.id, v.id, ci.quantity, v.stock, v.size, v.color, p.id, p.name, p.status,
	                 ` + unitPrice + `::text, (` + unitPrice + ` * ci.quantity)::text,
	                 p.weight_gram * ci.quantity
	            FROM cart_items ci
	            JOIN product_variants v ON v.id = ci.variant_id
	            JOIN products p ON p.id = v.product_id
	           WHERE ci.user_id = $1`
	args := []any{userID}
	if len(ids) > 0 {
		query += ` AND ci.id = ANY($2::uuid[])`
		args = append(args, ids)
	}
	query += ` ORDER BY v.id, ci.id FOR UPDATE OF ci, v`

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, wrap("checkout lock cart", err)
	}
	defer rows.Close()

	var lines []cartLine
	for rows.Next() {
		var l cartLine
		if err := rows.Scan(&l.CartItemID, &l.VariantID, &l.Quantity, &l.Stock, &l.Size, &l.Color,
			&l.ProductID, &l.ProductName, &l.ProductStatus, &l.UnitPrice, &l.Subtotal, &l.WeightGram); err != nil {
			return nil, wrap("checkout scan cart", err)
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return nil, wrap("checkout cart rows", err)
	}
	return lines, nil
}

const listSelect = `SELECT %s,
	       (SELECT COALESCE(SUM(oi.quantity), 0) FROM order_items oi WHERE oi.order_id = o.id),
	       pay.status, COALESCE(pay.has_proof, false), u.email
	  FROM orders o
	  JOIN users u ON u.id = o.user_id
	  LEFT JOIN LATERAL (SELECT p.status, p.proof_url IS NOT NULL AS has_proof
	                       FROM payments p WHERE p.order_id = o.id
	                      ORDER BY p.created_at, p.id LIMIT 1) pay ON true`

// List mengembalikan order sesuai filter, terbaru dulu.
func (r *Repository) List(ctx context.Context, f Filter, limit, offset int) ([]ListItem, int, error) {
	var conds []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, strings.ReplaceAll(cond, "$?", fmt.Sprintf("$%d", len(args))))
	}
	if f.UserID != "" {
		add("o.user_id = $?", f.UserID)
	}
	if f.Status != "" {
		add("o.order_status = $?", f.Status)
	}
	if f.Q != "" {
		add(`(o.order_number ILIKE $? ESCAPE '\' OR o.recipient_name ILIKE $? ESCAPE '\')`, "%"+database.EscapeLike(f.Q)+"%")
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	pageArgs := append(append([]any{}, args...), limit, offset)
	b := &pgx.Batch{}
	b.Queue("SELECT COUNT(*) FROM orders o"+where, args...)
	b.Queue(fmt.Sprintf(listSelect, orderColumns)+where+
		fmt.Sprintf(" ORDER BY o.created_at DESC, o.id LIMIT $%d OFFSET $%d", len(pageArgs)-1, len(pageArgs)), pageArgs...)
	br := r.db.SendBatch(ctx, b)
	defer func() { _ = br.Close() }()

	var total int
	if err := br.QueryRow().Scan(&total); err != nil {
		return nil, 0, wrap("count", err)
	}
	rows, err := br.Query()
	if err != nil {
		return nil, 0, wrap("list", err)
	}
	out, err := collectListItems(rows)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func collectListItems(rows pgx.Rows) ([]ListItem, error) {
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ListItem, error) {
		var it ListItem
		dst := append(orderDest(&it.Order), &it.ItemCount, &it.PaymentStatus, &it.HasProof, &it.CustomerEmail)
		err := row.Scan(dst...)
		return it, err
	})
	if err != nil {
		return nil, wrap("list scan", err)
	}
	if out == nil {
		out = []ListItem{}
	}
	return out, nil
}

// Items mengembalikan item sebuah order.
func (r *Repository) Items(ctx context.Context, orderID string) ([]Item, error) {
	rows, err := r.db.Query(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, wrap("items", err)
	}
	return collectItems(rows)
}

// GetWithItems mengambil order dan item-itemnya dalam satu round trip.
func (r *Repository) GetWithItems(ctx context.Context, id string) (Order, []Item, error) {
	b := &pgx.Batch{}
	b.Queue("SELECT "+orderColumns+" FROM orders o WHERE o.id = $1", id)
	b.Queue(itemsQuery, id)
	br := r.db.SendBatch(ctx, b)
	defer func() { _ = br.Close() }()

	o, err := scanOrder(br.QueryRow())
	if err != nil {
		return Order{}, nil, wrap("get with items", err)
	}
	rows, err := br.Query()
	if err != nil {
		return Order{}, nil, wrap("get with items", err)
	}
	items, err := collectItems(rows)
	if err != nil {
		return Order{}, nil, err
	}
	return o, items, nil
}

func collectItems(rows pgx.Rows) ([]Item, error) {
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Item, error) {
		var it Item
		err := row.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.VariantID, &it.ProductName,
			&it.Size, &it.Color, &it.UnitPrice, &it.Quantity, &it.Subtotal, &it.ImageURL)
		return it, err
	})
	if err != nil {
		return nil, wrap("items scan", err)
	}
	if out == nil {
		out = []Item{}
	}
	return out, nil
}

// change menggambarkan satu perubahan status order.
type change struct {
	to string
	// ownerID, bila diisi, mewajibkan order milik user tersebut.
	ownerID string
	// allowed menggantikan aturan transitions (mis. aksi pelanggan).
	allowed func(from string) bool
	// guard menjalankan pemeriksaan tambahan di dalam transaksi.
	guard  func(ctx context.Context, tx pgx.Tx, id string) error
	reason *string
	// onUpdate dijalankan setelah status diperbarui (mis. simpan resi).
	onUpdate func(context.Context, pgx.Tx, Order) error
}

// transition mengubah status order secara atomik: mengunci order, memeriksa
// aturan, mengembalikan stok bila dibatalkan, lalu menyimpan status baru.
func (r *Repository) transition(ctx context.Context, id string, c change) (Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Order{}, wrap("update status begin", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var from, owner string
	err = tx.QueryRow(ctx, `SELECT order_status, user_id FROM orders WHERE id = $1 FOR UPDATE`, id).Scan(&from, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	if err != nil {
		return Order{}, wrap("update status lock", err)
	}
	if c.ownerID != "" && owner != c.ownerID {
		return Order{}, ErrNotFound
	}
	allowed := canTransition(from, c.to)
	if c.allowed != nil {
		allowed = c.allowed(from)
	}
	if !allowed {
		return Order{}, ErrInvalidTransition
	}
	if c.guard != nil {
		if err := c.guard(ctx, tx, id); err != nil {
			return Order{}, err
		}
	}

	if c.to == StatusCancelled {
		// Stok dikembalikan sesuai jumlah yang dipesan.
		_, err := tx.Exec(ctx,
			`UPDATE product_variants v
			    SET stock = v.stock + s.qty
			   FROM (SELECT variant_id, SUM(quantity) AS qty
			           FROM order_items WHERE order_id = $1
			          GROUP BY variant_id) s
			  WHERE v.id = s.variant_id`, id)
		if err != nil {
			return Order{}, wrap("update status restore stock", err)
		}
	}

	o, err := scanOrder(tx.QueryRow(ctx,
		`UPDATE orders AS o
		    SET order_status = $2, updated_at = now(),
		        cancel_reason = CASE WHEN $2 = 'cancelled' THEN $3 ELSE o.cancel_reason END
		  WHERE o.id = $1
		  RETURNING `+orderColumns, id, c.to, c.reason))
	if err != nil {
		return Order{}, wrap("update status", err)
	}
	if c.onUpdate != nil {
		if err := c.onUpdate(ctx, tx, o); err != nil {
			return Order{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, wrap("update status commit", err)
	}
	return o, nil
}

// noProofUnderReview menolak pembatalan bila bukti bayar sedang menunggu
// verifikasi admin (pelanggan mungkin sudah membayar).
func noProofUnderReview(ctx context.Context, tx pgx.Tx, id string) error {
	var reviewing bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM payments
		                 WHERE order_id = $1 AND status = 'menunggu' AND proof_url IS NOT NULL)`, id).Scan(&reviewing)
	if err != nil {
		return wrap("check proof", err)
	}
	if reviewing {
		return ErrProofUnderReview
	}
	return nil
}

// ExpiredIDs mengembalikan order menunggu pembayaran yang batas bayarnya lewat.
func (r *Repository) ExpiredIDs(ctx context.Context, limit int) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id FROM orders
		  WHERE order_status = $1 AND payment_due_at < now()
		  ORDER BY payment_due_at LIMIT $2`, StatusPendingPayment, limit)
	if err != nil {
		return nil, wrap("expired ids", err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, wrap("expired ids scan", err)
	}
	return ids, nil
}

// stillExpired memastikan order masih lewat batas bayar saat dikunci.
func stillExpired(ctx context.Context, tx pgx.Tx, id string) error {
	var expired bool
	if err := tx.QueryRow(ctx,
		`SELECT payment_due_at < now() FROM orders WHERE id = $1`, id).Scan(&expired); err != nil {
		return wrap("check expiry", err)
	}
	if !expired {
		return ErrInvalidTransition
	}
	return noProofUnderReview(ctx, tx, id)
}

func (r *Repository) contactOf(ctx context.Context, orderID string) (contact, error) {
	var c contact
	err := r.db.QueryRow(ctx,
		`SELECT u.email, u.name FROM orders o JOIN users u ON u.id = o.user_id WHERE o.id = $1`,
		orderID).Scan(&c.Email, &c.Name)
	if err != nil {
		return contact{}, fmt.Errorf("order: contact: %w", err)
	}
	return c, nil
}

func orderDest(o *Order) []any {
	return []any{&o.ID, &o.OrderNumber, &o.UserID, &o.RecipientName, &o.RecipientPhone,
		&o.ShippingAddress, &o.ShippingCity, &o.ShippingKelurahan, &o.Subtotal, &o.ShippingCost,
		&o.TotalAmount, &o.OrderStatus,
		&o.ShippingDistrict, &o.ShippingSubdistrictID, &o.ShippingZipCode,
		&o.CourierCode, &o.CourierService, &o.ShippingETD, &o.ShippingWeightGram,
		&o.CreatedAt, &o.UpdatedAt, &o.PaymentDueAt, &o.CancelReason}
}

func scanOrder(row pgx.Row) (Order, error) {
	var o Order
	err := row.Scan(orderDest(&o)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	return o, err
}

// paymentDeadlineLabel memformat batas bayar untuk email (WIB).
func paymentDeadlineLabel(t *time.Time) string {
	if t == nil {
		return "-"
	}
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	return t.In(loc).Format("02 Jan 2006, 15:04") + " WIB"
}

func wrap(op string, err error) error {
	if errors.Is(err, ErrNotFound) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "uq_orders_order_number" {
		return errDuplicateNumber
	}
	return fmt.Errorf("order: %s: %w", op, err)
}
