// Package database membuat koneksi pool ke PostgreSQL.
package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"hii-stranger1/backend/internal/pkg/config"
)

const pingTimeout = 5 * time.Second

// NewPool membuka pool koneksi, menerapkan statement_timeout, dan memastikan
// database dapat dijangkau sebelum aplikasi menerima request.
func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("database: parse DB_URL: %w", err)
	}
	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MinConns = cfg.DBMinConns
	poolCfg.ConnConfig.RuntimeParams["statement_timeout"] = strconv.FormatInt(cfg.DBStatementTimeout.Milliseconds(), 10)

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return pool, nil
}
