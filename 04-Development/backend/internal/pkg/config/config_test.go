package config

import (
	"strings"
	"testing"
)

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("DB_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("JWT_SECRET", strings.Repeat("s", 40))
	t.Setenv("JWT_ISSUER", "test")
}

func TestLoadMaxConcurrentRequests(t *testing.T) {
	tests := []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{"", 200, false},
		{"50", 50, false},
		{"0", 0, false},
		{"-1", 0, true},
		{"banyak", 0, true},
	}
	for _, tt := range tests {
		t.Run("MAX_CONCURRENT_REQUESTS="+tt.raw, func(t *testing.T) {
			setRequired(t)
			t.Setenv("MAX_CONCURRENT_REQUESTS", tt.raw)
			cfg, err := Load()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "MAX_CONCURRENT_REQUESTS") {
					t.Fatalf("err = %v, want error MAX_CONCURRENT_REQUESTS", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.MaxConcurrentRequests != tt.want {
				t.Errorf("MaxConcurrentRequests = %d, want %d", cfg.MaxConcurrentRequests, tt.want)
			}
		})
	}
}

func TestLoadTrustProxyHeaders(t *testing.T) {
	tests := []struct {
		raw     string
		want    bool
		wantErr bool
	}{
		{"", false, false},
		{"true", true, false},
		{"false", false, false},
		{"ya", false, true},
	}
	for _, tt := range tests {
		t.Run("TRUST_PROXY_HEADERS="+tt.raw, func(t *testing.T) {
			setRequired(t)
			t.Setenv("TRUST_PROXY_HEADERS", tt.raw)
			cfg, err := Load()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "TRUST_PROXY_HEADERS") {
					t.Fatalf("err = %v, want error TRUST_PROXY_HEADERS", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.TrustProxyHeaders != tt.want {
				t.Errorf("TrustProxyHeaders = %v, want %v", cfg.TrustProxyHeaders, tt.want)
			}
		})
	}
}
