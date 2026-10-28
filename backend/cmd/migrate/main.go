package main

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Gagal membaca file .env: ", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL belum diatur di file .env")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Gagal membuka koneksi database: ", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Gagal terhubung ke PostgreSQL: ", err)
	}
	log.Println("Berhasil terhubung ke PostgreSQL!")

	// Pastikan folder migrations mengarah ke backend/migrations.
	migrationPath, err := filepath.Abs("migrations")
	if err != nil {
		log.Fatal("Gagal menentukan lokasi migrations: ", err)
	}

	if _, err := os.Stat(migrationPath); err != nil {
		log.Fatal("Folder migrations tidak ditemukan: ", err)
	}

	sourceURL := "file://" + filepath.ToSlash(migrationPath)
	log.Println("Lokasi migration:", sourceURL)

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		log.Fatal("Gagal membuat migrator: ", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		log.Println("Status migration belum tersedia:", err)
	} else {
		log.Printf("Versi migration: %d, dirty: %v", version, dirty)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Migration sudah terbaru.")
			return
		}
		log.Fatal("Gagal menjalankan migration: ", err)
	}

	log.Println("Migration berhasil dijalankan!")
}
