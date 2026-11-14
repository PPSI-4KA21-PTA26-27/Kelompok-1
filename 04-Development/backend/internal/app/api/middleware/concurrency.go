package middleware

import (
	"net/http"

	"hii-stranger1/backend/internal/pkg/response"
)

// MaxConcurrent membatasi jumlah request API yang diproses bersamaan. Bila
// penuh, request langsung dijawab 503 + Retry-After (tanpa menunggu) supaya
// lonjakan tidak menumpuk goroutine dan koneksi database. max <= 0 mematikan
// batas. Path di exempt (dicocokkan persis, mis. health check) tidak dihitung;
// /metrics sudah berada di luar rantai API.
func MaxConcurrent(max int, exempt ...string) Middleware {
	return func(next http.Handler) http.Handler {
		if max <= 0 {
			return next
		}
		sem := make(chan struct{}, max)
		skip := make(map[string]struct{}, len(exempt))
		for _, p := range exempt {
			skip[p] = struct{}{}
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				next.ServeHTTP(w, r)
			default:
				w.Header().Set("Retry-After", "1")
				response.Error(w, http.StatusServiceUnavailable, "server_busy", "Server sedang sibuk. Coba lagi sebentar lagi.")
			}
		})
	}
}
