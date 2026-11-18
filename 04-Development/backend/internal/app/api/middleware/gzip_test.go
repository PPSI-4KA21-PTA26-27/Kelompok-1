package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzip(t *testing.T) {
	h := Gzip()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", r.URL.Query().Get("ct"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	do := func(ae, ct string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/?ct="+ct, nil)
		if ae != "" {
			req.Header.Set("Accept-Encoding", ae)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	rec := do("gzip", "application/json")
	zr, err := gzip.NewReader(rec.Body)
	if err != nil || rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("respons JSON harus gzip: %v %v", err, rec.Header())
	}
	if b, _ := io.ReadAll(zr); string(b) != `{"ok":true}` {
		t.Fatalf("isi salah: %s", b)
	}
	if rec := do("", "application/json"); rec.Header().Get("Content-Encoding") != "" {
		t.Fatal("tanpa Accept-Encoding tidak boleh gzip")
	}
	h204 := Gzip()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec204 := httptest.NewRecorder()
	h204.ServeHTTP(rec204, req)
	if rec204.Body.Len() != 0 || rec204.Header().Get("Content-Encoding") != "" {
		t.Fatal("204 harus tanpa body dan tanpa gzip")
	}
	if rec := do("gzip", "image/png"); rec.Header().Get("Content-Encoding") != "" {
		t.Fatal("non-JSON tidak boleh gzip")
	}
}
