package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// maxCachedBody membatasi ukuran satu response yang disimpan di cache.
	maxCachedBody = 1 << 20

	// flushTimeout membatasi waktu pengosongan cache setelah request tulis.
	flushTimeout = 2 * time.Second

	defaultCachePrefix = "cache:resp:"
)

// ResponseCache menyimpan response GET publik di Redis (microcache) supaya
// endpoint yang paling sering dibuka (produk dan kategori) tidak selalu
// menunggu database.
//
// Aturan supaya data tidak basi:
//   - Hanya request tanpa header Authorization yang di-cache. Admin dan user
//     yang login selalu mendapat data langsung dari database.
//   - Hanya response 200 yang disimpan, paling lama ttl.
//   - Setiap request tulis yang sukses (POST/PUT/PATCH/DELETE) di endpoint
//     mana pun mengosongkan seluruh cache, termasuk checkout yang mengubah stok.
//   - Cache ada di Redis, jadi pengosongan berlaku untuk semua instance.
//
// Pengosongan memakai nomor generasi: setiap key memuat generasi saat ini,
// dan Flush cukup menaikkan generasi (INCR) tanpa menghapus key satu per satu.
// Key generasi lama tidak dibaca lagi dan hilang sendiri setelah ttl. Response
// yang mulai dibuat sebelum pengosongan tersimpan di generasi lama, sehingga
// data lama tidak pernah masuk lagi ke cache yang baru.
//
// Bila Redis bermasalah, request tetap dilayani langsung dari handler.
type ResponseCache struct {
	rdb    *redis.Client
	ttl    time.Duration
	prefix string
}

// NewResponseCache membuat cache di Redis dengan masa simpan ttl.
func NewResponseCache(rdb *redis.Client, ttl time.Duration) *ResponseCache {
	return &ResponseCache{rdb: rdb, ttl: ttl, prefix: defaultCachePrefix}
}

// Cached dipasang pada rute GET publik yang boleh di-cache. Header X-Cache
// bernilai HIT atau MISS untuk memudahkan pengecekan.
func (c *ResponseCache) Cached() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || r.Header.Get("Authorization") != "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			gen, err := c.generation(ctx)
			if err != nil {
				logCacheErr(ctx, "read generation", err)
				next.ServeHTTP(w, r)
				return
			}

			key := c.entryKey(gen, r.URL.RequestURI())
			raw, err := c.rdb.Get(ctx, key).Bytes()
			switch {
			case err == nil:
				if contentType, body, ok := decodeEntry(raw); ok {
					h := w.Header()
					h.Set("Content-Type", contentType)
					h.Set("X-Cache", "HIT")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(body) // #nosec G705 -- replay respons handler sendiri dari Redis internal; nosniff dari SecurityHeaders
					return
				}
			case !errors.Is(err, redis.Nil):
				logCacheErr(ctx, "read entry", err)
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-Cache", "MISS")
			rec := &bodyRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			ct := w.Header().Get("Content-Type")
			// hanya JSON yang di-cache, supaya body non-JSON (mis. HTML) tidak diputar ulang dari Redis
			if rec.status != http.StatusOK || rec.tooBig || !strings.HasPrefix(ct, "application/json") {
				return
			}
			entry := encodeEntry(ct, rec.buf.Bytes())
			if err := c.rdb.Set(ctx, key, entry, c.ttl).Err(); err != nil {
				logCacheErr(ctx, "store entry", err)
			}
		})
	}
}

// InvalidateOnWrite dipasang global: setelah request tulis selesai tanpa
// error (status < 400), seluruh cache dikosongkan. skip berisi awalan path
// yang tidak pernah mengubah data yang di-cache (mis. login, keranjang),
// supaya cache tidak terus-menerus kosong karena aktivitas pelanggan.
func (c *ResponseCache) InvalidateOnWrite(skip ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}
			for _, prefix := range skip {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if rec.status >= http.StatusBadRequest {
				return
			}
			// Data sudah berubah, jadi cache tetap dikosongkan walaupun client
			// sudah memutus koneksi.
			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), flushTimeout)
			defer cancel()
			if err := c.Flush(ctx); err != nil {
				slog.ErrorContext(ctx, "response cache: flush failed, public data may be stale until ttl",
					"path", r.URL.Path, "err", err)
			}
		})
	}
}

// Flush mengosongkan seluruh cache dengan menaikkan generasi.
func (c *ResponseCache) Flush(ctx context.Context) error {
	if err := c.rdb.Incr(ctx, c.genKey()).Err(); err != nil {
		return fmt.Errorf("response cache: flush: %w", err)
	}
	return nil
}

func (c *ResponseCache) genKey() string { return c.prefix + "gen" }

func (c *ResponseCache) entryKey(gen, uri string) string { return c.prefix + gen + ":" + uri }

// generation mengembalikan generasi cache saat ini ("0" bila belum pernah
// dikosongkan).
func (c *ResponseCache) generation(ctx context.Context) (string, error) {
	gen, err := c.rdb.Get(ctx, c.genKey()).Result()
	if errors.Is(err, redis.Nil) {
		return "0", nil
	}
	return gen, err
}

// encodeEntry menyimpan Content-Type dan body dalam satu value:
// "<content-type>\n<body>". Content-Type tidak pernah berisi baris baru.
func encodeEntry(contentType string, body []byte) []byte {
	out := make([]byte, 0, len(contentType)+1+len(body))
	out = append(out, contentType...)
	out = append(out, '\n')
	return append(out, body...)
}

func decodeEntry(raw []byte) (contentType string, body []byte, ok bool) {
	i := bytes.IndexByte(raw, '\n')
	if i < 0 {
		return "", nil, false
	}
	return string(raw[:i]), raw[i+1:], true
}

// logCacheErr mencatat error Redis, kecuali karena request dibatalkan client.
func logCacheErr(ctx context.Context, op string, err error) {
	if ctx.Err() != nil {
		return
	}
	slog.WarnContext(ctx, "response cache: "+op, "err", err)
}

// bodyRecorder meneruskan response ke client sambil menyalin body-nya.
type bodyRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
	tooBig bool
}

func (b *bodyRecorder) WriteHeader(code int) {
	b.status = code
	b.ResponseWriter.WriteHeader(code)
}

func (b *bodyRecorder) Write(p []byte) (int, error) {
	// middleware ini hanya dipasang pada endpoint JSON: default ke JSON, bukan sniffing
	if b.ResponseWriter.Header().Get("Content-Type") == "" {
		b.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	if !b.tooBig {
		if b.buf.Len()+len(p) > maxCachedBody {
			b.tooBig = true
			b.buf.Reset()
		} else {
			b.buf.Write(p)
		}
	}
	return b.ResponseWriter.Write(p)
}
