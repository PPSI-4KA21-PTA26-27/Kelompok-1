package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"hii-stranger1/backend/db"
	"hii-stranger1/backend/response"
)

// Health memastikan server hidup dan database dapat dijangkau.
func Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		log.Printf("health: ping database gagal: %v", err)
		response.Error(w, http.StatusServiceUnavailable, "database tidak tersedia")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
