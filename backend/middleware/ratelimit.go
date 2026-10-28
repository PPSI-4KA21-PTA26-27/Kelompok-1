package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"hii-stranger1/backend/response"
)

// Alur: routes -> middleware.RateLimit -> handler (login, daftar, lupa sandi).
// Membatasi jumlah percobaan per alamat IP dalam satu jendela waktu, untuk menahan tebak-tebakan sandi.
// Catatan: hitungan disimpan di memori server ini, jadi hilang saat restart dan tidak dibagi
// antar-server. Cukup untuk satu server; kalau nanti jalan di banyak server, pindah ke penyimpanan bersama.

type visit struct {
	count   int
	resetAt time.Time
}

// RateLimit mengizinkan paling banyak limit request per IP dalam satu window. Sisanya dapat 429.
// Setiap pemanggilan membuat hitungan sendiri, jadi login dan daftar tidak saling mempengaruhi.
func RateLimit(limit int, window time.Duration) Middleware {
	var (
		mu        sync.Mutex
		visits    = make(map[string]*visit)
		lastSweep = time.Now()
	)

	// allow mengembalikan true kalau request boleh lanjut; kalau tidak, berapa lama harus menunggu.
	allow := func(ip string, now time.Time) (bool, time.Duration) {
		mu.Lock()
		defer mu.Unlock()

		// Bersihkan IP yang jendelanya sudah lewat, supaya map tidak membesar terus.
		if now.Sub(lastSweep) > window {
			for k, v := range visits {
				if now.After(v.resetAt) {
					delete(visits, k)
				}
			}
			lastSweep = now
		}

		v, ok := visits[ip]
		if !ok || now.After(v.resetAt) {
			visits[ip] = &visit{count: 1, resetAt: now.Add(window)}
			return true, 0
		}
		if v.count >= limit {
			return false, v.resetAt.Sub(now)
		}
		v.count++
		return true, 0
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// RemoteAddr dipakai apa adanya. Header X-Forwarded-For sengaja diabaikan karena bisa dipalsukan klien.
			// Kalau nanti di belakang reverse proxy, IP asli perlu diambil dari proxy yang dipercaya.
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			ok, wait := allow(ip, time.Now())
			if !ok {
				w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
				response.Error(w, http.StatusTooManyRequests, "terlalu banyak percobaan, coba lagi nanti")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
