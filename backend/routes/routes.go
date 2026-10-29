package routes

import (
	"net/http"

	"hii-stranger1/backend/config"
	"hii-stranger1/backend/handlers"
	"hii-stranger1/backend/middleware"
)

// Alur satu request:
//   Logger -> Recover -> [RateLimit atau Auth, dipasang per route] -> handlers -> services -> repository -> db
// File ini merakit mux dan middleware global. Daftar route per fitur ada di file masing-masing
// (auth: routes/auth.go, katalog: routes/catalog.go).

// New merakit seluruh route beserta middleware global.
func New(cfg config.Config) (http.Handler, error) {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/v1/health", handlers.Health)

	// Authentication
	if err := registerAuthRoutes(mux, cfg); err != nil {
		return nil, err
	}

	// Katalog (publik): kategori, produk, banner, gambar unggahan
	if err := registerCatalogRoutes(mux); err != nil {
		return nil, err
	}

	return middleware.Chain(
		mux,
		middleware.Logger,
		middleware.Recover,
	), nil
}
