package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGinMux(t *testing.T) {
	m := newGinMux()
	var gotPattern, gotID, gotSub string
	m.HandleFunc("GET /api/v1/orders/{orderId}", func(w http.ResponseWriter, r *http.Request) {
		gotPattern, gotID = r.Pattern, r.PathValue("orderId")
		w.WriteHeader(http.StatusOK)
	})
	m.HandleFunc("GET /api/v1/orders/{orderId}/payment/proof", func(w http.ResponseWriter, r *http.Request) {
		gotSub = r.PathValue("orderId")
		w.WriteHeader(http.StatusOK)
	})
	m.HandleFunc("GET /api/v1/orders", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	m.HandleFunc("/", notFound)

	tests := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/api/v1/orders/abc", http.StatusOK},
		{http.MethodGet, "/api/v1/orders/xyz/payment/proof", http.StatusOK},
		{http.MethodGet, "/api/v1/orders", http.StatusOK},
		{http.MethodGet, "/api/v1/orders/", http.StatusNotFound},
		{http.MethodDelete, "/api/v1/orders/abc", http.StatusNotFound},
		{http.MethodGet, "/api/v1/tidak-ada", http.StatusNotFound},
	}
	for _, tc := range tests {
		rec := httptest.NewRecorder()
		m.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != tc.want {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.path, rec.Code, tc.want)
		}
	}
	if gotPattern != "GET /api/v1/orders/{orderId}" || gotID != "abc" {
		t.Errorf("pattern = %q, orderId = %q", gotPattern, gotID)
	}
	if gotSub != "xyz" {
		t.Errorf("nested orderId = %q", gotSub)
	}
}

func TestGinMuxVersionAliases(t *testing.T) {
	m := newGinMux()
	m.version = "1.2.8"
	m.aliases = []string{"/api/v1.2.8", "/api/v1.2.7"}
	var gotPattern string
	m.HandleFunc("GET /api/v1/orders/{orderId}", func(w http.ResponseWriter, r *http.Request) {
		gotPattern = r.Pattern
		w.WriteHeader(http.StatusOK)
	})
	m.HandleFunc("/", notFound)
	for _, p := range []string{"/api/v1/orders/a", "/api/v1.2.8/orders/a", "/api/v1.2.7/orders/a"} {
		gotPattern = ""
		rec := httptest.NewRecorder()
		m.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != http.StatusOK || gotPattern != "GET /api/v1/orders/{orderId}" {
			t.Errorf("%s = %d pattern %q", p, rec.Code, gotPattern)
		}
		if rec.Header().Get("X-API-Version") != "1.2.8" {
			t.Errorf("%s: header X-API-Version kosong", p)
		}
	}
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1.2.6/orders/a", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("versi tak terdaftar = %d, want 404", rec.Code)
	}
}
