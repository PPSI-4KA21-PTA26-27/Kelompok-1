package shipment

import (
	"context"

	"hii-stranger1/backend/internal/pkg/mailer"
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

// Upsert membuat atau memperbarui pengiriman (admin), email resi
// dikirim modul order saat status jadi dikirim.
func (s *Service) Upsert(ctx context.Context, in Shipment) (Shipment, bool, error) {
	ok, err := s.repo.Shippable(ctx, in.OrderID)
	if err != nil {
		return Shipment{}, false, err
	}
	if !ok {
		return Shipment{}, false, ErrOrderNotShippable
	}
	out, created, err := s.repo.Upsert(ctx, in)
	if err != nil {
		return Shipment{}, false, err
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
