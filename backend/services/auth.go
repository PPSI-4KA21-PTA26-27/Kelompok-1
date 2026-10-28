package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"hii-stranger1/backend/config"
	"hii-stranger1/backend/models"
	"hii-stranger1/backend/repository"
)

// Alur: handlers (auth, akun) -> AuthService -> repository (user, reset_token) -> database.
// Pengaturan JWT dan URL frontend datang dari config. Semua aturan bisnis auth ada di file ini.

// resetTokenTTL: tautan reset berlaku 15 menit (sesuai desain lupa_kata_sandi).
const resetTokenTTL = 15 * time.Minute

var (
	// ErrInvalidCredentials sengaja tidak membedakan "email tidak ada" dan "sandi salah",
	// supaya orang tidak bisa mengecek email mana yang terdaftar.
	ErrInvalidCredentials = errors.New("email atau kata sandi salah")
	// ErrInvalidResetToken: token reset tidak ada, sudah dipakai, atau sudah kedaluwarsa.
	ErrInvalidResetToken = errors.New("tautan reset tidak valid atau sudah kedaluwarsa")
)

type AuthService struct {
	users  *repository.UserRepository
	resets *repository.ResetTokenRepository
	cfg    config.Config
	sender ResetSender // nil kalau SMTP belum diatur

	// dummyHash dipakai saat email tidak ditemukan, supaya waktu respons login
	// tetap sama dan tidak membocorkan email mana yang terdaftar.
	dummyHash []byte
}

func NewAuthService(users *repository.UserRepository, resets *repository.ResetTokenRepository, cfg config.Config, sender ResetSender) (*AuthService, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("dummy-password-untuk-timing"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("siapkan hash cadangan: %w", err)
	}
	return &AuthService{users: users, resets: resets, cfg: cfg, sender: sender, dummyHash: dummy}, nil
}

// Register membuat akun customer baru lalu langsung memberi token masuk.
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthResult, error) {
	fullName := strings.TrimSpace(req.FullName)
	email := normalizeEmail(req.Email)
	phone := normalizePhone(req.Phone)

	if err := collect(map[string]string{
		"full_name": validateFullName(fullName),
		"email":     validateEmail(email),
		"phone":     validatePhone(phone),
		"password":  validatePassword(req.Password, email),
	}); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Role selalu customer di sini. Admin tidak bisa dibuat lewat pendaftaran publik.
	u := &models.User{
		FullName:     fullName,
		Email:        email,
		Phone:        phone,
		Role:         models.RoleCustomer,
		PasswordHash: string(hash),
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err // termasuk repository.ErrEmailTaken
	}
	return s.authResult(u)
}

// Login memeriksa email + kata sandi, lalu memberi token.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResult, error) {
	email := normalizeEmail(req.Email)

	passwordMsg := ""
	if req.Password == "" {
		passwordMsg = "Kata sandi wajib diisi"
	}
	if err := collect(map[string]string{
		"email":    validateEmail(email),
		"password": passwordMsg,
	}); err != nil {
		return nil, err
	}

	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(req.Password)) // hanya untuk menyamakan waktu
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return s.authResult(u)
}

// ForgotPassword membuat token reset lalu mengirim tautannya lewat email.
// Hasilnya sama (nil) baik email terdaftar maupun tidak, supaya orang tidak bisa menebak
// email mana yang punya akun. Kegagalan kirim email hanya dicatat di log server.
func (s *AuthService) ForgotPassword(ctx context.Context, req models.ForgotPasswordRequest) error {
	email := normalizeEmail(req.Email)
	if err := collect(map[string]string{"email": validateEmail(email)}); err != nil {
		return err
	}

	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	if s.sender == nil {
		log.Printf("lupa sandi: SMTP belum diatur, email reset untuk user %d tidak dikirim", u.ID)
		return nil
	}

	token, err := newResetToken()
	if err != nil {
		return err
	}
	if err := s.resets.Create(ctx, u.ID, hashToken(token), time.Now().Add(resetTokenTTL)); err != nil {
		return err
	}

	link := strings.TrimRight(s.cfg.FrontendURL, "/") + "/reset-password?token=" + token
	if err := s.sender.SendResetLink(ctx, u.Email, link); err != nil {
		log.Printf("lupa sandi: gagal kirim email reset untuk user %d: %v", u.ID, err)
	}
	return nil
}

// ResetPassword memakai token dari email untuk menetapkan kata sandi baru.
func (s *AuthService) ResetPassword(ctx context.Context, req models.ResetPasswordRequest) error {
	token := strings.TrimSpace(req.Token)

	tokenMsg := ""
	if token == "" {
		tokenMsg = "Token wajib diisi"
	}
	if err := collect(map[string]string{
		"token":        tokenMsg,
		"new_password": validatePassword(req.NewPassword, ""),
	}); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	err = s.resets.ResetPassword(ctx, hashToken(token), string(hash), time.Now())
	if errors.Is(err, repository.ErrNotFound) {
		return ErrInvalidResetToken
	}
	return err
}

// GetProfile mengambil data akun milik user yang sedang masuk.
func (s *AuthService) GetProfile(ctx context.Context, userID int64) (*models.User, error) {
	return s.users.FindByID(ctx, userID)
}

// UpdateProfile mengubah nama dan nomor HP. Email dan role tidak bisa diubah dari sini.
func (s *AuthService) UpdateProfile(ctx context.Context, userID int64, req models.UpdateProfileRequest) (*models.User, error) {
	fullName := strings.TrimSpace(req.FullName)
	phone := normalizePhone(req.Phone)

	if err := collect(map[string]string{
		"full_name": validateFullName(fullName),
		"phone":     validatePhone(phone),
	}); err != nil {
		return nil, err
	}

	if err := s.users.UpdateProfile(ctx, userID, fullName, phone); err != nil {
		return nil, err
	}
	return s.users.FindByID(ctx, userID)
}

// ChangePassword mengganti kata sandi setelah memastikan kata sandi lama benar.
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, req models.ChangePasswordRequest) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Salah di sini dikirim sebagai error field (422), bukan 401, supaya frontend
	// tidak mengira sesi habis lalu mengeluarkan pengguna.
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.CurrentPassword)) != nil {
		return &ValidationError{Fields: map[string]string{"current_password": "Kata sandi saat ini salah"}}
	}
	if err := collect(map[string]string{"new_password": validatePassword(req.NewPassword, u.Email)}); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.users.UpdatePassword(ctx, userID, string(hash))
}

func (s *AuthService) authResult(u *models.User) (*models.AuthResult, error) {
	token, err := IssueToken(s.cfg.JWTSecret, s.cfg.JWTTTL, u, time.Now())
	if err != nil {
		return nil, err
	}
	return &models.AuthResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.cfg.JWTTTL.Seconds()),
		User:        *u,
	}, nil
}

// newResetToken membuat token acak 256-bit (hex) dari generator acak yang aman untuk kriptografi.
func newResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("buat token reset: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// hashToken: hanya hash ini yang disimpan di database.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
