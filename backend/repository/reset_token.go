package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Alur: services.AuthService -> ResetTokenRepository -> db.DB (PostgreSQL).
// Yang disimpan hanya hash token. Token asli cuma ada di email milik pengguna.

type ResetTokenRepository struct {
	db *sql.DB
}

func NewResetTokenRepository(db *sql.DB) *ResetTokenRepository {
	return &ResetTokenRepository{db: db}
}

// Create menyimpan token reset baru untuk satu user.
func (r *ResetTokenRepository) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("simpan token reset: %w", err)
	}
	return nil
}

// ResetPassword memakai token (sekali pakai) lalu mengganti kata sandi, semuanya dalam satu transaksi:
// kalau salah satu langkah gagal, tidak ada yang berubah.
// Mengembalikan ErrNotFound kalau token tidak ada, sudah dipakai, atau sudah kedaluwarsa.
func (r *ResetTokenRepository) ResetPassword(ctx context.Context, tokenHash, newPasswordHash string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai transaksi reset: %w", err)
	}
	defer tx.Rollback() // aman dipanggil setelah Commit; hanya berefek kalau Commit belum terjadi

	// Tandai token terpakai hanya kalau masih berlaku. Dua request bersamaan dengan token yang sama:
	// satu saja yang lolos, karena UPDATE ini mengunci barisnya.
	var userID int64
	err = tx.QueryRowContext(ctx,
		`UPDATE password_reset_tokens SET used_at = $2
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2
		 RETURNING user_id`,
		tokenHash, now).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("pakai token reset: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET password_hash = $2 WHERE id = $1`, userID, newPasswordHash); err != nil {
		return fmt.Errorf("ganti password: %w", err)
	}

	// Token lain milik user ini ikut dimatikan supaya tautan lama tidak bisa dipakai lagi.
	if _, err := tx.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = $2 WHERE user_id = $1 AND used_at IS NULL`,
		userID, now); err != nil {
		return fmt.Errorf("matikan token lain: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reset password: %w", err)
	}
	return nil
}
