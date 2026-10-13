package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"kel1/backend/internal/httpapi/response"
)

// SecurityHeaders menambahkan header pengaman standar untuk API JSON.
// HSTS hanya dikirim di production (di belakang TLS terminator seperti Railway).
func SecurityHeaders(production bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Cross-Origin-Resource-Policy", "same-site")
		h.Set("Cache-Control", "no-store")
		if production {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
		c.Next()
	}
}

// CORS memakai allowlist origin eksplisit. Wildcard ditolak karena AllowCredentials aktif
// (refresh token lewat cookie HttpOnly di Fase 2).
func CORS(origins []string) (gin.HandlerFunc, error) {
	cfg := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", RequestIDHeader},
		ExposeHeaders:    []string{RequestIDHeader},
		AllowCredentials: true,
		MaxAge:           10 * time.Minute,
	}
	for _, o := range origins {
		if o == "*" {
			return nil, errWildcardOrigin
		}
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cors.New(cfg), nil
}

type middlewareError string

func (e middlewareError) Error() string { return string(e) }

const errWildcardOrigin = middlewareError("CORS: wildcard origin tidak diizinkan bersama credentials")

// BodyLimit membatasi ukuran body. Request dengan Content-Length berlebih ditolak langsung (413);
// untuk body tanpa Content-Length, pembacaan akan gagal saat melewati batas.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			response.Abort(c, http.StatusRequestEntityTooLarge, response.CodePayloadTooLarge, "Request body too large")
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// RequestTimeout memberi deadline pada context request supaya query database dan
// panggilan eksternal ikut dibatalkan. Handler wajib meneruskan c.Request.Context().
func RequestTimeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
