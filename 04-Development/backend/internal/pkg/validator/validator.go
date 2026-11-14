// Package validator berisi pemeriksaan input yang dipakai berulang oleh handler.
package validator

import (
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"hii-stranger1/backend/internal/pkg/response"
)

const (
	maxEmailLen = 254
	maxPhoneLen = 20
	maxURLLen   = 2048
	maxMoneyLen = 20
)

// moneyPattern sama dengan schema Money pada openapi.yaml.
var moneyPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// Collector mengumpulkan kesalahan validasi per field.
type Collector struct {
	errs []response.FieldError
}

// Check mencatat kesalahan untuk field bila ok bernilai false.
func (c *Collector) Check(ok bool, field, message string) {
	if !ok {
		c.errs = append(c.errs, response.FieldError{Field: field, Message: message})
	}
}

// Errors mengembalikan kesalahan yang terkumpul (kosong bila semua valid).
func (c *Collector) Errors() []response.FieldError {
	return c.errs
}

// RuneLenBetween memeriksa jumlah karakter (rune) s berada di antara lo dan hi.
func RuneLenBetween(s string, lo, hi int) bool {
	n := utf8.RuneCountInString(s)
	return n >= lo && n <= hi
}

// ByteLenBetween memeriksa jumlah byte s berada di antara lo dan hi.
func ByteLenBetween(s string, lo, hi int) bool {
	return len(s) >= lo && len(s) <= hi
}

// Email memeriksa bentuk alamat email polos (tanpa nama tampilan) dengan domain
// yang memuat titik. Pemanggil bertanggung jawab melakukan trim dan lowercase.
func Email(s string) bool {
	if s == "" || len(s) > maxEmailLen || strings.ContainsAny(s, " \t\r\n") {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return false
	}
	at := strings.LastIndex(s, "@")
	return strings.Contains(s[at+1:], ".")
}

// Phone memeriksa nomor telepon: maksimal 20 karakter, hanya angka, +, spasi,
// tanda hubung, dan kurung, serta memuat setidaknya satu angka.
func Phone(s string) bool {
	if s == "" || len(s) > maxPhoneLen {
		return false
	}
	for _, r := range s {
		if !strings.ContainsRune("+0123456789 -()", r) {
			return false
		}
	}
	return strings.ContainsAny(s, "0123456789")
}

// UUID memeriksa bentuk kanonik UUID: 36 karakter, digit heksadesimal dengan
// tanda hubung pada posisi 8-4-4-4-12. Dipakai sebelum ID dari path atau query
// masuk ke database, supaya nilai asal-asalan tidak berubah jadi error SQL.
func UUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		switch i {
		case 8, 13, 18, 23:
			if s[i] != '-' {
				return false
			}
		default:
			c := s[i]
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				return false
			}
		}
	}
	return true
}

// HTTPURL memeriksa s adalah URL absolut berskema http atau https yang punya
// host, maksimal 2048 byte. Dipakai untuk image_url dan link_url.
func HTTPURL(s string) bool {
	if s == "" || len(s) > maxURLLen || strings.ContainsAny(s, " \t\r\n") {
		return false
	}
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// Money memeriksa nilai uang berupa string desimal tak negatif, contoh
// "150000" atau "150000.50" (schema Money), maksimal 20 karakter.
func Money(s string) bool {
	return len(s) <= maxMoneyLen && moneyPattern.MatchString(s)
}
