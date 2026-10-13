package database_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"kel1/backend/internal/config"
	"kel1/backend/internal/database"
)

// Test integrasi: butuh PostgreSQL sungguhan. Set TEST_DATABASE_URL ke database KOSONG khusus test
// (test ini menjalankan down sampai bersih). Contoh: make test-integration.
func testURL(t *testing.T) string {
	t.Helper()
	u := os.Getenv("TEST_DATABASE_URL")
	if u == "" {
		t.Skip("TEST_DATABASE_URL tidak di-set; integration test dilewati")
	}
	return u
}

func testDBConfig(url string) config.DatabaseConfig {
	return config.DatabaseConfig{
		URL:              url,
		MaxConns:         4,
		MinConns:         0,
		MaxConnLifetime:  time.Minute,
		MaxConnIdleTime:  time.Minute,
		ConnectTimeout:   5 * time.Second,
		StatementTimeout: 2 * time.Second,
	}
}

func TestNewPoolConnectsAndAppliesSessionSettings(t *testing.T) {
	url := testURL(t)
	ctx := context.Background()

	pool, err := database.NewPool(ctx, testDBConfig(url), "hii-test")
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()

	var app, timeout string
	if err := pool.QueryRow(ctx, "SELECT current_setting('application_name'), current_setting('statement_timeout')").Scan(&app, &timeout); err != nil {
		t.Fatalf("query: %v", err)
	}
	if app != "hii-test" {
		t.Errorf("application_name = %q", app)
	}
	if timeout != "2s" {
		t.Errorf("statement_timeout = %q, want 2s", timeout)
	}

	// statement_timeout benar-benar memutus query yang berjalan terlalu lama.
	if _, err := pool.Exec(ctx, "SELECT pg_sleep(5)"); err == nil {
		t.Error("query lambat seharusnya dibatalkan oleh statement_timeout")
	}
}

func TestNewPoolFailsFastAndDoesNotLeakPassword(t *testing.T) {
	cfg := testDBConfig("postgres://app:PasswordRahasia99@127.0.0.1:1/nodb?sslmode=disable")
	cfg.ConnectTimeout = time.Second
	_, err := database.NewPool(context.Background(), cfg, "hii-test")
	if err == nil {
		t.Fatal("seharusnya gagal terhubung")
	}
	if strings.Contains(err.Error(), "PasswordRahasia99") {
		t.Errorf("error membocorkan password: %v", err)
	}
}

func TestNewPoolRejectsGarbageURLWithoutLeak(t *testing.T) {
	_, err := database.NewPool(context.Background(), testDBConfig("://bukan url:RahasiaXYZ@"), "hii-test")
	if err == nil {
		t.Fatal("seharusnya error")
	}
	if strings.Contains(err.Error(), "RahasiaXYZ") {
		t.Errorf("error membocorkan kredensial: %v", err)
	}
}

func TestMigrationsUpDownUp(t *testing.T) {
	url := testURL(t)
	ctx := context.Background()

	pool, err := database.NewPool(ctx, testDBConfig(url), "hii-test")
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()

	mg, err := database.NewMigrator(url, "../../migrations")
	if err != nil {
		t.Fatalf("NewMigrator: %v", err)
	}
	defer func() { _ = mg.Close() }()

	hasFn := func() bool {
		var ok bool
		if err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'set_updated_at')").Scan(&ok); err != nil {
			t.Fatalf("cek fungsi: %v", err)
		}
		return ok
	}

	// Mulai dari kondisi bersih agar test dapat diulang.
	if v, _, _ := mg.Version(); v > 0 {
		if err := mg.Steps(-int(v)); err != nil {
			t.Fatalf("reset: %v", err)
		}
	}

	if err := mg.Up(); err != nil {
		t.Fatalf("up: %v", err)
	}
	if !hasFn() {
		t.Error("set_updated_at harus ada setelah up")
	}
	if err := mg.Up(); err != nil {
		t.Errorf("up kedua harus idempotent, dapat: %v", err)
	}
	if v, dirty, err := mg.Version(); err != nil || v < 1 || dirty {
		t.Errorf("versi setelah up: v=%d dirty=%v err=%v", v, dirty, err)
	}

	if err := mg.Steps(-1); err != nil {
		t.Fatalf("down: %v", err)
	}
	if hasFn() {
		t.Error("set_updated_at harus hilang setelah down")
	}

	if err := mg.Up(); err != nil {
		t.Fatalf("up ulang setelah down: %v", err)
	}
	if !hasFn() {
		t.Error("set_updated_at harus kembali setelah up ulang")
	}
}
