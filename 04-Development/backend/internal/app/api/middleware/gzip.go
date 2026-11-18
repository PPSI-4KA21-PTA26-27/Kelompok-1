package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(nil) }}

// Gzip mengompres respons JSON bila klien mengirim Accept-Encoding: gzip.
// Tipe lain (gambar, bukti bayar) dilewati karena sudah terkompresi.
func Gzip() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Add("Vary", "Accept-Encoding")
			gw := &gzipWriter{ResponseWriter: w, out: w}
			defer gw.close()
			next.ServeHTTP(gw, r)
		})
	}
}

type gzipWriter struct {
	http.ResponseWriter
	zw      *gzip.Writer
	out     io.Writer // zw bila dikompres, selain itu writer asli
	started bool
}

func (g *gzipWriter) start() {
	if g.started {
		return
	}
	g.started = true
	h := g.Header()
	if strings.HasPrefix(h.Get("Content-Type"), "application/json") && h.Get("Content-Encoding") == "" {
		h.Del("Content-Length")
		h.Set("Content-Encoding", "gzip")
		g.zw = gzipPool.Get().(*gzip.Writer)
		g.zw.Reset(g.ResponseWriter)
		g.out = g.zw
	}
}

func (g *gzipWriter) WriteHeader(code int) {
	// Status tanpa body (1xx, 204, 304) tidak boleh diberi tubuh gzip.
	if code < 200 || code == http.StatusNoContent || code == http.StatusNotModified {
		g.started = true
	} else {
		g.start()
	}
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	g.start()
	return g.out.Write(b)
}

func (g *gzipWriter) close() {
	if g.zw != nil {
		_ = g.zw.Close()
		gzipPool.Put(g.zw)
	}
}

// Flush meneruskan data terkompresi ke klien (mis. streaming).
func (g *gzipWriter) Flush() {
	if g.zw != nil {
		_ = g.zw.Flush()
	}
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap supaya http.ResponseController tetap menjangkau writer asli.
func (g *gzipWriter) Unwrap() http.ResponseWriter { return g.ResponseWriter }
