package banner

import "context"

// Service berisi logika banner.
type Service struct {
	repo *Repository
}

// NewService membuat Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List mengembalikan banner sesuai filter.
func (s *Service) List(ctx context.Context, f ListFilter) ([]Banner, error) {
	return s.repo.List(ctx, f)
}

// Create menyimpan banner baru.
func (s *Service) Create(ctx context.Context, b Banner) (Banner, error) {
	return s.repo.Create(ctx, b)
}

// Update menerapkan perubahan sebagian. validate dipanggil pada hasil akhir
// (mis. memastikan jadwal selesai setelah jadwal mulai).
func (s *Service) Update(ctx context.Context, id string, ch Changes, validate func(Banner) error) (Banner, error) {
	b, err := s.repo.Get(ctx, id)
	if err != nil {
		return Banner{}, err
	}
	ch.apply(&b)
	if err := validate(b); err != nil {
		return Banner{}, err
	}
	return s.repo.Save(ctx, b)
}

// Delete menghapus banner.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (ch Changes) apply(b *Banner) {
	if ch.Placement != nil {
		b.Placement = *ch.Placement
	}
	if ch.ImageURL != nil {
		b.ImageURL = *ch.ImageURL
	}
	set := func(dst **string, o optText) {
		if o.Set {
			*dst = o.Value
		}
	}
	set(&b.Title, ch.Title)
	set(&b.Subtitle, ch.Subtitle)
	set(&b.CTALabel, ch.CTALabel)
	set(&b.MobileImageURL, ch.MobileImageURL)
	set(&b.LinkURL, ch.LinkURL)
	if ch.SortOrder != nil {
		b.SortOrder = *ch.SortOrder
	}
	if ch.IsActive != nil {
		b.IsActive = *ch.IsActive
	}
	if ch.StartsAt.Set {
		b.StartsAt = ch.StartsAt.Value
	}
	if ch.EndsAt.Set {
		b.EndsAt = ch.EndsAt.Value
	}
}
