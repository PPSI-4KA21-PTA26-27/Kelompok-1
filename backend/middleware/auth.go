package middleware

import (
	"context"
	"net/http"
	"strings"

	"hii-stranger1/backend/models"
	"hii-stranger1/backend/response"
	"hii-stranger1/backend/services"
)

// Alur: routes -> middleware.Auth (secret JWT dari config) -> handler.
// Auth memeriksa token lalu menyimpan id + role user di context request.
// Handler membaca id itu lewat UserID(r.Context()), BUKAN dari isi request dari klien.

type ctxKey struct{}

// UserID mengambil id user yang sudah diverifikasi oleh Auth.
func UserID(ctx context.Context) (int64, bool) {
	c, ok := ctx.Value(ctxKey{}).(*services.TokenClaims)
	if !ok {
		return 0, false
	}
	return c.UserID, true
}

// Auth menolak (401) request tanpa token Bearer yang valid dan belum kedaluwarsa.
func Auth(secret []byte) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				response.Error(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
				return
			}
			claims, err := services.ParseToken(secret, token)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "sesi tidak valid atau sudah berakhir")
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin hanya meloloskan role admin (403 untuk yang lain).
// Harus dipasang SETELAH Auth, karena membaca hasil Auth dari context.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, ok := r.Context().Value(ctxKey{}).(*services.TokenClaims)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
			return
		}
		if c.Role != models.RoleAdmin {
			response.Error(w, http.StatusForbidden, "akses ditolak")
			return
		}
		next.ServeHTTP(w, r)
	})
}
