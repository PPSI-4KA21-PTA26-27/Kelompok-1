package post

import "context"

// Service berisi logika artikel.
type Service struct {
	repo *Repository
}

// NewService membuat Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List mengembalikan daftar artikel.
func (s *Service) List(ctx context.Context, f ListFilter, limit, offset int) ([]Post, int, error) {
	return s.repo.List(ctx, f, limit, offset)
}

// GetPublic mengambil artikel terbit berdasarkan slug.
func (s *Service) GetPublic(ctx context.Context, slug string) (Detail, error) {
	return s.repo.Get(ctx, slug, true, true)
}

// Get mengambil artikel (semua status) berdasarkan id.
func (s *Service) Get(ctx context.Context, id string) (Detail, error) {
	return s.repo.Get(ctx, id, false, false)
}

// Create membuat artikel baru dengan penulis authorID.
func (s *Service) Create(ctx context.Context, authorID string, in Input) (Detail, error) {
	id, err := s.repo.Save(ctx, "", authorID, in)
	if err != nil {
		return Detail{}, err
	}
	return s.Get(ctx, id)
}

// Update menerapkan perubahan sebagian (apply) di atas artikel saat ini.
// validate dijalankan pada hasil akhir.
func (s *Service) Update(ctx context.Context, id string, apply func(*Input), validate func(Input) error) (Detail, error) {
	cur, err := s.Get(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	in := Input{
		Slug: cur.Slug, Title: cur.Title, Excerpt: cur.Excerpt, Content: cur.Content,
		CoverImageURL: cur.CoverImageURL, Tag: cur.Tag, Status: cur.Status,
		PublishedAt: cur.PublishedAt, ProductIDs: cur.ProductIDs,
	}
	apply(&in)
	if err := validate(in); err != nil {
		return Detail{}, err
	}
	if _, err := s.repo.Save(ctx, id, "", in); err != nil {
		return Detail{}, err
	}
	return s.Get(ctx, id)
}

// Delete menghapus artikel.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
