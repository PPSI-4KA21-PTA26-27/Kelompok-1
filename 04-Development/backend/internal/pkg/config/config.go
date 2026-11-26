// Package config membaca konfigurasi aplikasi dari environment variable.
// File .env di direktori kerja ikut dimuat bila ada; variable yang sudah di-set
// di environment tidak ditimpa.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const minJWTSecretLen = 32

// Config adalah seluruh konfigurasi yang dibutuhkan aplikasi saat ini.
type Config struct {
	AppEnv     string
	AppVersion string
	// LegacyAPIVersions: versi lama yang masih dilayani lewat /api/v<versi>.
	LegacyAPIVersions []string
	Port              string

	DBURL              string
	DBMaxConns         int32
	DBMinConns         int32
	DBStatementTimeout time.Duration

	JWTSecret      string
	JWTIssuer      string
	AccessTokenTTL time.Duration

	LogLevel           slog.Level
	MaxBodyBytes       int64
	HTTPRequestTimeout time.Duration
	FrontendURL        string
	// TrustProxyHeaders: baca IP klien dari CF-Connecting-IP / X-Real-IP /
	// X-Forwarded-For. Aktifkan hanya di belakang proxy (mis. Railway).
	TrustProxyHeaders bool

	// SMTP untuk email notifikasi. SMTPHost kosong berarti email dimatikan.
	SMTPHost   string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	SMTPFrom   string
	OwnerEmail string

	// Cloudinary untuk unggah gambar produk. Kosong berarti
	// endpoint unggah menjawab 503.
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	CloudinaryFolder    string

	// RajaOngkir untuk data wilayah & ongkir sampai kelurahan.
	// RajaOngkirOriginID adalah ID kelurahan asal toko (dari endpoint
	// /regions/districts/{id}/subdistricts). Couriers adalah kurir yang dipakai
	// toko (default "pos" = POS Indonesia); layanan termurahnya dipilih otomatis.
	RajaOngkirAPIKey   string
	RajaOngkirBaseURL  string
	RajaOngkirOriginID int
	RajaOngkirCouriers string

	// Redis untuk cache dan penyimpanan data sementara.
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// MaxConcurrentRequests: batas request API yang diproses bersamaan
	// (MAX_CONCURRENT_REQUESTS, default 200); 0 = nonaktif.
	MaxConcurrentRequests int
}

// Load membaca dan memvalidasi konfigurasi. Semua masalah dilaporkan sekaligus
// supaya tidak perlu memperbaiki satu per satu.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("config: load .env: %w", err)
	}

	var l loader

	var level slog.Level
	if err := level.UnmarshalText([]byte(l.str("LOG_LEVEL", "info"))); err != nil {
		l.fail("LOG_LEVEL must be one of debug, info, warn, error")
	}

	cfg := Config{
		AppEnv:            l.str("APP_ENV", "dev"),
		AppVersion:        l.str("APP_VERSION", "1.3.0"),
		LegacyAPIVersions: splitList(l.str("API_LEGACY_VERSIONS", "")),
		Port:              l.str("PORT", "8080"),
		DBURL:             l.required("DB_URL"),
		// Default 10 koneksi: hasil benchmark (4 core) paling stabil di kisaran
		// 2-3x jumlah core. Lebih banyak tidak menambah kecepatan karena
		// PostgreSQL sendiri dibatasi CPU.
		DBMaxConns:         int32(l.integer("DB_MAX_CONNS", 10, 32)), // #nosec G115 -- ParseInt bitSize 32 menjamin muat int32
		DBMinConns:         int32(l.integer("DB_MIN_CONNS", 2, 32)),  // #nosec G115 -- ParseInt bitSize 32 menjamin muat int32
		DBStatementTimeout: l.duration("DB_STATEMENT_TIMEOUT", 5*time.Second),
		JWTSecret:          l.required("JWT_SECRET"),
		JWTIssuer:          l.required("JWT_ISSUER"),
		AccessTokenTTL:     l.duration("ACCESS_TOKEN_TTL", 15*time.Minute),
		LogLevel:           level,
		MaxBodyBytes:       l.integer("MAX_BODY_BYTES", 1<<20, 64),
		HTTPRequestTimeout: l.duration("HTTP_REQUEST_TIMEOUT", 25*time.Second),
		FrontendURL:        strings.TrimRight(l.str("FRONTEND_URL", ""), "/"),
		TrustProxyHeaders:  l.boolean("TRUST_PROXY_HEADERS", false),

		SMTPHost:   l.str("SMTP_HOST", ""),
		SMTPPort:   l.str("SMTP_PORT", "587"),
		SMTPUser:   l.str("SMTP_USER", ""),
		SMTPPass:   l.str("SMTP_PASS", ""),
		SMTPFrom:   l.str("SMTP_FROM", ""),
		OwnerEmail: l.str("OWNER_EMAIL", ""),

		CloudinaryCloudName: l.str("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    l.str("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: l.str("CLOUDINARY_API_SECRET", ""),
		CloudinaryFolder:    l.str("CLOUDINARY_FOLDER", ""),

		RajaOngkirAPIKey:   l.str("RAJAONGKIR_API_KEY", ""),
		RajaOngkirBaseURL:  strings.TrimRight(l.str("RAJAONGKIR_BASE_URL", "https://rajaongkir.komerce.id/api/v1"), "/"),
		RajaOngkirOriginID: int(l.integer("RAJAONGKIR_ORIGIN_ID", 0, 32)),
		RajaOngkirCouriers: l.str("RAJAONGKIR_COURIERS", "pos"),

		RedisAddr:     l.str("REDIS_ADDR", "localhost:6379"),
		RedisPassword: l.str("REDIS_PASSWORD", ""),
		RedisDB:       int(l.integer("REDIS_DB", 0, 32)),

		MaxConcurrentRequests: int(l.integer("MAX_CONCURRENT_REQUESTS", 200, 32)),
	}

	if port, err := strconv.Atoi(cfg.Port); err != nil || port < 1 || port > 65535 {
		l.fail("PORT must be a number between 1 and 65535")
	}
	if cfg.JWTSecret != "" && len(cfg.JWTSecret) < minJWTSecretLen {
		l.fail("JWT_SECRET must be at least %d characters", minJWTSecretLen)
	}
	if cfg.DBMaxConns < 1 {
		l.fail("DB_MAX_CONNS must be at least 1")
	}
	if cfg.DBMinConns < 0 || cfg.DBMinConns > cfg.DBMaxConns {
		l.fail("DB_MIN_CONNS must be between 0 and DB_MAX_CONNS")
	}
	if cfg.MaxConcurrentRequests < 0 {
		l.fail("MAX_CONCURRENT_REQUESTS must be 0 (disabled) or positive")
	}
	if cfg.MaxBodyBytes < 1 {
		l.fail("MAX_BODY_BYTES must be at least 1")
	}
	// Gmail (dan kebanyakan SMTP) wajib login, jadi user & password harus ada.
	if cfg.SMTPHost != "" && (cfg.SMTPUser == "" || cfg.SMTPPass == "") {
		l.fail("SMTP_USER and SMTP_PASS are required when SMTP_HOST is set")
	}
	if cfg.RajaOngkirOriginID < 0 {
		l.fail("RAJAONGKIR_ORIGIN_ID must be a positive subdistrict id")
	}
	if port, err := strconv.Atoi(cfg.SMTPPort); err != nil || port < 1 || port > 65535 {
		l.fail("SMTP_PORT must be a number between 1 and 65535")
	}

	if cfg.RedisDB < 0 || cfg.RedisDB > 15 {
		l.fail("REDIS_DB must be between 0 and 15")
	}

	if err := errors.Join(l.errs...); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

// loader mengumpulkan error selama membaca variable agar bisa dilaporkan bersamaan.
type loader struct {
	errs []error
}

func (l *loader) fail(format string, args ...any) {
	l.errs = append(l.errs, fmt.Errorf(format, args...))
}

func (l *loader) str(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func (l *loader) required(key string) string {
	v := l.str(key, "")
	if v == "" {
		l.fail("%s is required", key)
	}
	return v
}

func (l *loader) integer(key string, def int64, bits int) int64 {
	raw := l.str(key, "")
	if raw == "" {
		return def
	}
	n, err := strconv.ParseInt(raw, 10, bits)
	if err != nil {
		l.fail("%s must be an integer", key)
		return def
	}
	return n
}

func (l *loader) boolean(key string, def bool) bool {
	raw := l.str(key, "")
	if raw == "" {
		return def
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		l.fail("%s must be a boolean (true/false)", key)
		return def
	}
	return b
}

func (l *loader) duration(key string, def time.Duration) time.Duration {
	raw := l.str(key, "")
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		l.fail("%s must be a positive duration such as 15m", key)
		return def
	}
	return d
}

// splitList memecah daftar dipisah koma, membuang spasi dan entri kosong.
func splitList(s string) []string {
	var out []string
	for p := range strings.SplitSeq(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
