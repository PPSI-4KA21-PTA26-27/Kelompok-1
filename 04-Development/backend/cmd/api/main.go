// Command api adalah entry point server HTTP Hii Stranger.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"kel1/backend/internal/config"
	"kel1/backend/internal/database"
	"kel1/backend/internal/httpapi"
	"kel1/backend/internal/platform/logging"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logging.New(cfg.LogLevel, cfg.IsProduction(), os.Stdout)
	slog.SetDefault(log)
	for _, w := range cfg.Warnings {
		log.Warn("config_warning", slog.String("detail", w))
	}
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Database, "hii-stranger-api")
	if err != nil {
		return err
	}
	defer pool.Close() // dijalankan setelah server berhenti menerima request

	handler, err := httpapi.NewRouter(httpapi.Dependencies{Config: cfg, Logger: log, DB: pool})
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Info("server_starting", slog.String("addr", srv.Addr), slog.String("env", cfg.Env))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
	}

	log.Info("server_shutting_down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown gagal: %w", err)
	}
	log.Info("server_stopped")
	return nil
}
