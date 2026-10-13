package middleware

import (
	"io"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"kel1/backend/internal/httpapi/response"
)

// AccessLog mencatat satu baris terstruktur per request.
//
// Yang dicatat: method, pola route (bukan URL mentah), status, latency, ukuran, IP, request ID.
// Yang TIDAK dicatat: query string, header, dan body, karena bisa memuat token atau password.
func AccessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := c.Writer.Status()

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		case route == "/healthz" || route == "/readyz":
			level = slog.LevelDebug // probe berkala jangan memenuhi log
		}

		log.LogAttrs(c.Request.Context(), level, "http_request",
			slog.String("request_id", GetRequestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("route", route),
			slog.Int("status", status),
			slog.Float64("latency_ms", float64(time.Since(start).Microseconds())/1000),
			slog.Int("bytes", c.Writer.Size()),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

// Recovery menangkap panic, mencatat detail di log, dan hanya mengirim pesan generik ke klien.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		log.LogAttrs(c.Request.Context(), slog.LevelError, "panic_recovered",
			slog.String("request_id", GetRequestID(c)),
			slog.Any("panic", recovered),
			slog.String("stack", string(debug.Stack())),
		)
		response.Abort(c, 500, response.CodeInternal, "Internal server error")
	})
}
