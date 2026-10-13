// Package database menyediakan koneksi PostgreSQL (pgxpool) dan runner migrasi (golang-migrate).
package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"kel1/backend/internal/config"
)

// NewPool membuat connection pool dan memastikan database benar-benar terjangkau.
//
// statement_timeout diset di sisi server sebagai pengaman terakhir bila ada query yang
// lolos tanpa context deadline. Pesan error sengaja tidak menyertakan DATABASE_URL.
func NewPool(ctx context.Context, cfg config.DatabaseConfig, appName string) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, errors.New("DATABASE_URL tidak dapat diparse")
	}

	pc.MaxConns = cfg.MaxConns
	pc.MinConns = cfg.MinConns
	pc.MaxConnLifetime = cfg.MaxConnLifetime
	pc.MaxConnIdleTime = cfg.MaxConnIdleTime
	pc.HealthCheckPeriod = 30 * time.Second
	pc.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	if pc.ConnConfig.RuntimeParams == nil {
		pc.ConnConfig.RuntimeParams = map[string]string{}
	}
	pc.ConnConfig.RuntimeParams["application_name"] = appName
	pc.ConnConfig.RuntimeParams["statement_timeout"] = strconv.FormatInt(cfg.StatementTimeout.Milliseconds(), 10)

	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("membuat pool database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database tidak terjangkau: %w", err)
	}
	return pool, nil
}
