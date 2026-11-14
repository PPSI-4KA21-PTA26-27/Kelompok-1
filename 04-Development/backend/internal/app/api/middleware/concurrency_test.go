package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestMaxConcurrent(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, 2)
	h := MaxConcurrent(2, "/api/v1/health")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		started <- struct{}{}
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	do := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}

	var wg sync.WaitGroup
	codes := make([]int, 2)
	for i := range codes {
		wg.Add(1)
		go func(i int) { defer wg.Done(); codes[i] = do("/api/v1/products").Code }(i)
	}
	<-started
	<-started

	// Kapasitas penuh: request berikutnya ditolak, health check tetap lolos.
	rec := do("/api/v1/products")
	if rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("saat penuh status = %d Retry-After = %q, want 503 + Retry-After", rec.Code, rec.Header().Get("Retry-After"))
	}
	// Path lain yang berakhiran /health tetap dihitung.
	if rec := do("/api/v1/products/health"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("/products/health saat penuh status = %d, want 503", rec.Code)
	}
	healthDone := make(chan int)
	go func() { healthDone <- do("/api/v1/health").Code }()
	<-started // health masuk handler walau semua slot terpakai
	close(release)
	if code := <-healthDone; code != http.StatusOK {
		t.Errorf("health status = %d, want 200", code)
	}
	wg.Wait()
	for i, c := range codes {
		if c != http.StatusOK {
			t.Errorf("request ke-%d status = %d, want 200", i+1, c)
		}
	}
	if rec := do("/api/v1/products"); rec.Code != http.StatusOK {
		t.Errorf("setelah slot kosong status = %d, want 200", rec.Code)
	}
}

func TestMaxConcurrentDisabled(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	rec := httptest.NewRecorder()
	MaxConcurrent(0)(inner).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
