// Package auth menangani registrasi, login, verifikasi email, dan lupa password.
package auth

import (
	"context"
	"errors"
	"fmt"

	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/mailer"
	"hii-stranger1/backend/internal/pkg/password"
	"hii-stranger1/backend/internal/pkg/token"
)

// ErrInvalidCredentials dikembalikan untuk email/password salah atau akun nonaktif.
var ErrInvalidCredentials = errors.New("auth: invalid credentials")

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

// Register membuat akun pelanggan aktif lalu langsung menerbitkan token login.
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
