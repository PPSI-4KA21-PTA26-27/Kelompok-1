package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsUsesRoutePattern(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /test-metrics/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	mux.HandleFunc("GET /test-metrics/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	// Timeout membuat salinan request; Metrics tetap harus membaca pola rute.
	h := Chain(mux, Recover(), Timeout(5e9), Metrics())

	route := "/test-metrics/items/{id}"
	before := testutil.ToFloat64(httpRequests.WithLabelValues("GET", route, "418"))
	for _, id := range []string{"a", "b"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/test-metrics/items/"+id, nil))
	}
	if got := testutil.ToFloat64(httpRequests.WithLabelValues("GET", route, "418")) - before; got != 2 {
		t.Fatalf("dua request beda ID harus tercatat di satu rute, dapat %v", got)
	}

	before = testutil.ToFloat64(httpRequests.WithLabelValues("GET", "/test-metrics/panic", "500"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test-metrics/panic", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("panic harus dijawab 500 oleh Recover, dapat %d", rec.Code)
	}
	if got := testutil.ToFloat64(httpRequests.WithLabelValues("GET", "/test-metrics/panic", "500")) - before; got != 1 {
		t.Fatalf("request yang panic harus tercatat sebagai 500, dapat %v", got)
	}
}

func TestRouteLabel(t *testing.T) {
	tests := map[string]string{
		"":                       "unmatched",
		"/":                      "/",
		"GET /api/v1/products":   "/api/v1/products",
		"POST /api/v1/cart/{id}": "/api/v1/cart/{id}",
	}
	for in, want := range tests {
		if got := routeLabel(in); got != want {
			t.Errorf("routeLabel(%q) = %q, want %q", in, got, want)
		}
	}
}
