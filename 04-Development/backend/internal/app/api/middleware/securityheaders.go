package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders memasang header keamanan pada semua respons API. Header
// di-set sebelum handler berjalan, jadi handler tetap bisa menimpanya. CSP
// ketat aman karena API hanya menyajikan JSON/file, tidak ada HTML. HSTS hanya
// dikirim untuk request HTTPS: r.TLS, atau X-Forwarded-Proto: https bila
// trustProxy true (TRUST_PROXY_HEADERS), supaya dev lokal http tidak terkena.
func SecurityHeaders(trustProxy bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			if isHTTPS(r, trustProxy) {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isHTTPS(r *http.Request, trustProxy bool) bool {
	if r.TLS != nil {
		return true
	}
	if !trustProxy {
		return false
	}
	proto, _, _ := strings.Cut(r.Header.Get("X-Forwarded-Proto"), ",")
	return strings.EqualFold(strings.TrimSpace(proto), "https")
}
