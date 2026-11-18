// Package user mengelola data pengguna: model, akses data, dan endpoint profil.
package user

import (
	"encoding/json"
	"errors"
	"time"
)

// Nilai role dan status yang dipakai aplikasi. Kolom di database bertipe TEXT
// tanpa constraint, jadi nilai-nilai ini hanya dijaga oleh aplikasi.
const (
	RoleCustomer = "customer"
	RoleOwner    = "owner"
	// RoleStaff mengelola operasional toko (katalog, pesanan, pembayaran).
	// Seperti admin, staff tidak bisa belanja.
	RoleStaff      = "staff"
	StatusActive   = "active"
	StatusInactive = "inactive"
)

var (
	// ErrNotFound berarti user tidak ada.
	ErrNotFound = errors.New("user: not found")
	// ErrEmailTaken berarti email sudah dipakai akun lain.
	ErrEmailTaken = errors.New("user: email already registered")
	// ErrWrongPassword berarti password saat ini yang dikirim tidak cocok.
	ErrWrongPassword = errors.New("user: wrong password")
	// ErrSelfLockout berarti admin mencoba menurunkan role atau menonaktifkan
	// akunnya sendiri.
	ErrSelfLockout = errors.New("user: admin cannot demote or deactivate self")
	// ErrInvalidVerification: token verifikasi email tidak dikenal atau kedaluwarsa.
	ErrInvalidVerification = errors.New("user: invalid or expired verification token")
	// ErrInvalidPasswordReset: token reset password tidak dikenal atau kedaluwarsa.
	ErrInvalidPasswordReset = errors.New("user: invalid or expired password reset token")
)

// OptionalString membedakan field JSON yang tidak dikirim (Set false), yang
// dikirim null (Set true, Value nil), dan yang dikirim berisi teks.
type OptionalString struct {
	Set   bool
	Value *string
}

// UnmarshalJSON menandai field sebagai dikirim. Method ini tidak dipanggil
// bila field tidak ada di JSON, sehingga Set tetap false.
func (o *OptionalString) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	o.Value = &s
	return nil
}

// ProfileChanges adalah perubahan profil yang sudah divalidasi. Name nil berarti
// nama tidak diubah. Phone.Set true dengan Value nil berarti nomor dihapus.
type ProfileChanges struct {
	Name  *string
	Phone OptionalString
}

// User mencerminkan satu baris tabel users. Jangan kirim langsung ke client:
// gunakan Response agar PasswordHash tidak pernah bocor.
type User struct {
	ID              string
	Name            string
	Email           string
	PasswordHash    string
	Phone           *string
	Role            string
	Status          string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
}

// Response adalah bentuk user di API (schema User pada openapi.yaml).
type Response struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Phone           *string    `json:"phone"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Response mengubah User menjadi bentuk yang aman dikirim ke client.
func (u User) Response() Response {
	return Response{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Phone:           u.Phone,
		Role:            u.Role,
		Status:          u.Status,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
	}
}

// IsBackoffice true untuk akun staf toko (admin atau staff).
func IsBackoffice(role string) bool {
	return role == RoleOwner || role == RoleStaff
}
