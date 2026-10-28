package middleware

import (
	"log"
	"net/http"

	"hii-stranger1/backend/response"
)

type Middleware func(http.Handler) http.Handler

// Chain membungkus handler; middleware pertama menjadi lapisan terluar.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Recover mengubah panic menjadi respons 500 tanpa membocorkan detail ke klien.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				response.Error(w, http.StatusInternalServerError, "terjadi kesalahan pada server")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
