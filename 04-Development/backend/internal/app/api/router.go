package api

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// paramPattern mencocokkan parameter path gaya net/http ("{orderId}").
var paramPattern = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)

// ginMux adalah lapisan routing Gin yang tipis (SDD D-02): rute tetap
// ditulis dengan pola net/http ("GET /api/v1/orders/{orderId}") dan handler
// tetap http.Handler biasa, tetapi pencocokan rute dikerjakan Gin. Adaptor
// mengisi r.PathValue dari parameter Gin dan r.Pattern dengan pola aslinya,
// sehingga handler dan middleware (mis. label metrik) tidak perlu berubah.
type ginMux struct {
	engine *gin.Engine
	// aliases adalah awalan tambahan (mis. "/api/v1.2.8") yang melayani rute
	// yang sama dengan apiPrefix. Pola rute tetap kanonik (apiPrefix) agar label
	// metrik stabil.
	aliases []string
	// version dikirim sebagai header X-API-Version bila tidak kosong.
	version string
}

func newGinMux() *ginMux {
	if os.Getenv(gin.EnvGinMode) == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.New()
	// Samakan perilaku dengan ServeMux: tidak ada redirect otomatis untuk
	// garis miring di akhir atau perbaikan huruf besar/kecil pada path.
	e.RedirectTrailingSlash = false
	e.RedirectFixedPath = false
	e.HandleMethodNotAllowed = false
	e.UseRawPath = true
	e.UnescapePathValues = true
	return &ginMux{engine: e}
}

// Handle mendaftarkan h untuk pola "METHOD /path/{param}". Pola "/" tanpa
// method dipakai sebagai jawaban untuk rute yang tidak dikenal.
func (m *ginMux) Handle(pattern string, h http.Handler) {
	method, path, ok := strings.Cut(pattern, " ")
	if !ok {
		if pattern != "/" {
			panic("ginMux: pola tanpa method hanya boleh \"/\": " + pattern)
		}
		m.engine.NoRoute(adapt("", h, m.version))
		m.engine.NoMethod(adapt("", h, m.version))
		return
	}
	handler := adapt(pattern, h, m.version)
	m.engine.Handle(method, paramPattern.ReplaceAllString(path, ":$1"), handler)
	if rest, ok := strings.CutPrefix(path, apiPrefix); ok && (rest == "" || rest[0] == '/') {
		for _, a := range m.aliases {
			m.engine.Handle(method, paramPattern.ReplaceAllString(a+rest, ":$1"), handler)
		}
	}
}

// HandleFunc sama dengan Handle untuk fungsi handler.
func (m *ginMux) HandleFunc(pattern string, f func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(f))
}

func (m *ginMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.engine.ServeHTTP(w, r)
}

// adapt menjalankan http.Handler di dalam Gin. Request yang sama (pointer
// yang sama dengan yang diterima middleware di luar Gin) diberi pola rute dan
// nilai parameter path.
func adapt(pattern string, h http.Handler, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if version != "" {
			c.Writer.Header().Set("X-API-Version", version)
		}
		r := c.Request
		r.Pattern = pattern
		for _, p := range c.Params {
			r.SetPathValue(p.Key, p.Value)
		}
		h.ServeHTTP(c.Writer, r)
	}
}
