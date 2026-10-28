package models

import "time"

// Alur: models dipakai oleh repository (isi dari database), services (aturan bisnis),
// dan handlers (bentuk JSON). Paket ini tidak memanggil paket lain.

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// User mewakili satu baris tabel users.
type User struct {
	ID        int64     `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`

	// json:"-" = tidak pernah ikut terkirim ke klien.
	PasswordHash string `json:"-"`
}

// RegisterRequest: isi form "Daftar Akun".
type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// LoginRequest: isi form "Masuk".
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ForgotPasswordRequest: isi form "Lupa Kata Sandi".
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest: kata sandi baru + token dari tautan email.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// UpdateProfileRequest: ubah nama dan nomor HP di halaman akun.
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// ChangePasswordRequest: ganti kata sandi saat sudah masuk.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// AuthResult: jawaban sukses untuk daftar dan masuk.
type AuthResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // detik
	User        User   `json:"user"`
}
