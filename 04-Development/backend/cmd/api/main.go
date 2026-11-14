// Command api menjalankan server HTTP backend.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"hii-stranger1/backend/internal/app/api"
	"hii-stranger1/backend/internal/pkg/config"
	"hii-stranger1/backend/internal/pkg/database"
	redisclient "hii-stranger1/backend/internal/pkg/redis"
)

const (
	MyAppName = "hii-stranger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	slog.SetDefault(newLogger(cfg))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer pool.Close()

	redisClient, err := redisclient.New(
		ctx,
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	if err != nil {
		return fmt.Errorf("initialize redis: %w", err)
	}
	defer func() { _ = redisClient.Close() }()
	slog.Info(
		"server starting",
		"app", MyAppName,
		"version", cfg.AppVersion,
		"addr", ":"+cfg.Port,
		"env", cfg.AppEnv,
	)

	if err := api.New(cfg, pool, redisClient).Run(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}

func newLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}

	if cfg.AppEnv == "dev" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
