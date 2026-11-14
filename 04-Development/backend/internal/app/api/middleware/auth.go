package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/token"
)

// Principal adalah identitas pemanggil yang diambil dari token yang valid.
// Nilainya berasal dari token, bukan dari body atau header buatan client.
type Principal struct {
	UserID string
	Role   string
}

type principalKey struct{}

// PrincipalFrom mengambil Principal yang dipasang oleh Authenticate.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}

// Authenticate menolak request tanpa Bearer token yang valid dengan 401, dan
// memasang Principal ke context bila valid.
func Authenticate(tokens *token.Manager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authed, ok := withPrincipal(tokens, r)
			if !ok {
				unauthorized(w)
				return
			}
			next.ServeHTTP(w, authed)
		})
	}
}

// AuthenticateOptional dipakai untuk rute publik yang isinya bisa berbeda bagi
// admin. Request tanpa header Authorization lolos sebagai anonim (tanpa
// Principal). Bila header dikirim, token harus valid; token rusak atau
// kedaluwarsa dijawab 401 dan tidak diam-diam diperlakukan sebagai anonim.
func AuthenticateOptional(tokens *token.Manager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				next.ServeHTTP(w, r)
				return
			}
			authed, ok := withPrincipal(tokens, r)
			if !ok {
				unauthorized(w)
				return
			}
			next.ServeHTTP(w, authed)
		})
	}
}

// withPrincipal memverifikasi Bearer token dan mengembalikan request yang
// membawa Principal. ok false bila token tidak ada atau tidak valid.
func withPrincipal(tokens *token.Manager, r *http.Request) (*http.Request, bool) {
	raw, ok := bearerToken(r)
	if !ok {
		return nil, false
	}
	claims, err := tokens.Parse(raw)
	if err != nil {
		return nil, false
	}
	ctx := context.WithValue(r.Context(), principalKey{}, Principal{UserID: claims.Subject, Role: claims.Role})
	return r.WithContext(ctx), true
}

// RequireRole menolak pemanggil yang rolenya bukan role dengan 403. Wajib
// dipasang setelah Authenticate; tanpa Principal di context dijawab 401. Role
// dibaca dari token, jadi perubahan role baru berlaku pada token berikutnya.
func RequireRole(roles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFrom(r.Context())
			if !ok {
				unauthorized(w)
				return
			}
			if !slices.Contains(roles, p.Role) {
				response.Forbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	scheme, raw, found := strings.Cut(r.Header.Get("Authorization"), " ")
	raw = strings.TrimSpace(raw)
	if !found || !strings.EqualFold(scheme, "Bearer") || raw == "" {
		return "", false
	}
	return raw, true
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	response.Error(w, http.StatusUnauthorized, "unauthorized", "Autentikasi diperlukan.")
}
