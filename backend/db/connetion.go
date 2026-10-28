package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func Connect() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL belum diatur")
	}

	// Connect to PostgreSQL
	var err error
	DB, err = sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Gagal membuka koneksi database:", err)
	}

	// Verify connection
	if err := DB.Ping(); err != nil {
		log.Fatal("Gagal terhubung ke PostgreSQL:", err)
	}

	log.Println("Berhasil terhubung ke PostgreSQL")
}

func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Println("Gagal menutup koneksi database:", err)
		}
	}
}
