package sitesetting

import "context"

// Service berisi logika pengaturan situs.
type Service struct {
	repo *Repository
}

// NewService membuat Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Get mengembalikan pengaturan situs.
func (s *Service) Get(ctx context.Context) (Settings, error) {
	return s.repo.Get(ctx)
}

// Update menerapkan perubahan sebagian di atas pengaturan saat ini.
func (s *Service) Update(ctx context.Context, apply func(*Settings)) (Settings, error) {
	cur, err := s.repo.Get(ctx)
	if err != nil {
		return Settings{}, err
	}
	apply(&cur)
	return s.repo.Save(ctx, cur)
}
