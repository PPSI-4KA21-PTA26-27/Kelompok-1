package database

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // driver database "pgx5"
	_ "github.com/golang-migrate/migrate/v4/source/file"     // sumber migrasi dari file
)

// Migrator membungkus golang-migrate. golang-migrate memakai advisory lock PostgreSQL,
// sehingga dua proses yang menjalankan migrasi bersamaan tidak saling menimpa. Meski begitu,
// migrasi tetap dijalankan sebagai langkah deploy terpisah (cmd/migrate), bukan dari tiap instance API.
type Migrator struct {
	m *migrate.Migrate
}

func NewMigrator(databaseURL, migrationsDir string) (*Migrator, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return nil, errors.New("DATABASE_URL tidak dapat diparse")
	}
	u.Scheme = "pgx5"

	abs, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("path migrasi: %w", err)
	}
	source := "file://" + filepath.ToSlash(abs)

	m, err := migrate.New(source, u.String())
	if err != nil {
		return nil, fmt.Errorf("inisialisasi migrasi: %w", err)
	}
	return &Migrator{m: m}, nil
}

func (mg *Migrator) Close() error {
	srcErr, dbErr := mg.m.Close()
	return errors.Join(srcErr, dbErr)
}

// Up menerapkan semua migrasi yang belum jalan. Tidak ada perubahan bukan error.
func (mg *Migrator) Up() error { return ignoreNoChange(mg.m.Up()) }

// Steps menjalankan n migrasi (n negatif = rollback sebanyak |n| langkah).
func (mg *Migrator) Steps(n int) error { return ignoreNoChange(mg.m.Steps(n)) }

// Force menandai versi tertentu sebagai versi saat ini dan membersihkan status dirty.
// Hanya untuk pemulihan manual setelah migrasi gagal di tengah jalan.
func (mg *Migrator) Force(version int) error { return mg.m.Force(version) }

// Version mengembalikan versi saat ini. version==0 dan dirty==false berarti belum ada migrasi.
func (mg *Migrator) Version() (version uint, dirty bool, err error) {
	v, d, err := mg.m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return v, d, err
}

func ignoreNoChange(err error) error {
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}
