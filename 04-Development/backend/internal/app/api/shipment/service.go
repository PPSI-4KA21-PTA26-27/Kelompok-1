package shipment

import (
	"context"
	"log/slog"

	"hii-stranger1/backend/internal/pkg/mailer"
	"hii-stranger1/backend/internal/pkg/tracking"
)

// Service berisi logika bisnis pengiriman.
type Service struct {
	repo *Repository
	mail *mailer.Mailer
}

// NewService membuat Service. mail boleh nonaktif (SMTP belum diisi).
func NewService(repo *Repository, mail *mailer.Mailer) *Service {
	return &Service{repo: repo, mail: mail}
}

// Get mengambil pengiriman order. Hanya pemilik order atau admin; selain itu
// dianggap tidak ada.
func (s *Service) Get(ctx context.Context, orderID, userID string, admin bool) (Shipment, error) {
	owner, err := s.repo.OrderOwner(ctx, orderID)
	if err != nil {
		return Shipment{}, err
	}
	if !admin && owner != userID {
		return Shipment{}, ErrOrderNotFound
	}
	return s.repo.GetByOrder(ctx, orderID)
}

// Upsert membuat atau memperbarui pengiriman (admin), lalu mengirim info
// kurir dan resi ke email pelanggan.
func (s *Service) Upsert(ctx context.Context, in Shipment) (Shipment, bool, error) {
	out, created, err := s.repo.Upsert(ctx, in)
	if err != nil {
		return Shipment{}, false, err
	}

	if c, err := s.repo.contactOf(ctx, out.OrderID); err == nil {
		e := mailer.Email{
			Heading:    "Info pengiriman diperbarui",
			Paragraphs: []string{"Halo " + c.Name + ", ada pembaruan pengiriman untuk pesanan " + c.OrderNumber + "."},
			Details: [][2]string{
				{"No. order", c.OrderNumber},
				{"Kurir", out.CourierName},
				{"Status kirim", shippingLabel(out.ShippingStatus)},
			},
			Button: &mailer.Link{Label: "Lihat pesanan", URL: s.mail.AppURL("/orders/" + out.OrderID)},
		}
		if out.TrackingNumber != nil {
			e.Details = append(e.Details, [2]string{"No. resi", *out.TrackingNumber})
			for _, l := range tracking.Links(out.CourierName, *out.TrackingNumber) {
				e.Links = append(e.Links, mailer.Link{Label: l.Label, URL: l.URL})
			}
		}
		s.mail.NotifyEmail(c.Email, "Update pengiriman pesanan "+c.OrderNumber, e)
	} else {
		slog.WarnContext(ctx, "shipment: load contact for email failed", "error", err)
	}
	return out, created, nil
}

func shippingLabel(s string) string {
	switch s {
	case ShippingPending:
		return "Menunggu pickup"
	case ShippingShipped:
		return "Dalam perjalanan"
	case ShippingDelivered:
		return "Diterima"
	case ShippingReturned:
		return "Dikembalikan"
	}
	return s
}
