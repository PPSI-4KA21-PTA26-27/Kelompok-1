package product

import "context"

// Service berisi logika bisnis produk.
type Service struct {
	repo *Repository
}

// NewService membuat Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List mengambil satu halaman produk beserta total data yang cocok. Non-admin
// hanya melihat produk berstatus active, apa pun isi f.Status.
func (s *Service) List(ctx context.Context, f Filter, admin bool, limit, offset int) ([]Product, int, error) {
	if !admin {
		f.Status = StatusActive
	}
	return s.repo.List(ctx, f, limit, offset)
}

// Get mengambil satu produk. Produk yang tidak active dianggap tidak ada
// (ErrNotFound) kecuali pemanggil admin.
func (s *Service) Get(ctx context.Context, id string, admin bool) (Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Product{}, err
	}
	if p.Status != StatusActive && !admin {
		return Product{}, ErrNotFound
	}
	return p, nil
}

// GetDetail mengambil produk beserta seluruh varian dan gambarnya. Aturan
// visibilitas sama dengan Get.
func (s *Service) GetDetail(ctx context.Context, id string, admin bool) (Detail, error) {
	p, variants, images, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	if p.Status != StatusActive && !admin {
		return Detail{}, ErrNotFound
	}
	return Detail{Response: p.Response(), Variants: variants, Images: images}, nil
}

// Create membuat produk baru. Produk pre-order wajib punya PoEstimateDays.
func (s *Service) Create(ctx context.Context, p Product) (Product, error) {
	if p.AvailabilityType == AvailabilityPreorder && p.PoEstimateDays == nil {
		return Product{}, ErrPreorderEstimate
	}
	return s.repo.Create(ctx, p)
}

// Update mengubah produk. Hasil akhirnya tetap harus memenuhi aturan pre-order.
func (s *Service) Update(ctx context.Context, id string, ch Changes) (Product, error) {
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Product{}, err
	}
	availability, estimate := cur.AvailabilityType, cur.PoEstimateDays
	if ch.AvailabilityType != nil {
		availability = *ch.AvailabilityType
	}
	if ch.PoEstimateDays != nil {
		estimate = ch.PoEstimateDays
	}
	if availability == AvailabilityPreorder && estimate == nil {
		return Product{}, ErrPreorderEstimate
	}
	return s.repo.Update(ctx, id, ch)
}

// Delete menghapus produk. Mengembalikan ErrNotFound bila tidak ada dan ErrInUse
// bila sudah pernah dipesan.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
