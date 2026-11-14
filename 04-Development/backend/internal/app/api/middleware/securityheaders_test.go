package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	const hsts = "max-age=31536000; includeSubDomains"
	tests := []struct {
		name     string
		trust    bool
		tls      bool
		xfp      string
		wantHSTS bool
	}{
		{"http biasa tanpa HSTS", false, false, "", false},
		{"XFP diabaikan bila proxy tidak dipercaya", false, false, "https", false},
		{"TLS langsung", false, true, "", true},
		{"XFP https dengan proxy dipercaya", true, false, "https", true},
		{"XFP berantai elemen pertama", true, false, "HTTPS, http", true},
		{"XFP http dengan proxy dipercaya", true, false, "http", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := SecurityHeaders(tt.trust)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
			if tt.tls {
				req.TLS = &tls.ConnectionState{}
			}
			if tt.xfp != "" {
				req.Header.Set("X-Forwarded-Proto", tt.xfp)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			want := map[string]string{
				"X-Content-Type-Options":  "nosniff",
				"X-Frame-Options":         "DENY",
				"Referrer-Policy":         "no-referrer",
				"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
				"Permissions-Policy":      "geolocation=(), microphone=(), camera=()",
			}
			for k, v := range want {
				if got := rec.Header().Values(k); len(got) != 1 || got[0] != v {
					t.Errorf("%s = %q, want [%q]", k, got, v)
				}
			}
			got := rec.Header().Get("Strict-Transport-Security")
			if tt.wantHSTS && got != hsts {
				t.Errorf("HSTS = %q, want %q", got, hsts)
			}
			if !tt.wantHSTS && got != "" {
				t.Errorf("HSTS = %q, want kosong", got)
			}
		})
	}
}
