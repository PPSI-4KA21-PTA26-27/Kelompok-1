package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func hit(h http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRateLimit(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := RateLimit(3, time.Hour)(next)

	for i := 1; i <= 3; i++ {
		if rec := hit(h, "10.0.0.1:1111"); rec.Code != http.StatusOK {
			t.Fatalf("request ke-%d status = %d, want 200", i, rec.Code)
		}
	}

	rec := hit(h, "10.0.0.1:2222") // port beda, IP sama: tetap dihitung satu pengguna
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request ke-4 status = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("header Retry-After seharusnya diisi")
	}

	if rec := hit(h, "10.0.0.2:1111"); rec.Code != http.StatusOK {
		t.Errorf("IP lain terkena limit IP pertama, status = %d", rec.Code)
	}
}

func TestRateLimitWindowResets(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := RateLimit(1, 50*time.Millisecond)(next)

	if rec := hit(h, "10.0.0.1:1"); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec := hit(h, "10.0.0.1:1"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	time.Sleep(80 * time.Millisecond)
	if rec := hit(h, "10.0.0.1:1"); rec.Code != http.StatusOK {
		t.Errorf("setelah jendela lewat status = %d, want 200", rec.Code)
	}
}
