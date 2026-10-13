// Package httpapi merakit router HTTP: middleware global, route, dan handler infrastruktur.
// Handler fitur (auth, produk, order, dst.) hidup di modul masing-masing dan didaftarkan di sini.
package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"kel1/backend/internal/config"
	"kel1/backend/internal/httpapi/middleware"
	"kel1/backend/internal/httpapi/response"
)

type Dependencies struct {
	Config *config.Config
	Logger *slog.Logger
	DB     Pinger
}

// NewRouter membangun gin.Engine lengkap dengan middleware keamanan dan observability.
// Pemanggil bertanggung jawab memanggil gin.SetMode sebelum ini.
func NewRouter(d Dependencies) (*gin.Engine, error) {
	cfg := d.Config

	corsMW, err := middleware.CORS(cfg.HTTP.CORSAllowedOrigins)
	if err != nil {
		return nil, fmt.Errorf("setup CORS: %w", err)
	}

	engine := gin.New()
	engine.HandleMethodNotAllowed = true
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false
	// Default: tidak percaya proxy mana pun, sehingga X-Forwarded-For palsu tidak bisa
	// memalsukan IP klien (penting untuk rate limiting di Fase 2).
	if err := engine.SetTrustedProxies(cfg.HTTP.TrustedProxies); err != nil {
		return nil, fmt.Errorf("setup trusted proxies: %w", err)
	}

	// Urutan penting: RequestID dulu agar semua log punya ID; AccessLog membungkus Recovery
	// supaya panic yang dipulihkan tetap tercatat sebagai status 500.
	engine.Use(
		middleware.RequestID(),
		middleware.AccessLog(d.Logger),
		middleware.Recovery(d.Logger),
		middleware.SecurityHeaders(cfg.IsProduction()),
		corsMW,
		middleware.BodyLimit(cfg.HTTP.MaxBodyBytes),
		middleware.RequestTimeout(cfg.HTTP.RequestTimeout),
	)

	engine.NoRoute(func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "Route not found", nil)
	})
	engine.NoMethod(func(c *gin.Context) {
		response.Error(c, http.StatusMethodNotAllowed, response.CodeMethodNotAllowed, "Method not allowed", nil)
	})

	registerRoutes(engine, d)
	return engine, nil
}

func registerRoutes(engine *gin.Engine, d Dependencies) {
	engine.GET("/healthz", healthz)
	engine.GET("/readyz", readyz(d.DB, d.Logger))

	// Grup /api/v1 ditambahkan per fase:
	//   Fase 2: /auth, /me, /admin/auth
	//   Fase 3: /categories, /products, /admin/categories, /admin/products
	//   Fase 4: /cart, /orders, /admin/orders
	//   Fase 5: /payments, /webhooks/midtrans
}
