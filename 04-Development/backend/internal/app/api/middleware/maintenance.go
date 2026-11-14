package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"time"

	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/token"
)

// maintenanceCacheTTL membatasi seberapa sering status maintenance dibaca dari
// basis data. Perubahan oleh owner berlaku paling lambat setelah selang ini.
const maintenanceCacheTTL = 5 * time.Second

// defaultMaintenanceMessage dipakai bila owner tidak mengisi pesan.
const defaultMaintenanceMessage = "Website sedang dalam pemeliharaan. Silakan coba lagi nanti."

// MaintenanceStatus membaca apakah mode maintenance aktif beserta pesannya
// (site_settings.maintenance_mode dan maintenance_message).
type MaintenanceStatus func(ctx context.Context) (on bool, message string, err error)

// Maintenance menegakkan mode maintenance di server (SDD 3.5, KF-15): saat
// aktif, permintaan pelanggan dan pengunjung dijawab 503 dengan pesan
// pemeliharaan, sedangkan pemanggil dengan token role allowedRoles (staff dan
// owner) tetap dilayani. Beberapa rute selalu dibuka supaya staff dan owner
// tetap bisa masuk dan frontend bisa menampilkan pesan: cek kesehatan, baca
// pengaturan situs, login, lupa/reset password, verifikasi email, dan profil
// sendiri. Bila status gagal dibaca, permintaan dilanjutkan (fail open) agar
// gangguan basis data tidak menutup seluruh situs.
func Maintenance(tokens *token.Manager, status MaintenanceStatus, prefix string, allowedRoles ...string) Middleware {
	open := map[string]bool{
		"GET " + prefix + "/health":                true,
		"GET " + prefix + "/site-settings":         true,
		"POST " + prefix + "/auth/login":           true,
		"POST " + prefix + "/auth/password/forgot": true,
		"POST " + prefix + "/auth/password/reset":  true,
		"POST " + prefix + "/auth/verify-email":    true,
		"GET " + prefix + "/users/me":              true,
	}
	cache := &maintenanceCache{read: status}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions || open[r.Method+" "+r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			on, msg := cache.get(r.Context())
			if !on || hasRole(tokens, r, allowedRoles) {
				next.ServeHTTP(w, r)
				return
			}
			if msg == "" {
				msg = defaultMaintenanceMessage
			}
			w.Header().Set("Retry-After", "300")
			response.Error(w, http.StatusServiceUnavailable, "maintenance", msg)
		})
	}
}

// hasRole bernilai true bila request membawa token valid dengan salah satu role.
func hasRole(tokens *token.Manager, r *http.Request, roles []string) bool {
	raw, ok := bearerToken(r)
	if !ok {
		return false
	}
	claims, err := tokens.Parse(raw)
	return err == nil && slices.Contains(roles, claims.Role)
}

// maintenanceCache menyimpan status maintenance terakhir selama
// maintenanceCacheTTL.
type maintenanceCache struct {
	read MaintenanceStatus

	mu      sync.Mutex
	until   time.Time
	on      bool
	message string
}

func (c *maintenanceCache) get(ctx context.Context) (bool, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Now().Before(c.until) {
		return c.on, c.message
	}
	on, msg, err := c.read(ctx)
	if err != nil {
		slog.WarnContext(ctx, "maintenance: read site settings failed", "error", err)
		return false, ""
	}
	c.on, c.message, c.until = on, msg, time.Now().Add(maintenanceCacheTTL)
	return on, msg
}
