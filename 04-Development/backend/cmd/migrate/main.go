// Command migrate menjalankan migrasi database. Dijalankan sebagai langkah deploy terpisah
// (Railway pre-deploy / docker compose), BUKAN dari tiap instance API.
//
// Pemakaian:
//
//	migrate up              terapkan semua migrasi baru
//	migrate down N          rollback N langkah (N wajib, tidak ada "down semua")
//	migrate version         tampilkan versi saat ini
//	migrate force VERSION   pulihkan status dirty ke VERSION
//
// Env: DATABASE_URL (wajib), MIGRATIONS_DIR (default "migrations").
// Di production, down dan force ditolak kecuali MIGRATE_ALLOW_DESTRUCTIVE=true.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"kel1/backend/internal/database"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("perintah wajib: up | down N | version | force VERSION")
	}
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return errors.New("DATABASE_URL wajib diisi")
	}
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}

	mg, err := database.NewMigrator(dbURL, dir)
	if err != nil {
		return err
	}
	defer func() { _ = mg.Close() }()

	switch args[0] {
	case "up":
		if err := mg.Up(); err != nil {
			return fmt.Errorf("up gagal: %w", err)
		}
		return printVersion(mg, "migrasi selesai")

	case "down":
		if err := guardDestructive(); err != nil {
			return err
		}
		n, err := positiveArg(args, "down N")
		if err != nil {
			return err
		}
		if err := mg.Steps(-n); err != nil {
			return fmt.Errorf("down gagal: %w", err)
		}
		return printVersion(mg, "rollback selesai")

	case "version":
		return printVersion(mg, "versi saat ini")

	case "force":
		if err := guardDestructive(); err != nil {
			return err
		}
		v, err := positiveArg(args, "force VERSION")
		if err != nil {
			return err
		}
		if err := mg.Force(v); err != nil {
			return fmt.Errorf("force gagal: %w", err)
		}
		return printVersion(mg, "versi dipaksa")

	default:
		return fmt.Errorf("perintah tidak dikenal %q", args[0])
	}
}

func guardDestructive() error {
	if os.Getenv("APP_ENV") == "production" && os.Getenv("MIGRATE_ALLOW_DESTRUCTIVE") != "true" {
		return errors.New("down/force diblokir di production; set MIGRATE_ALLOW_DESTRUCTIVE=true bila memang disengaja")
	}
	return nil
}

func positiveArg(args []string, usage string) (int, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("pemakaian: migrate %s", usage)
	}
	n, err := strconv.Atoi(args[1])
	if err != nil || n < 1 {
		return 0, fmt.Errorf("argumen harus angka >= 1 (migrate %s)", usage)
	}
	return n, nil
}

func printVersion(mg *database.Migrator, prefix string) error {
	v, dirty, err := mg.Version()
	if err != nil {
		return fmt.Errorf("membaca versi: %w", err)
	}
	fmt.Printf("%s: versi=%d dirty=%t\n", prefix, v, dirty)
	return nil
}
