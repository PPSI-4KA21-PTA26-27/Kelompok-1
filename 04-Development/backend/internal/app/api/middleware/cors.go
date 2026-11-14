package middleware

import "net/http"

// CORS mengizinkan satu origin frontend saja. Bila allowedOrigin kosong, CORS
// dimatikan dan tidak ada header yang ditambahkan. Autentikasi memakai header
// Authorization (bukan cookie), jadi credentials tidak diaktifkan.
func CORS(allowedOrigin string) Middleware {
	return func(next http.Handler) http.Handler {
		if allowedOrigin == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && origin == allowedOrigin {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")

				isPreflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
				if isPreflight {
					h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
					h.Set("Access-Control-Max-Age", "600")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
