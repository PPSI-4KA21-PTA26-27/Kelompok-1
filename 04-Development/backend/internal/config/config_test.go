package config

import (
	"strings"
	"testing"
	"time"
)

const goodSecret = "k9Xq2mV7pL4nR8tY1wZ6bC3dF5hJ0aSe"

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func baseEnv(extra map[string]string) map[string]string {
	m := map[string]string{
		"DATABASE_URL":      "postgres://app:pw@localhost:5432/hii?sslmode=disable",
		"JWT_ACCESS_SECRET": goodSecret,
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestFromEnvDefaults(t *testing.T) {
	cfg, err := FromEnv(env(baseEnv(nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Env != EnvDevelopment || cfg.HTTP.Port != "8080" {
		t.Errorf("default env/port salah: %s %s", cfg.Env, cfg.HTTP.Port)
	}
	if cfg.JWT.AccessTTL != 15*time.Minute || cfg.JWT.RefreshTTL != 720*time.Hour {
		t.Errorf("default TTL salah: %v %v", cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	}
	if cfg.JWT.Issuer != "hii-stranger-api" {
		t.Errorf("issuer default salah: %s", cfg.JWT.Issuer)
	}
	if got := cfg.HTTP.CORSAllowedOrigins; len(got) != 1 || got[0] != "http://localhost:5173" {
		t.Errorf("CORS default dev salah: %v", got)
	}
}

func TestFromEnvRejectsInvalid(t *testing.T) {
	prodOK := baseEnv(map[string]string{
		"APP_ENV":              "production",
		"CORS_ALLOWED_ORIGINS": "https://hii-stranger.example.com",
	})
	with := func(over map[string]string) map[string]string {
		m := map[string]string{}
		for k, v := range prodOK {
			m[k] = v
		}
		for k, v := range over {
			m[k] = v
		}
		return m
	}

	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"database url kosong", baseEnv(map[string]string{"DATABASE_URL": ""}), "DATABASE_URL wajib"},
		{"database url bukan postgres", baseEnv(map[string]string{"DATABASE_URL": "mysql://a:b@h/db"}), "DATABASE_URL harus berformat"},
		{"secret kosong", baseEnv(map[string]string{"JWT_ACCESS_SECRET": ""}), "JWT_ACCESS_SECRET wajib"},
		{"secret pendek", baseEnv(map[string]string{"JWT_ACCESS_SECRET": "pendek"}), "terlalu pendek"},
		{"secret monoton", baseEnv(map[string]string{"JWT_ACCESS_SECRET": strings.Repeat("ab", 20)}), "terlalu monoton"},
		{"secret placeholder", baseEnv(map[string]string{"JWT_ACCESS_SECRET": "ganti-dengan-string-random-panjang-0123456789"}), "placeholder"},
		{"app env salah", baseEnv(map[string]string{"APP_ENV": "staging"}), "APP_ENV harus"},
		{"port salah", baseEnv(map[string]string{"PORT": "abc"}), "PORT harus"},
		{"ttl format salah", baseEnv(map[string]string{"ACCESS_TOKEN_TTL": "lima menit"}), "ACCESS_TOKEN_TTL"},
		{"access ttl terlalu panjang", baseEnv(map[string]string{"ACCESS_TOKEN_TTL": "24h"}), "ACCESS_TOKEN_TTL harus antara"},
		{"refresh <= access", baseEnv(map[string]string{"ACCESS_TOKEN_TTL": "1h", "REFRESH_TOKEN_TTL": "1h"}), "REFRESH_TOKEN_TTL harus lebih besar"},
		{"cors wildcard", baseEnv(map[string]string{"CORS_ALLOWED_ORIGINS": "*"}), "wildcard"},
		{"cors subdomain wildcard", baseEnv(map[string]string{"CORS_ALLOWED_ORIGINS": "https://*.example.com"}), "wildcard"},
		{"cors dengan path", baseEnv(map[string]string{"CORS_ALLOWED_ORIGINS": "https://a.com/app"}), "tidak boleh memuat path"},
		{"prod cors http", with(map[string]string{"CORS_ALLOWED_ORIGINS": "http://a.com"}), "https di production"},
		{"prod tanpa cors", with(map[string]string{"CORS_ALLOWED_ORIGINS": ""}), "wajib diisi di production"},
		{"min conns > max", baseEnv(map[string]string{"DB_MIN_CONNS": "20", "DB_MAX_CONNS": "5"}), "DB_MIN_CONNS"},
		{"proxy tidak valid", baseEnv(map[string]string{"TRUSTED_PROXIES": "bukan-ip"}), "TRUSTED_PROXIES"},
		{"request timeout >= write", baseEnv(map[string]string{"HTTP_REQUEST_TIMEOUT": "40s"}), "HTTP_REQUEST_TIMEOUT"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := FromEnv(env(tc.env))
			if err == nil {
				t.Fatalf("seharusnya error, dapat config: %+v", cfg)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q tidak memuat %q", err, tc.wantErr)
			}
		})
	}
}

func TestErrorsNeverLeakSecrets(t *testing.T) {
	const pw = "SuperRahasiaPw123"
	_, err := FromEnv(env(map[string]string{
		"DATABASE_URL":      "mysql://app:" + pw + "@host/db",
		"JWT_ACCESS_SECRET": "pendek-sekali",
	}))
	if err == nil {
		t.Fatal("seharusnya error")
	}
	for _, leaked := range []string{pw, "pendek-sekali"} {
		if strings.Contains(err.Error(), leaked) {
			t.Errorf("pesan error membocorkan nilai sensitif %q", leaked)
		}
	}
}

func TestFromEnvCollectsAllErrors(t *testing.T) {
	_, err := FromEnv(env(map[string]string{"PORT": "x"}))
	if err == nil {
		t.Fatal("seharusnya error")
	}
	for _, want := range []string{"PORT", "DATABASE_URL", "JWT_ACCESS_SECRET"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error tidak memuat %s: %v", want, err)
		}
	}
}

func TestFromEnvProductionValid(t *testing.T) {
	cfg, err := FromEnv(env(baseEnv(map[string]string{
		"APP_ENV":              "production",
		"CORS_ALLOWED_ORIGINS": "https://a.example.com/, https://b.example.com, https://a.example.com",
		"TRUSTED_PROXIES":      "10.0.0.0/8, 127.0.0.1",
	})))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("IsProduction harus true")
	}
	want := []string{"https://a.example.com", "https://b.example.com"}
	got := cfg.HTTP.CORSAllowedOrigins
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("origin dinormalisasi dan di-dedupe, dapat %v", got)
	}
	if len(cfg.Warnings) == 0 {
		t.Error("sslmode=disable di production harus memberi warning")
	}
}

func TestFrontendURLFallbackForCORS(t *testing.T) {
	cfg, err := FromEnv(env(baseEnv(map[string]string{"FRONTEND_URL": "http://localhost:3000"})))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got := cfg.HTTP.CORSAllowedOrigins; len(got) != 1 || got[0] != "http://localhost:3000" {
		t.Errorf("fallback FRONTEND_URL salah: %v", got)
	}
}
