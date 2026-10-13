package httpapi

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"kel1/backend/internal/httpapi/middleware"
	"kel1/backend/internal/httpapi/response"
)

// Pinger diimplementasikan oleh *pgxpool.Pool. Dibuat interface supaya readiness mudah diuji.
type Pinger interface {
	Ping(ctx context.Context) error
}

const readinessTimeout = 2 * time.Second

// healthz: liveness. Hanya menjawab bahwa proses hidup; tidak menyentuh dependency.
func healthz(c *gin.Context) {
	response.OK(c, "OK", gin.H{"status": "ok"})
}

// readyz: readiness. Memastikan database terjangkau. Detail kegagalan hanya masuk log,
// tidak pernah dikirim ke klien (tidak membocorkan host, kredensial, atau konfigurasi).
func readyz(db Pinger, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			log.WarnContext(ctx, "readiness_check_failed",
				slog.String("request_id", middleware.GetRequestID(c)),
				slog.String("error", err.Error()),
			)
			response.Error(c, 503, response.CodeUnavailable, "Service not ready", nil)
			return
		}
		response.OK(c, "Ready", gin.H{"status": "ready"})
	}
}
