package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// MaxBody membatasi ukuran body request. Pembacaan yang melewati batas gagal
// dengan *http.MaxBytesError, yang diubah menjadi 413 oleh response.Decode.
//
// Request multipart/form-data (unggah berkas) dilewati di sini karena
// batasnya berbeda per endpoint; handler unggah memasang batasnya sendiri
// lewat fileupload.Read.
func MaxBody(n int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
				next.ServeHTTP(w, r)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout memasang batas waktu pada context request, sehingga query database
// dan panggilan lain yang memakai context ikut dibatalkan saat waktunya habis.
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
