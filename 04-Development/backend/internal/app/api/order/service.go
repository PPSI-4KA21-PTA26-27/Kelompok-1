package order

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"hii-stranger1/backend/internal/app/api/ongkir"
	"hii-stranger1/backend/internal/app/api/payment"
	"hii-stranger1/backend/internal/app/api/shipment"
	"hii-stranger1/backend/internal/pkg/mailer"
	"hii-stranger1/backend/internal/pkg/tracking"
)

const (
	numberAttempts = 3

	// expireInterval adalah jeda pemeriksaan order yang lewat batas bayar.
	expireInterval = 5 * time.Minute
	expireBatch    = 100
)

// Service berisi logika bisnis order.
type Service struct {
	repo      *Repository
	ongkir    *ongkir.Service
	payments  *payment.Repository
	shipments *shipment.Repository
	mail      *mailer.Mailer
}

// NewService membuat Service.
func NewService(
	repo *Repository,
	ongkirSvc *ongkir.Service,
	payments *payment.Repository,
	shipments *shipment.Repository,
	mail *mailer.Mailer,
) *Service {
	return &Service{repo: repo, ongkir: ongkirSvc, payments: payments, shipments: shipments, mail: mail}
}

// Checkout membuat order dari keranjang. Harga, ongkir, dan tagihan QRIS
// dihitung ulang di server; nilai dari client tidak dipercaya.
func (s *Service) Checkout(ctx context.Context, userID string, in CheckoutInput) (Detail, error) {
	// Dicek lebih dulu supaya tidak memanggil RajaOngkir untuk akun yang
	// belum boleh memesan; dicek ulang di dalam transaksi checkout.
	if err := s.repo.CanOrder(ctx, userID); err != nil {
		return Detail{}, err
	}
	weight, err := s.ongkir.CartWeight(ctx, userID, in.CartItemIDs)
	if errors.Is(err, ongkir.ErrCartEmpty) {
		if len(in.CartItemIDs) > 0 {
			return Detail{}, ErrCartItemNotFound
		}
		return Detail{}, ErrCartEmpty
	}
	if err != nil {
		return Detail{}, err
	}
	addr, err := s.ongkir.Resolve(ctx, in.Destination)
	if err != nil {
		return Detail{}, err
	}
	opt, err := s.ongkir.Cheapest(ctx, addr.SubdistrictID, weight)
	if err != nil {
		return Detail{}, err
	}
	ship := shipping{Address: addr, Option: opt, WeightGram: weight}

	var o Order
	for attempt := 1; ; attempt++ {
		number, err := newOrderNumber()
		if err != nil {
			return Detail{}, err
		}
		o, err = s.repo.Checkout(ctx, userID, number, ship, in)
		if errors.Is(err, errDuplicateNumber) && attempt < numberAttempts {
			continue
		}
		if err != nil {
			return Detail{}, err
		}
		break
	}

	d, err := s.Get(ctx, o.ID, userID, false)
	if err != nil {
		return Detail{}, err
	}
	s.notifyCheckout(ctx, d, addr, opt)
	return d, nil
}

func (s *Service) notifyCheckout(ctx context.Context, d Detail, addr ongkir.Address, opt ongkir.Option) {
	courier := strings.ToUpper(opt.CourierCode) + " " + opt.Service
	s.mail.NotifyOwnerEmail("Order baru: "+d.OrderNumber, mailer.Email{
		Heading:    "Order baru masuk",
		Paragraphs: []string{fmt.Sprintf("%s (%s) membuat pesanan %d item. Order otomatis batal bila belum dibayar sampai batas waktu.", d.RecipientName, d.RecipientPhone, len(d.Items))},
		Details: [][2]string{
			{"No. order", d.OrderNumber},
			{"Total", "Rp " + d.TotalAmount},
			{"Kurir", courier + " (Rp " + opt.Cost + ")"},
			{"Tujuan", addr.Subdistrict + ", " + addr.City},
		},
		Button: &mailer.Link{Label: "Buka pesanan di admin", URL: s.mail.AppURL("/admin/pesanan/" + d.ID)},
	})
}

// List mengembalikan daftar order sesuai filter.
func (s *Service) List(ctx context.Context, f Filter, limit, offset int) ([]ListItem, int, error) {
	return s.repo.List(ctx, f, limit, offset)
}

// Get mengembalikan detail order. Non-admin hanya bisa melihat ordernya sendiri.
func (s *Service) Get(ctx context.Context, id, userID string, admin bool) (Detail, error) {
	o, items, err := s.repo.GetWithItems(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	if !admin && o.UserID != userID {
		return Detail{}, ErrNotFound
	}
	d := Detail{Order: o, Items: items}

	p, err := s.payments.GetByOrder(ctx, id)
	switch {
	case err == nil:
		d.Payment = &p
	case !errors.Is(err, payment.ErrNotFound):
		return Detail{}, err
	}
	sh, err := s.shipments.GetByOrder(ctx, id)
	switch {
	case err == nil:
		d.Shipment = &sh
	case !errors.Is(err, shipment.ErrNotFound):
		return Detail{}, err
	}
	return d, nil
}

// Summary mengembalikan ringkasan dashboard admin.
func (s *Service) Summary(ctx context.Context) (Summary, error) {
	return s.repo.Summary(ctx)
}

// UpdateStatus mengubah status order oleh admin. Status shipped mewajibkan no.
// resi dan sekaligus menyimpan data pengiriman dalam transaksi yang sama.
func (s *Service) UpdateStatus(ctx context.Context, id, status string, ship *ShipInfo, reason *string) (Order, error) {
	var sh shipment.Shipment
	c := change{to: status, reason: reason}
	if status == StatusShipped {
		if ship == nil || ship.TrackingNumber == "" {
			return Order{}, ErrTrackingRequired
		}
		c.onUpdate = func(ctx context.Context, tx pgx.Tx, o Order) error {
			tracking := ship.TrackingNumber
			var err error
			sh, _, err = s.shipments.UpsertTx(ctx, tx, shipment.Shipment{
				OrderID:        o.ID,
				TrackingNumber: &tracking,
				CoverageStatus: shipment.CoverageCovered,
				ShippingStatus: shipment.ShippingShipped,
			})
			return err
		}
	}
	if status == StatusCompleted {
		c.onUpdate = markDelivered
	}

	o, err := s.repo.transition(ctx, id, c)
	if err != nil {
		return Order{}, err
	}
	s.notifyStatus(ctx, o, &sh)
	return o, nil
}

// Cancel membatalkan order milik pelanggan selama belum dibayar dan bukti
// bayar tidak sedang diverifikasi. Stok dikembalikan.
func (s *Service) Cancel(ctx context.Context, id, userID string, reason *string) (Order, error) {
	if reason == nil {
		r := "Dibatalkan oleh pelanggan"
		reason = &r
	}
	o, err := s.repo.transition(ctx, id, change{
		to:      StatusCancelled,
		ownerID: userID,
		allowed: func(from string) bool { return from == StatusPendingPayment },
		guard:   noProofUnderReview,
		reason:  reason,
	})
	if err != nil {
		return Order{}, err
	}
	s.notifyStatus(ctx, o, nil)
	return o, nil
}

// Complete dipakai pelanggan untuk konfirmasi paket sudah diterima.
func (s *Service) Complete(ctx context.Context, id, userID string) (Order, error) {
	o, err := s.repo.transition(ctx, id, change{
		to:       StatusCompleted,
		ownerID:  userID,
		allowed:  func(from string) bool { return from == StatusShipped },
		onUpdate: markDelivered,
	})
	if err != nil {
		return Order{}, err
	}
	return o, nil
}

func markDelivered(ctx context.Context, tx pgx.Tx, o Order) error {
	_, err := tx.Exec(ctx,
		`UPDATE shipments SET shipping_status = $2, updated_at = now() WHERE order_id = $1`,
		o.ID, shipment.ShippingDelivered)
	if err != nil {
		return fmt.Errorf("order: mark delivered: %w", err)
	}
	return nil
}

// ExpireUnpaid membatalkan order yang melewati batas bayar (tanpa bukti yang
// sedang diverifikasi) dan mengembalikan stoknya. Mengembalikan jumlah order
// yang dibatalkan.
func (s *Service) ExpireUnpaid(ctx context.Context) (int, error) {
	ids, err := s.repo.ExpiredIDs(ctx, expireBatch)
	if err != nil {
		return 0, err
	}
	reason := "Batas waktu pembayaran habis"
	n := 0
	for _, id := range ids {
		o, err := s.repo.transition(ctx, id, change{
			to:      StatusCancelled,
			allowed: func(from string) bool { return from == StatusPendingPayment },
			guard:   stillExpired,
			reason:  &reason,
		})
		if errors.Is(err, ErrProofUnderReview) || errors.Is(err, ErrInvalidTransition) || errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return n, err
		}
		n++
		s.notifyStatus(ctx, o, nil)
	}
	return n, nil
}

// RunExpiry menjalankan ExpireUnpaid berkala sampai ctx selesai.
func (s *Service) RunExpiry(ctx context.Context) {
	t := time.NewTicker(expireInterval)
	defer t.Stop()
	for {
		if n, err := s.ExpireUnpaid(ctx); err != nil {
			if ctx.Err() == nil {
				slog.Error("order: expire unpaid failed", "error", err)
			}
		} else if n > 0 {
			slog.Info("order: unpaid orders cancelled", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (s *Service) notifyStatus(ctx context.Context, o Order, sh *shipment.Shipment) {
	// Email pelanggan hanya untuk kabar penting: dikirim dan batal (konfirmasi dikirim modul payment).
	if o.OrderStatus != StatusShipped && o.OrderStatus != StatusCancelled {
		return
	}
	c, err := s.repo.contactOf(ctx, o.ID)
	if err != nil {
		slog.WarnContext(ctx, "order: load contact for email failed", "error", err)
		return
	}
	orderURL := s.mail.AppURL("/pesanan/" + o.ID)
	e := mailer.Email{
		Heading:    "Status pesanan: " + statusLabel(o.OrderStatus),
		Paragraphs: []string{"Halo " + c.Name + ", status pesanan " + o.OrderNumber + " sekarang " + strings.ToLower(statusLabel(o.OrderStatus)) + "."},
		Details:    [][2]string{{"No. order", o.OrderNumber}, {"Status", statusLabel(o.OrderStatus)}},
		Button:     &mailer.Link{Label: "Lihat pesanan", URL: orderURL},
	}
	subject := "Status pesanan " + o.OrderNumber + ": " + statusLabel(o.OrderStatus)

	switch o.OrderStatus {
	case StatusShipped:
		if sh == nil || sh.TrackingNumber == nil {
			break
		}
		resi := *sh.TrackingNumber
		subject = "Pesanan " + o.OrderNumber + " sudah dikirim — no. resi " + resi
		e.Heading = "Pesanan kamu sedang dalam perjalanan"
		e.Paragraphs = []string{
			"Halo " + c.Name + ", paket untuk pesanan " + o.OrderNumber + " sudah kami serahkan ke kurir.",
			"Lacak paket dengan nomor resi di bawah. Status di situs kurir biasanya muncul beberapa jam setelah paket diserahkan.",
		}
		e.Details = [][2]string{{"No. order", o.OrderNumber}, {"Kurir", sh.CourierName}, {"No. resi", resi}}
		if o.ShippingETD != nil && *o.ShippingETD != "" {
			e.Details = append(e.Details, [2]string{"Estimasi", etdLabel(*o.ShippingETD)})
		}
		e.Button = &mailer.Link{Label: "Lacak pesanan", URL: orderURL}
		for _, l := range tracking.Links(sh.CourierName, resi) {
			e.Links = append(e.Links, mailer.Link{Label: l.Label, URL: l.URL})
		}
		e.Note = "Sudah menerima paket? Konfirmasi lewat tombol \"Pesanan diterima\" di halaman pesanan."
	case StatusCancelled:
		if o.CancelReason != nil {
			e.Details = append(e.Details, [2]string{"Alasan", *o.CancelReason})
		}
		e.Note = "Bila kamu sudah membayar, balas email ini dengan bukti pembayaran supaya kami bantu proses pengembalian dana."
	}
	s.mail.NotifyEmail(c.Email, subject, e)
}

// etdLabel menormalkan estimasi RajaOngkir ("2 day", "1-2 days", "3") jadi
// "2 hari" / "1-2 hari".
func etdLabel(etd string) string {
	v := strings.TrimSpace(etd)
	for _, suf := range []string{"days", "day", "hari", "HARI"} {
		v = strings.TrimSpace(strings.TrimSuffix(v, suf))
	}
	if v == "" {
		return "-"
	}
	return v + " hari"
}

func canTransition(from, to string) bool {
	return slices.Contains(transitions[from], to)
}

func validStatus(s string) bool {
	switch s {
	case StatusPendingPayment, StatusPaid, StatusProcessing, StatusShipped, StatusCompleted, StatusCancelled:
		return true
	}
	return false
}

func statusLabel(s string) string {
	switch s {
	case StatusPendingPayment:
		return "Menunggu pembayaran"
	case StatusPaid:
		return "Sudah dibayar"
	case StatusProcessing:
		return "Sedang dikemas"
	case StatusShipped:
		return "Dikirim"
	case StatusCompleted:
		return "Selesai"
	case StatusCancelled:
		return "Dibatalkan"
	}
	return s
}

// newOrderNumber membuat nomor order unik: HS + waktu + 6 hex acak.
func newOrderNumber() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("order: random order number: %w", err)
	}
	return "HS" + time.Now().Format("20060102150405") + "-" + strings.ToUpper(hex.EncodeToString(b)), nil
}
