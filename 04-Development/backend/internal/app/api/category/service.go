package category

import "context"

// Service berisi logika bisnis kategori.
type Service struct {
	repo *Repository
}

// NewService membuat Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List mengambil satu halaman kategori beserta total data yang cocok. Kategori
// nonaktif hanya ikut bila includeInactive true (pemanggil harus admin).
func (s *Service) List(ctx context.Context, includeInactive bool, limit, offset int) ([]Category, int, error) {
	return s.repo.List(ctx, includeInactive, limit, offset)
}

// Get mengambil satu kategori. Kategori nonaktif dianggap tidak ada (ErrNotFound)
// kecuali includeInactive true (pemanggil harus admin).
func (s *Service) Get(ctx context.Context, id string, includeInactive bool) (Category, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Category{}, err
	}
	if !c.IsActive && !includeInactive {
		return Category{}, ErrNotFound
	}
	return c, nil
}

// Create membuat kategori baru. Mengembalikan ErrNameTaken bila nama sudah ada.
func (s *Service) Create(ctx context.Context, name string, isActive bool) (Category, error) {
	return s.repo.Create(ctx, name, isActive)
}

// Update mengubah kategori. Mengembalikan ErrNotFound atau ErrNameTaken.
func (s *Service) Update(ctx context.Context, id string, ch Changes) (Category, error) {
	return s.repo.Update(ctx, id, ch)
}

// Delete menghapus kategori. Mengembalikan ErrNotFound bila tidak ada.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
