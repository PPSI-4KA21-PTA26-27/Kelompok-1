package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const testOrigin = "http://localhost:3000"

func corsHandler(origin string, called *bool) http.Handler {
	return CORS(origin)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	}))
}

func TestCORSAllowedOrigin(t *testing.T) {
	var called bool
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Header.Set("Origin", testOrigin)
	rec := httptest.NewRecorder()
	corsHandler(testOrigin, &called).ServeHTTP(rec, req)

	if !called || rec.Code != http.StatusOK {
		t.Fatalf("called = %v, status = %d", called, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("Allow-Origin = %q", got)
	}
	if rec.Header().Get("Vary") != "Origin" {
		t.Errorf("Vary = %q, want Origin", rec.Header().Get("Vary"))
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Error("credentials tidak boleh diaktifkan")
	}
}

func TestCORSPreflight(t *testing.T) {
	var called bool
	req := httptest.NewRequest(http.MethodOptions, "/cart", nil)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodDelete)
	rec := httptest.NewRecorder()
	corsHandler(testOrigin, &called).ServeHTTP(rec, req)

	if called {
		t.Error("preflight tidak boleh diteruskan ke handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
	for _, h := range []string{"Access-Control-Allow-Methods", "Access-Control-Allow-Headers", "Access-Control-Max-Age"} {
		if rec.Header().Get(h) == "" {
			t.Errorf("header %s kosong", h)
		}
	}
}

func TestCORSOtherOriginOrDisabled(t *testing.T) {
	tests := []struct{ name, allowed, origin, method string }{
		{"origin lain", testOrigin, "http://evil.example", http.MethodGet},
		{"preflight origin lain", testOrigin, "http://evil.example", http.MethodOptions},
		{"tanpa origin", testOrigin, "", http.MethodGet},
		{"CORS dimatikan", "", testOrigin, http.MethodGet},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var called bool
			req := httptest.NewRequest(tc.method, "/", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			req.Header.Set("Access-Control-Request-Method", http.MethodGet)
			rec := httptest.NewRecorder()
			corsHandler(tc.allowed, &called).ServeHTTP(rec, req)
			if !called {
				t.Error("request harus diteruskan ke handler")
			}
			if rec.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Error("Allow-Origin tidak boleh diset")
			}
		})
	}
}
