// Package password membungkus bcrypt untuk hashing dan verifikasi password.
package password

import (
	"log/slog"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

const cost = 12

// Hash menghasilkan hash bcrypt. Password maksimal 72 byte; validasi panjang
// dilakukan oleh pemanggil sebelum sampai di sini.
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify mengembalikan true bila plain cocok dengan hash.
func Verify(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

var dummyHash = sync.OnceValues(func() (string, error) {
	return Hash("dummy-password-for-timing")
})

// VerifyDummy menjalankan perbandingan bcrypt terhadap hash palsu. Dipanggil
// saat akun tidak ditemukan supaya waktu respons login sama dengan akun yang
// ada, sehingga keberadaan email tidak bisa ditebak dari lamanya respons.
func VerifyDummy(plain string) {
	hash, err := dummyHash()
	if err != nil {
		slog.Warn("password: dummy hash unavailable, login timing not equalized", "error", err)
		return
	}
	Verify(hash, plain)
}
