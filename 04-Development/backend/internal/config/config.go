package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

type Config struct {
	AppEnv      string
	AppPort     string
	DB          DBConfig
	JWTSecret   string
	JWTExpiry   time.Duration
	CORSOrigins []string
}

// Load membaca konfigurasi dari environment (.env otomatis dibaca kalau ada).
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("file .env tidak ditemukan, pakai environment variable sistem")
	}

	return &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "kel1_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		},
		JWTSecret:   getEnv("JWT_SECRET", "ganti-secret-ini-di-production"),
		JWTExpiry:   time.Duration(getEnvInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"), ","),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
