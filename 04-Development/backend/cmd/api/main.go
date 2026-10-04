package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kel1/backend/internal/config"
	"kel1/backend/internal/database"
	"kel1/backend/internal/models"
	"kel1/backend/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}

	if err := db.AutoMigrate(models.All()...); err != nil {
		log.Fatalf("gagal migrate: %v", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router.Setup(cfg, db),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		log.Printf("server jalan di http://localhost:%s (env: %s)", cfg.AppPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown paksa: %v", err)
	}
	log.Println("server berhenti")
}
