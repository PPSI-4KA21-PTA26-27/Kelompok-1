package user

import (
	"context"
	"fmt"

	"hii-stranger1/backend/internal/pkg/password"
)

// Service berisi logika bisnis pengguna: baca profil, ubah profil, dan ganti
// password.
type Service struct {
	repo *Repository
}

// NewService membuat Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Get mengambil user berdasarkan ID.
func (s *Service) Get(ctx context.Context, id string) (User, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateProfile mengubah nama dan/atau telepon user. Mengembalikan ErrNotFound
// bila akun tidak ada atau tidak aktif.
func (s *Service) UpdateProfile(ctx context.Context, id string, c ProfileChanges) (User, error) {
	return s.repo.UpdateProfile(ctx, id, c)
}

// ChangePassword mengganti password setelah memverifikasi password saat ini.
// Mengembalikan ErrNotFound bila akun tidak ada atau tidak aktif, dan
// ErrWrongPassword bila password saat ini salah.
func (s *Service) ChangePassword(ctx context.Context, id, current, next string) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if u.Status != StatusActive {
		return ErrNotFound
	}
	if !password.Verify(u.PasswordHash, current) {
		return ErrWrongPassword
	}

	hash, err := password.Hash(next)
	if err != nil {
		return fmt.Errorf("user: hash password: %w", err)
	}
	return s.repo.UpdatePasswordHash(ctx, id, hash)
}
