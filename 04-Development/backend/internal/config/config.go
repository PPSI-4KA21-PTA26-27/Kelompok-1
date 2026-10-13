// Package config memuat dan memvalidasi konfigurasi dari environment variable.
//
// Prinsip:
//   - Tidak ada secret default. Secret yang wajib dan tidak valid membuat startup gagal.
//   - Pesan error tidak pernah memuat nilai secret atau DATABASE_URL (berisi password).
//   - Semua masalah dikumpulkan sekaligus supaya bisa diperbaiki dalam satu kali coba.
package config

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

type DatabaseConfig struct {
	URL              string
	MaxConns         int32
	MinConns         int32
	MaxConnLifetime  time.Duration
	MaxConnIdleTime  time.Duration
	ConnectTimeout   time.Duration
	StatementTimeout time.Duration
}

type JWTConfig struct {
	AccessSecret string
	Issuer       string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
}

type HTTPConfig struct {
	Port               string
	ReadHeaderTimeout  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
	RequestTimeout     time.Duration
	MaxBodyBytes       int64
	CORSAllowedOrigins []string
	TrustedProxies     []string
}

type Config struct {
	Env         string
	LogLevel    string
	FrontendURL string
	HTTP        HTTPConfig
	Database    DatabaseConfig
	JWT         JWTConfig

	// Warnings berisi konfigurasi yang valid tetapi sebaiknya ditinjau.
	Warnings []string
}

func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

// Load membaca file .env (jika ada, hanya untuk kenyamanan lokal) lalu environment.
// Environment variable yang sudah di-set selalu menang atas isi .env.
func Load() (*Config, error) {
	_ = godotenv.Load() // .env opsional; error "file tidak ada" memang diabaikan
	return FromEnv(os.Getenv)
}

// FromEnv membangun Config dari fungsi pembaca env. Dipisahkan dari Load agar mudah diuji.
func FromEnv(getenv func(string) string) (*Config, error) {
	p := &parser{get: getenv}

	cfg := &Config{
		Env:      p.oneOf("APP_ENV", EnvDevelopment, EnvDevelopment, EnvTest, EnvProduction),
		LogLevel: p.oneOf("LOG_LEVEL", "info", "debug", "info", "warn", "error"),
	}
	prod := cfg.Env == EnvProduction

	cfg.HTTP = HTTPConfig{
		Port:              p.port("PORT", "8080"),
		ReadHeaderTimeout: p.duration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second, time.Second, time.Minute),
		ReadTimeout:       p.duration("HTTP_READ_TIMEOUT", 15*time.Second, time.Second, 5*time.Minute),
		WriteTimeout:      p.duration("HTTP_WRITE_TIMEOUT", 30*time.Second, time.Second, 5*time.Minute),
		IdleTimeout:       p.duration("HTTP_IDLE_TIMEOUT", 60*time.Second, time.Second, 10*time.Minute),
		ShutdownTimeout:   p.duration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second, time.Second, 2*time.Minute),
		RequestTimeout:    p.duration("HTTP_REQUEST_TIMEOUT", 25*time.Second, time.Second, 5*time.Minute),
		MaxBodyBytes:      p.integer("MAX_BODY_BYTES", 1<<20, 1<<10, 10<<20),
		TrustedProxies:    p.proxies("TRUSTED_PROXIES"),
	}
	if cfg.HTTP.RequestTimeout >= cfg.HTTP.WriteTimeout {
		p.fail("HTTP_REQUEST_TIMEOUT harus lebih kecil dari HTTP_WRITE_TIMEOUT")
	}

	cfg.FrontendURL = p.origin("FRONTEND_URL", prod, false)
	cfg.HTTP.CORSAllowedOrigins = p.corsOrigins(cfg.FrontendURL, prod)

	cfg.Database = DatabaseConfig{
		URL:              p.databaseURL("DATABASE_URL", prod, &cfg.Warnings),
		MaxConns:         int32(p.integer("DB_MAX_CONNS", 10, 1, 200)),
		MinConns:         int32(p.integer("DB_MIN_CONNS", 2, 0, 200)),
		MaxConnLifetime:  p.duration("DB_MAX_CONN_LIFETIME", 30*time.Minute, time.Minute, 24*time.Hour),
		MaxConnIdleTime:  p.duration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute, time.Second, 24*time.Hour),
		ConnectTimeout:   p.duration("DB_CONNECT_TIMEOUT", 10*time.Second, time.Second, time.Minute),
		StatementTimeout: p.duration("DB_STATEMENT_TIMEOUT", 15*time.Second, 100*time.Millisecond, 5*time.Minute),
	}
	if cfg.Database.MinConns > cfg.Database.MaxConns {
		p.fail("DB_MIN_CONNS tidak boleh lebih besar dari DB_MAX_CONNS")
	}

	cfg.JWT = JWTConfig{
		AccessSecret: p.jwtSecret("JWT_ACCESS_SECRET"),
		Issuer:       p.str("JWT_ISSUER", "hii-stranger-api"),
		AccessTTL:    p.duration("ACCESS_TOKEN_TTL", 15*time.Minute, time.Minute, time.Hour),
		RefreshTTL:   p.duration("REFRESH_TOKEN_TTL", 30*24*time.Hour, time.Hour, 90*24*time.Hour),
	}
	if cfg.JWT.RefreshTTL <= cfg.JWT.AccessTTL {
		p.fail("REFRESH_TOKEN_TTL harus lebih besar dari ACCESS_TOKEN_TTL")
	}

	if err := errors.Join(p.errs...); err != nil {
		return nil, fmt.Errorf("konfigurasi tidak valid:\n%w", err)
	}
	return cfg, nil
}

// parser mengumpulkan error alih-alih berhenti di error pertama.
type parser struct {
	get  func(string) string
	errs []error
}

func (p *parser) fail(format string, args ...any) {
	p.errs = append(p.errs, fmt.Errorf(format, args...))
}

func (p *parser) str(key, def string) string {
	if v := strings.TrimSpace(p.get(key)); v != "" {
		return v
	}
	return def
}

func (p *parser) oneOf(key, def string, allowed ...string) string {
	v := strings.ToLower(p.str(key, def))
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	p.fail("%s harus salah satu dari: %s", key, strings.Join(allowed, ", "))
	return def
}

func (p *parser) port(key, def string) string {
	v := p.str(key, def)
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 || n > 65535 {
		p.fail("%s harus berupa angka 1-65535", key)
		return def
	}
	return v
}

func (p *parser) integer(key string, def, min, max int64) int64 {
	raw := p.str(key, "")
	if raw == "" {
		return def
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < min || n > max {
		p.fail("%s harus berupa angka antara %d dan %d", key, min, max)
		return def
	}
	return n
}

func (p *parser) duration(key string, def, min, max time.Duration) time.Duration {
	raw := p.str(key, "")
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		p.fail("%s bukan durasi yang valid (contoh: 15m, 720h)", key)
		return def
	}
	if d < min || d > max {
		p.fail("%s harus antara %s dan %s", key, min, max)
		return def
	}
	return d
}

func (p *parser) list(key string) []string {
	var out []string
	for _, part := range strings.Split(p.get(key), ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func (p *parser) proxies(key string) []string {
	items := p.list(key)
	for _, it := range items {
		if _, err := netip.ParsePrefix(it); err == nil {
			continue
		}
		if _, err := netip.ParseAddr(it); err == nil {
			continue
		}
		p.fail("%s memuat alamat IP/CIDR yang tidak valid", key)
		return nil
	}
	return items
}

// origin memvalidasi satu origin (scheme://host[:port]) tanpa path.
func (p *parser) origin(key string, prod, required bool) string {
	raw := p.str(key, "")
	if raw == "" {
		if required {
			p.fail("%s wajib diisi", key)
		}
		return ""
	}
	o, err := normalizeOrigin(raw, prod)
	if err != nil {
		p.fail("%s: %v", key, err)
		return ""
	}
	return o
}

func (p *parser) corsOrigins(frontendURL string, prod bool) []string {
	const key = "CORS_ALLOWED_ORIGINS"
	var out []string
	seen := map[string]bool{}
	for _, raw := range p.list(key) {
		o, err := normalizeOrigin(raw, prod)
		if err != nil {
			p.fail("%s: %v", key, err)
			continue
		}
		if !seen[o] {
			seen[o] = true
			out = append(out, o)
		}
	}
	if len(out) == 0 && frontendURL != "" {
		out = append(out, frontendURL)
	}
	if len(out) == 0 {
		if prod {
			p.fail("%s (atau FRONTEND_URL) wajib diisi di production", key)
		} else {
			out = []string{"http://localhost:5173"}
		}
	}
	return out
}

func normalizeOrigin(raw string, prod bool) (string, error) {
	if strings.Contains(raw, "*") {
		return "", errors.New("wildcard '*' tidak diizinkan; tulis origin lengkap, contoh https://example.com")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", errors.New("harus berupa origin lengkap, contoh https://example.com")
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "", errors.New("origin tidak boleh memuat path, query, fragment, atau kredensial")
	}
	if prod && u.Scheme != "https" {
		return "", errors.New("harus memakai https di production")
	}
	return u.Scheme + "://" + u.Host, nil
}

// databaseURL memvalidasi DATABASE_URL tanpa pernah menyertakannya dalam pesan error.
func (p *parser) databaseURL(key string, prod bool, warnings *[]string) string {
	raw := strings.TrimSpace(p.get(key))
	if raw == "" {
		p.fail("%s wajib diisi (contoh: postgres://user:pass@host:5432/dbname)", key)
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" || strings.Trim(u.Path, "/") == "" {
		p.fail("%s harus berformat postgres://user:pass@host:port/dbname", key)
		return ""
	}
	if prod && u.Query().Get("sslmode") == "disable" {
		*warnings = append(*warnings, "DATABASE_URL memakai sslmode=disable di production; pastikan koneksi berada di jaringan privat atau aktifkan TLS")
	}
	return raw
}

// jwtSecret menolak secret yang pendek, monoton, atau masih berupa placeholder.
func (p *parser) jwtSecret(key string) string {
	v := p.get(key)
	switch {
	case strings.TrimSpace(v) == "":
		p.fail("%s wajib diisi (acak, minimal 32 karakter)", key)
		return ""
	case len(v) < 32:
		p.fail("%s terlalu pendek, minimal 32 karakter", key)
		return ""
	case distinctChars(v) < 10:
		p.fail("%s terlalu monoton; gunakan nilai acak", key)
		return ""
	case looksLikePlaceholder(v):
		p.fail("%s masih berupa placeholder; ganti dengan nilai acak", key)
		return ""
	}
	return v
}

func distinctChars(s string) int {
	seen := map[rune]struct{}{}
	for _, r := range s {
		seen[r] = struct{}{}
	}
	return len(seen)
}

func looksLikePlaceholder(s string) bool {
	l := strings.ToLower(s)
	for _, bad := range []string{"changeme", "change-me", "change_me", "ganti", "replace", "your-secret", "your_secret", "placeholder", "example"} {
		if strings.Contains(l, bad) {
			return true
		}
	}
	return false
}
