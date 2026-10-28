package services

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Alur: dipanggil oleh AuthService sebelum menyentuh repository.
// Pesan error ditulis untuk ditampilkan langsung di bawah input pada form.

const (
	maxFullNameLen = 100 // sama dengan batas 0/100 di form daftar
	maxEmailLen    = 255 // sama dengan kolom users.email
	minPasswordLen = 10  // desain: "Minimal 10 karakter"
	maxPasswordLen = 72  // batas bcrypt; lebih dari ini akan dipotong diam-diam
)

// Nomor HP: boleh diawali +, hanya angka, 8-15 digit.
var phonePattern = regexp.MustCompile(`^\+?[0-9]{8,15}$`)

// ValidationError membawa pesan salah per field, contoh {"email": "Format email tidak valid"}.
// Handler mengubahnya menjadi respons 422.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "data tidak valid" }

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// normalizePhone membuang spasi dan tanda minus supaya "0812 3456-7890" tersimpan seragam.
func normalizePhone(phone string) string {
	return strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(phone))
}

// Fungsi validate* mengembalikan "" kalau valid, atau pesan salah untuk field itu.

func validateFullName(name string) string {
	n := utf8.RuneCountInString(name)
	if n == 0 {
		return "Nama lengkap wajib diisi"
	}
	if n > maxFullNameLen {
		return "Nama lengkap maksimal 100 karakter"
	}
	return ""
}

func validateEmail(email string) string {
	if email == "" {
		return "Email wajib diisi"
	}
	if len(email) > maxEmailLen {
		return "Email terlalu panjang"
	}
	// Address harus sama persis dengan input; ini menolak bentuk "Nama <a@b.com>".
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "Format email tidak valid"
	}
	return ""
}

// validatePhone: nomor HP boleh kosong (opsional di desain).
func validatePhone(phone string) string {
	if phone != "" && !phonePattern.MatchString(phone) {
		return "Nomor telepon tidak valid"
	}
	return ""
}

// validatePassword menerapkan syarat di desain: min 10 karakter, ada huruf dan angka,
// tidak sama dengan email. Isi email dengan "" kalau email belum diketahui.
func validatePassword(password, email string) string {
	if utf8.RuneCountInString(password) < minPasswordLen {
		return "Minimal 10 karakter"
	}
	if len(password) > maxPasswordLen {
		return "Maksimal 72 karakter"
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		hasLetter = hasLetter || unicode.IsLetter(r)
		hasDigit = hasDigit || unicode.IsDigit(r)
	}
	if !hasLetter || !hasDigit {
		return "Harus mengandung kombinasi huruf dan angka"
	}
	if email != "" && strings.EqualFold(password, email) {
		return "Tidak boleh sama dengan alamat email"
	}
	return ""
}

// collect menggabungkan hasil validasi; field yang valid (pesan kosong) tidak ikut.
// Mengembalikan nil kalau semua field valid.
func collect(fields map[string]string) error {
	for name, msg := range fields {
		if msg == "" {
			delete(fields, name)
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: fields}
}
