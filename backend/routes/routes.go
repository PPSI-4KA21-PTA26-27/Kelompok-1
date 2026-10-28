package routes

import (
	"net/http"

	"hii-stranger1/backend/handlers"
	"hii-stranger1/backend/middleware"
)

// New merakit seluruh route beserta middleware global.
func New() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/v1/health", handlers.Health)

	// Authentication
	registerAuthRoutes(mux)

	return middleware.Chain(
		mux,
		middleware.Logger,
		middleware.Recover,
	)
}
