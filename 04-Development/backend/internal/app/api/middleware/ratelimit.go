package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"hii-stranger1/backend/internal/pkg/response"
)

type visitor struct {
	count   int
	resetAt time.Time
}

type limiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	visitors  map[string]*visitor
	lastSweep time.Time
}

// allow mencatat satu request dari key. Bila batas terlampaui, mengembalikan
// false beserta sisa waktu tunggu. Entri kedaluwarsa dibersihkan secara berkala
// saat ada request, jadi tidak perlu goroutine terpisah.
func (l *limiter) allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) > l.window {
		for k, v := range l.visitors {
			if now.After(v.resetAt) {
				delete(l.visitors, k)
			}
		}
		l.lastSweep = now
	}

	v, ok := l.visitors[key]
	if !ok || now.After(v.resetAt) {
		l.visitors[key] = &visitor{count: 1, resetAt: now.Add(l.window)}
		return true, 0
	}
	if v.count >= l.limit {
		return false, v.resetAt.Sub(now)
	}
	v.count++
	return true, 0
}

// RateLimit membatasi jumlah request per alamat IP dalam satu jendela waktu
// tetap. Setiap pemanggilan membuat limiter sendiri, jadi pasang satu per rute.
// Penghitung tersimpan di memori satu instance. IP klien diambil lewat
// ClientIP; trustProxy true berarti header proxy dipercaya (lihat ClientIP).
func RateLimit(limit int, window time.Duration, trustProxy bool) Middleware {
	return rateLimitBy(limit, window, func(r *http.Request) string { return ClientIP(r, trustProxy) })
}

// RateLimitByUser seperti RateLimit, tetapi request yang membawa Principal
// (pasang setelah Authenticate) dihitung per user ID, sehingga user di balik
// NAT/IP yang sama tidak saling menghabiskan jatah. Tanpa Principal, per IP.
// Satu nilai Middleware berbagi satu penghitung; pakai ulang untuk satu kelompok rute.
func RateLimitByUser(limit int, window time.Duration, trustProxy bool) Middleware {
	return rateLimitBy(limit, window, func(r *http.Request) string {
		if p, ok := PrincipalFrom(r.Context()); ok && p.UserID != "" {
			return "u:" + p.UserID
		}
		return "ip:" + ClientIP(r, trustProxy)
	})
}

func rateLimitBy(limit int, window time.Duration, key func(*http.Request) string) Middleware {
	l := &limiter{limit: limit, window: window, visitors: make(map[string]*visitor)}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ok, retryAfter := l.allow(key(r), time.Now())
			if !ok {
				w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
				response.Error(w, http.StatusTooManyRequests, "too_many_requests", "Terlalu banyak permintaan. Coba lagi nanti.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// proxyIPHeaders berurutan menurut prioritas. CF-Connecting-IP diisi
// Cloudflare, X-Real-IP oleh banyak reverse proxy, lalu X-Forwarded-For.
var proxyIPHeaders = []string{"CF-Connecting-IP", "X-Real-IP", "X-Forwarded-For"}

// ClientIP mengembalikan IP klien. Bila trustProxy true, header proxy dibaca
// sesuai prioritas proxyIPHeaders (untuk X-Forwarded-For hanya elemen pertama)
// dan nilai yang bukan IP valid dilewati. Bila flag false atau tidak ada header
// yang valid, dipakai host dari RemoteAddr. Hanya aktifkan di belakang proxy
// yang menimpa header ini; tanpa proxy, klien bisa memalsukannya.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		for _, name := range proxyIPHeaders {
			v := r.Header.Get(name)
			if name == "X-Forwarded-For" {
				v, _, _ = strings.Cut(v, ",")
			}
			if ip := net.ParseIP(strings.TrimSpace(v)); ip != nil {
				return ip.String()
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
