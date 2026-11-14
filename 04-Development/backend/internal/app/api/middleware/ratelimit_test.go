package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimiterAllow(t *testing.T) {
	l := &limiter{limit: 2, window: time.Minute, visitors: map[string]*visitor{}}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 2; i++ {
		if ok, _ := l.allow("1.1.1.1", now); !ok {
			t.Fatalf("request ke-%d harus lolos", i+1)
		}
	}
	ok, wait := l.allow("1.1.1.1", now.Add(10*time.Second))
	if ok {
		t.Fatal("request ketiga dalam jendela harus ditolak")
	}
	if wait != 50*time.Second {
		t.Errorf("retryAfter = %v, want 50s", wait)
	}
	if ok, _ := l.allow("2.2.2.2", now); !ok {
		t.Error("IP lain punya kuota sendiri")
	}
	if ok, _ := l.allow("1.1.1.1", now.Add(time.Minute+time.Second)); !ok {
		t.Error("setelah jendela lewat, kuota harus pulih")
	}
}

func TestLimiterSweep(t *testing.T) {
	l := &limiter{limit: 1, window: time.Minute, visitors: map[string]*visitor{}}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l.allow("a", now)
	l.allow("b", now)
	l.allow("c", now.Add(2*time.Minute))
	if _, ok := l.visitors["a"]; ok {
		t.Error("entri kedaluwarsa harus dibersihkan")
	}
	if len(l.visitors) != 1 {
		t.Errorf("len(visitors) = %d, want 1", len(l.visitors))
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	h := RateLimit(1, time.Minute, false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	do := func(addr string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	if rec := do("10.0.0.1:1111"); rec.Code != http.StatusOK {
		t.Fatalf("request pertama status = %d", rec.Code)
	}
	// Port berbeda tetap dihitung sebagai IP yang sama.
	rec := do("10.0.0.1:2222")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request kedua status = %d, want 429", rec.Code)
	}
	if ra := rec.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Errorf("Retry-After = %q, want > 0", ra)
	}
	if rec := do("10.0.0.2:1111"); rec.Code != http.StatusOK {
		t.Errorf("IP lain status = %d, want 200", rec.Code)
	}
}

func TestClientIP(t *testing.T) {
	for addr, want := range map[string]string{"1.2.3.4:80": "1.2.3.4", "[::1]:80": "::1", "tanpa-port": "tanpa-port"} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = addr
		if got := ClientIP(r, false); got != want {
			t.Errorf("clientIP(%q) = %q, want %q", addr, got, want)
		}
	}
}

func TestClientIPProxyHeaders(t *testing.T) {
	tests := []struct {
		name    string
		trust   bool
		headers map[string]string
		want    string
	}{
		{"flag off mengabaikan header", false, map[string]string{"CF-Connecting-IP": "1.1.1.1", "X-Real-IP": "2.2.2.2", "X-Forwarded-For": "3.3.3.3"}, "10.0.0.1"},
		{"tanpa header pakai RemoteAddr", true, nil, "10.0.0.1"},
		{"CF-Connecting-IP prioritas pertama", true, map[string]string{"CF-Connecting-IP": "1.1.1.1", "X-Real-IP": "2.2.2.2", "X-Forwarded-For": "3.3.3.3"}, "1.1.1.1"},
		{"X-Real-IP sebelum XFF", true, map[string]string{"X-Real-IP": "2.2.2.2", "X-Forwarded-For": "3.3.3.3"}, "2.2.2.2"},
		{"XFF berantai ambil elemen pertama", true, map[string]string{"X-Forwarded-For": " 3.3.3.3 , 4.4.4.4, 10.0.0.9"}, "3.3.3.3"},
		{"CF tidak valid lanjut ke X-Real-IP", true, map[string]string{"CF-Connecting-IP": "bukan-ip", "X-Real-IP": "2.2.2.2"}, "2.2.2.2"},
		{"semua header tidak valid pakai RemoteAddr", true, map[string]string{"CF-Connecting-IP": "x", "X-Real-IP": "1.2.3", "X-Forwarded-For": "salah, 4.4.4.4"}, "10.0.0.1"},
		{"IPv6 di CF-Connecting-IP", true, map[string]string{"CF-Connecting-IP": "2001:db8::1"}, "2001:db8::1"},
		{"IPv6 di XFF", true, map[string]string{"X-Forwarded-For": "2001:DB8::2, 10.0.0.9"}, "2001:db8::2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = "10.0.0.1:5555"
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}
			if got := ClientIP(r, tt.trust); got != tt.want {
				t.Errorf("ClientIP = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimitByUser(t *testing.T) {
	h := RateLimitByUser(2, time.Minute, false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	do := func(userID, addr string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = addr
		if userID != "" {
			req = req.WithContext(context.WithValue(req.Context(), principalKey{}, Principal{UserID: userID}))
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	// Di bawah batas lolos, walau user yang sama berpindah IP.
	if c := do("u1", "10.0.0.1:1"); c != http.StatusOK {
		t.Fatalf("request 1 = %d", c)
	}
	if c := do("u1", "10.0.0.2:1"); c != http.StatusOK {
		t.Fatalf("request 2 = %d", c)
	}
	if c := do("u1", "10.0.0.3:1"); c != http.StatusTooManyRequests {
		t.Errorf("request 3 user sama = %d, want 429", c)
	}
	// User lain di IP yang sama punya jatah sendiri; tanpa Principal dihitung per IP.
	if c := do("u2", "10.0.0.1:1"); c != http.StatusOK {
		t.Errorf("user lain = %d, want 200", c)
	}
	if c := do("", "10.0.0.1:1"); c != http.StatusOK {
		t.Errorf("anonim = %d, want 200", c)
	}
}
