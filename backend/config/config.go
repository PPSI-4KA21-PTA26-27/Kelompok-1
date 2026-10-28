package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Alur: main -> config.Load() -> hasilnya dipakai routes (JWT, SMTP, URL frontend).
// Semua nilai rahasia dibaca dari environment (.env), tidak pernah ditulis di kode.

// minJWTSecretLen: kunci JWT yang terlalu pendek mudah ditebak.
const minJWTSecretLen = 32

type SMTP struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

// Configured true kalau pengiriman email sudah diatur lengkap.
func (s SMTP) Configured() bool {
	return s.Host != "" && s.Port != "" && s.From != ""
}

type Config struct {
	Port string

	// JWTSecret dipakai middleware.Auth dan services.AuthService untuk tanda tangan token.
	JWTSecret []byte
	// JWTTTL = masa berlaku access token (desain: sesi 15 menit).
	JWTTTL time.Duration

	// FrontendURL dipakai services.AuthService untuk membuat tautan reset kata sandi.
	FrontendURL string
	SMTP        SMTP
}

// Load membaca konfigurasi dari environment. Koneksi database ditangani paket db.
// Program berhenti dengan pesan jelas kalau konfigurasi wajib belum benar.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// PORT harus angka saja (contoh 8080), bukan alamat lengkap seperti http://localhost:8080.
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		log.Fatalf("PORT di .env harus berupa angka 1-65535 (contoh: 8080), isinya: %q", port)
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < minJWTSecretLen {
		log.Fatalf("JWT_SECRET wajib diisi di .env, minimal %d karakter", minJWTSecretLen)
	}

	ttlMinutes := 15
	if v := os.Getenv("JWT_TTL_MINUTES"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			log.Fatalf("JWT_TTL_MINUTES harus berupa angka menit lebih dari 0, isinya: %q", v)
		}
		ttlMinutes = n
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	return Config{
		Port:        port,
		JWTSecret:   []byte(secret),
		JWTTTL:      time.Duration(ttlMinutes) * time.Minute,
		FrontendURL: frontendURL,
		SMTP: SMTP{
			Host: os.Getenv("SMTP_HOST"),
			Port: os.Getenv("SMTP_PORT"),
			User: os.Getenv("SMTP_USER"),
			Pass: os.Getenv("SMTP_PASS"),
			From: os.Getenv("SMTP_FROM"),
		},
	}
}
