package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"hii-stranger1/backend/internal/pkg/response"
)

// Recover menangkap panic pada handler, mencatatnya beserta stack trace di
// log server, dan membalas 500 standar tanpa detail internal.
func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				slog.ErrorContext(r.Context(), "panic recovered", "panic", rec, "stack", string(debug.Stack()))
				response.Internal(w)
			}()
			next.ServeHTTP(w, r)
		})
	}
}
