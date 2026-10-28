package routes

import (
	"net/http"

	"hii-stranger1/backend/handlers"
)

func registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/Login", handlers.Login)
	mux.HandleFunc("POST /api/v1/Register", handlers.Register)
}
