// Package auth menangani registrasi, login, verifikasi email, dan lupa password.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/mailer"
	"hii-stranger1/backend/internal/pkg/password"
	"hii-stranger1/backend/internal/pkg/token"
)

// ErrInvalidCredentials dikembalikan untuk email/password salah atau akun nonaktif.
var ErrInvalidCredentials = errors.New("auth: invalid credentials")

// ErrAlreadyVerified: email user sudah terverifikasi, tidak perlu kirim ulang.
var ErrAlreadyVerified = errors.New("auth: email already verified")

// passwordResetTTL adalah masa berlaku tautan reset password. Dibuat
// pendek karena tautan ini setara dengan akses ke akun.
const passwordResetTTL = time.Hour

// verificationTTL adalah masa berlaku tautan verifikasi email.
const verificationTTL = 24 * time.Hour

// Service berisi logika autentikasi.
type Service struct {
	users  *user.Repository
	tokens *token.Manager
	mail   *mailer.Mailer
}

// NewService membuat Service.
func NewService(users *user.Repository, tokens *token.Manager, mail *mailer.Mailer) *Service {
	return &Service{users: users, tokens: tokens, mail: mail}
}

// RegisterInput adalah data registrasi pelanggan.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Phone    *string
}

// Result adalah token akses beserta profil user.
type Result struct {
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	ExpiresIn   int           `json:"expires_in"`
	User        user.Response `json:"user"`
}

// Register membuat akun pelanggan (belum terverifikasi) lalu mengirim email
// verifikasi. Pelanggan langsung login, tetapi baru boleh memesan setelah
// emailnya terverifikasi (dicek di checkout).
func (s *Service) Register(ctx context.Context, in RegisterInput) (Result, error) {
	hash, err := password.Hash(in.Password)
	if err != nil {
		return Result{}, fmt.Errorf("auth: hash password: %w", err)
	}
	u, err := s.users.Create(ctx, user.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: hash,
		Phone:        in.Phone,
		Role:         user.RoleCustomer,
		Status:       user.StatusActive,
	})
	if err != nil {
		return Result{}, err
	}
	if err := s.sendVerification(ctx, u); err != nil {
		// Akun tetap dibuat; pelanggan bisa meminta kirim ulang dari halaman akun.
		slog.ErrorContext(ctx, "auth: send verification failed", "error", err)
	}
	return s.issue(u)
}

// Login memverifikasi email & password lalu menerbitkan token.
func (s *Service) Login(ctx context.Context, email, plain string) (Result, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		password.VerifyDummy(plain)
		return Result{}, ErrInvalidCredentials
	}
	if err != nil {
		return Result{}, err
	}
	if !password.Verify(u.PasswordHash, plain) || u.Status != user.StatusActive {
		return Result{}, ErrInvalidCredentials
	}
	return s.issue(u)
}

// VerifyEmail memakai token dari tautan email (sekali pakai).
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) (user.User, error) {
	return s.users.ConsumeVerification(ctx, hashToken(rawToken))
}

// ResendVerification mengirim ulang tautan verifikasi untuk user yang login.
func (s *Service) ResendVerification(ctx context.Context, userID string) error {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.Status != user.StatusActive {
		return user.ErrNotFound
	}
	if u.EmailVerifiedAt != nil {
		return ErrAlreadyVerified
	}
	return s.sendVerification(ctx, u)
}

func (s *Service) sendVerification(ctx context.Context, u user.User) error {
	raw, err := newToken()
	if err != nil {
		return err
	}
	if err := s.users.SaveVerification(ctx, u.ID, hashToken(raw), time.Now().Add(verificationTTL)); err != nil {
		return err
	}
	link := s.mail.AppURL("/verifikasi-email?token=" + url.QueryEscape(raw))
	if !s.mail.Enabled() {
		// Tanpa SMTP (mis. pengembangan lokal) tautan ditulis ke log supaya
		// alur verifikasi tetap bisa dicoba.
		slog.WarnContext(ctx, "SMTP mati: tautan verifikasi email", "email", u.Email, "link", link)
		return nil
	}
	s.mail.NotifyEmail(u.Email, "Verifikasi email akun Hii Stranger", mailer.Email{
		Heading: "Verifikasi email kamu",
		Paragraphs: []string{
			"Halo " + u.Name + ", terima kasih sudah mendaftar di Hii Stranger.",
			"Klik tombol di bawah untuk memverifikasi email. Setelah terverifikasi kamu bisa langsung checkout pesanan.",
		},
		Button: &mailer.Link{Label: "Verifikasi email", URL: link},
		Note:   "Tautan berlaku 24 jam dan hanya bisa dipakai sekali. Abaikan email ini bila kamu tidak merasa mendaftar.",
	})
	return nil
}

func (s *Service) issue(u user.User) (Result, error) {
	tok, err := s.tokens.Issue(u.ID, u.Role)
	if err != nil {
		return Result{}, fmt.Errorf("auth: issue token: %w", err)
	}
	return Result{
		AccessToken: tok,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.tokens.TTL().Seconds()),
		User:        u.Response(),
	}, nil
}

// newToken membuat token acak 32 byte (URL-safe). Yang disimpan di database
// hanya hash SHA-256-nya, jadi kebocoran database tidak membocorkan tautan.
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RequestPasswordReset mengirim tautan reset ke email bila akun aktif.
// Email yang tidak terdaftar/nonaktif sengaja tidak dibedakan (selalu sukses)
// supaya endpoint tidak bisa dipakai menebak email yang terdaftar.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	u, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if u.Status != user.StatusActive {
		return nil
	}
	raw, err := newToken()
	if err != nil {
		return err
	}
	if err := s.users.SavePasswordReset(ctx, u.ID, hashToken(raw), time.Now().Add(passwordResetTTL)); err != nil {
		return err
	}
	link := s.mail.AppURL("/atur-ulang-sandi?token=" + url.QueryEscape(raw))
	if !s.mail.Enabled() {
		slog.WarnContext(ctx, "SMTP mati: tautan reset password", "email", u.Email, "link", link)
		return nil
	}
	s.mail.NotifyEmail(u.Email, "Atur ulang password akun Hii Stranger", mailer.Email{
		Heading: "Atur ulang password",
		Paragraphs: []string{
			"Halo " + u.Name + ", kami menerima permintaan untuk mengatur ulang password akun Hii Stranger kamu.",
			"Klik tombol di bawah untuk membuat password baru.",
		},
		Button: &mailer.Link{Label: "Buat password baru", URL: link},
		Note:   "Tautan berlaku 1 jam dan hanya bisa dipakai sekali. Abaikan email ini bila kamu tidak meminta reset password; password lama tetap berlaku.",
	})
	return nil
}

// ResetPassword mengganti password memakai token dari tautan email.
func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	hash, err := password.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("auth: hash password: %w", err)
	}
	_, err = s.users.ConsumePasswordReset(ctx, hashToken(rawToken), hash)
	return err
}
