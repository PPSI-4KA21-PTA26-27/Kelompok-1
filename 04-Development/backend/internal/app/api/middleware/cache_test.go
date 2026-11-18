package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// countingHandler menjawab 200 dengan isi berupa jumlah pemanggilan, supaya
// test bisa tahu apakah response datang dari cache atau dari handler.
func countingHandler(calls *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(strconv.Itoa(*calls)))
	})
}

func get(h http.Handler, path, auth string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// newTestCache memakai Redis di REDIS_TEST_ADDR (default localhost:6379,
// DB 15) dengan prefix unik per test. Test dilewati bila Redis tidak ada.
func newTestCache(t *testing.T) *ResponseCache {
	t.Helper()
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		if os.Getenv("CI") != "" {
			t.Fatalf("redis wajib di CI (%s): %v", addr, err)
		}
		t.Skipf("redis tidak tersedia di %s: %v", addr, err)
	}

	c := NewResponseCache(rdb, time.Minute)
	c.prefix = "test:" + t.Name() + ":" + strconv.FormatInt(time.Now().UnixNano(), 10) + ":"
	t.Cleanup(func() {
		ctx := context.Background()
		iter := rdb.Scan(ctx, 0, c.prefix+"*", 100).Iterator()
		for iter.Next(ctx) {
			rdb.Del(ctx, iter.Val())
		}
		_ = rdb.Close()
	})
	return c
}

func TestResponseCache(t *testing.T) {
	var calls int
	c := newTestCache(t)
	h := Chain(countingHandler(&calls), c.Cached())

	r := get(h, "/products?page=1", "")
	if r.Header().Get("X-Cache") != "MISS" || r.Body.String() != "1" {
		t.Fatalf("request pertama harus MISS, dapat %q %q", r.Header().Get("X-Cache"), r.Body.String())
	}
	r = get(h, "/products?page=1", "")
	if r.Header().Get("X-Cache") != "HIT" || r.Body.String() != "1" {
		t.Fatalf("request kedua harus HIT dari cache, dapat %q %q", r.Header().Get("X-Cache"), r.Body.String())
	}
	if ct := r.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type dari cache harus sama, dapat %q", ct)
	}
	if r := get(h, "/products?page=2", ""); r.Body.String() != "2" {
		t.Fatalf("query berbeda harus jadi entri berbeda, dapat %q", r.Body.String())
	}
	if r := get(h, "/products?page=1", "Bearer x"); r.Body.String() != "3" {
		t.Fatalf("request dengan token tidak boleh dari cache, dapat %q", r.Body.String())
	}

	// Request tulis yang sukses mengosongkan cache.
	write := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), c.InvalidateOnWrite("/cart"))
	write.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/products", nil))
	if r := get(h, "/products?page=1", ""); r.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("setelah POST cache harus kosong, dapat %q", r.Header().Get("X-Cache"))
	}

	// Path yang dikecualikan tidak mengosongkan cache.
	write.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/cart/items", nil))
	if r := get(h, "/products?page=1", ""); r.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("POST /cart tidak boleh mengosongkan cache, dapat %q", r.Header().Get("X-Cache"))
	}

	// Request tulis yang gagal tidak mengosongkan cache.
	failed := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}), c.InvalidateOnWrite())
	failed.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/products", nil))
	if r := get(h, "/products?page=1", ""); r.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("POST yang gagal tidak boleh mengosongkan cache, dapat %q", r.Header().Get("X-Cache"))
	}
}

func TestResponseCacheSkipsErrorsAndStaleWrites(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()

	notFound := Chain(http.NotFoundHandler(), c.Cached())
	get(notFound, "/products/x", "")
	if n := c.rdb.Exists(ctx, c.entryKey("0", "/products/x")).Val(); n != 0 {
		t.Fatal("response selain 200 tidak boleh disimpan")
	}

	// Cache dikosongkan saat response sedang dibuat: hasilnya tidak boleh
	// terbaca di generasi baru.
	var calls int
	slow := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			if err := c.Flush(r.Context()); err != nil {
				t.Errorf("flush: %v", err)
			}
		}
		_, _ = w.Write([]byte(strconv.Itoa(calls)))
	}), c.Cached())
	get(slow, "/categories", "")
	if r := get(slow, "/categories", ""); r.Header().Get("X-Cache") != "MISS" || r.Body.String() != "2" {
		t.Fatalf("response yang dibuat sebelum cache dikosongkan tidak boleh dipakai, dapat %q %q",
			r.Header().Get("X-Cache"), r.Body.String())
	}
}

func TestResponseCacheRedisDown(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 200 * time.Millisecond})
	t.Cleanup(func() { _ = rdb.Close() })
	c := NewResponseCache(rdb, time.Minute)

	var calls int
	h := Chain(countingHandler(&calls), c.Cached())
	for i := 1; i <= 2; i++ {
		r := get(h, "/products", "")
		if r.Code != http.StatusOK || r.Body.String() != strconv.Itoa(i) {
			t.Fatalf("redis mati: request harus tetap dilayani handler, dapat %d %q", r.Code, r.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	write := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), c.InvalidateOnWrite())
	write.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/products", nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("redis mati: request tulis tidak boleh gagal, dapat %d", rec.Code)
	}
}

func TestResponseCacheSkipsNonJSON(t *testing.T) {
	var calls int
	c := newTestCache(t)
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<b>x</b>"))
	}), c.Cached())
	get(h, "/page", "")
	if r := get(h, "/page", ""); r.Header().Get("X-Cache") == "HIT" || calls != 2 {
		t.Fatalf("respons non-JSON tidak boleh di-cache, calls=%d", calls)
	}
}

func TestBodyRecorderDefaultsToJSON(t *testing.T) {
	for _, tc := range []struct{ set, want string }{
		{"", "application/json; charset=utf-8"},
		{"text/plain", "text/plain"},
	} {
		rec := httptest.NewRecorder()
		if tc.set != "" {
			rec.Header().Set("Content-Type", tc.set)
		}
		b := &bodyRecorder{ResponseWriter: rec}
		_, _ = b.Write([]byte("{}"))
		if got := rec.Header().Get("Content-Type"); got != tc.want {
			t.Fatalf("Content-Type = %q, want %q", got, tc.want)
		}
	}
}
