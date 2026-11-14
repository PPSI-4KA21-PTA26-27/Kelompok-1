package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrik HTTP untuk Prometheus. Label route memakai pola rute (mis.
// /api/v1/products/{productId}), bukan path asli, supaya jumlah seri tetap
// kecil walaupun ID di URL berbeda-beda.
var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Jumlah request HTTP per method, rute, dan status.",
	}, []string{"method", "route", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Durasi request HTTP per method dan rute.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})

	httpInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Jumlah request HTTP yang sedang diproses.",
	})
)

// Metrics mencatat jumlah, status, dan durasi request. Pasang sebagai
// middleware paling dalam (langsung membungkus router) supaya r.Pattern
// sudah diisi mux saat dicatat. Request yang panic dicatat sebagai 500.
func Metrics() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			httpInFlight.Inc()
			// done tetap false bila handler panic. Panic tidak ditangkap di sini
			// supaya Recover tetap mendapat stack trace aslinya.
			done := false
			defer func() {
				httpInFlight.Dec()
				status := rec.status
				if !done {
					status = http.StatusInternalServerError
				}
				route := routeLabel(r.Pattern)
				httpRequests.WithLabelValues(r.Method, route, strconv.Itoa(status)).Inc()
				httpDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
			}()
			next.ServeHTTP(rec, r)
			done = true
		})
	}
}

// routeLabel membuang awalan method dari pola mux ("GET /api/v1/x" -> "/api/v1/x").
// Pola kosong berarti mux tidak menemukan rute (mis. 405).
func routeLabel(pattern string) string {
	if pattern == "" {
		return "unmatched"
	}
	if _, path, ok := strings.Cut(pattern, " "); ok {
		return path
	}
	return pattern
}
