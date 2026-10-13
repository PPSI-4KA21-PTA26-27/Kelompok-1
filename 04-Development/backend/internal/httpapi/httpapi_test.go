package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"kel1/backend/internal/config"
	"kel1/backend/internal/httpapi"
	"kel1/backend/internal/httpapi/middleware"
	"kel1/backend/internal/httpapi/response"
)

func init() { gin.SetMode(gin.TestMode) }

type fakeDB struct{ err error }

func (f fakeDB) Ping(context.Context) error { return f.err }

func testConfig(prod bool) *config.Config {
	env := config.EnvDevelopment
	if prod {
		env = config.EnvProduction
	}
	return &config.Config{
		Env: env,
		HTTP: config.HTTPConfig{
			Port:               "8080",
			RequestTimeout:     5 * time.Second,
			MaxBodyBytes:       1024,
			CORSAllowedOrigins: []string{"https://shop.example.com"},
		},
	}
}

func newEngine(t *testing.T, cfg *config.Config, db httpapi.Pinger, logBuf io.Writer) *gin.Engine {
	t.Helper()
	if logBuf == nil {
		logBuf = io.Discard
	}
	log := slog.New(slog.NewJSONHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	e, err := httpapi.NewRouter(httpapi.Dependencies{Config: cfg, Logger: log, DB: db})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return e
}

func do(e *gin.Engine, method, path string, mut func(*http.Request)) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if mut != nil {
		mut(req)
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("body bukan JSON envelope: %v\n%s", err, w.Body.String())
	}
	return env
}

func TestHealthz(t *testing.T) {
	e := newEngine(t, testConfig(false), fakeDB{}, nil)
	w := do(e, http.MethodGet, "/healthz", nil)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	if env := decode(t, w); !env.Success {
		t.Error("success harus true")
	}
}

func TestReadyz(t *testing.T) {
	t.Run("db sehat", func(t *testing.T) {
		e := newEngine(t, testConfig(false), fakeDB{}, nil)
		if w := do(e, http.MethodGet, "/readyz", nil); w.Code != 200 {
			t.Fatalf("status = %d", w.Code)
		}
	})
	t.Run("db mati: 503 tanpa membocorkan detail", func(t *testing.T) {
		secretErr := errors.New("dial tcp 10.1.2.3:5432: password=hunter2 refused")
		var logs bytes.Buffer
		e := newEngine(t, testConfig(false), fakeDB{err: secretErr}, &logs)
		w := do(e, http.MethodGet, "/readyz", nil)
		if w.Code != 503 {
			t.Fatalf("status = %d", w.Code)
		}
		env := decode(t, w)
		if env.Success || env.Error == nil || env.Error.Code != response.CodeUnavailable {
			t.Errorf("envelope salah: %+v", env)
		}
		for _, leaked := range []string{"hunter2", "10.1.2.3", "dial tcp"} {
			if strings.Contains(w.Body.String(), leaked) {
				t.Errorf("response membocorkan %q", leaked)
			}
		}
		if !strings.Contains(logs.String(), "readiness_check_failed") {
			t.Error("kegagalan readiness harus tercatat di log")
		}
	})
}

func TestNotFoundAndMethodNotAllowedUseEnvelope(t *testing.T) {
	e := newEngine(t, testConfig(false), fakeDB{}, nil)

	w := do(e, http.MethodGet, "/tidak-ada", nil)
	if w.Code != 404 || decode(t, w).Error.Code != response.CodeNotFound {
		t.Errorf("404 salah: %d %s", w.Code, w.Body.String())
	}

	w = do(e, http.MethodPost, "/healthz", nil)
	if w.Code != 405 || decode(t, w).Error.Code != response.CodeMethodNotAllowed {
		t.Errorf("405 salah: %d %s", w.Code, w.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	dev := do(newEngine(t, testConfig(false), fakeDB{}, nil), http.MethodGet, "/healthz", nil)
	for k, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"Cache-Control":          "no-store",
	} {
		if got := dev.Header().Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if dev.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS tidak boleh dikirim di development")
	}

	prod := do(newEngine(t, testConfig(true), fakeDB{}, nil), http.MethodGet, "/healthz", nil)
	if prod.Header().Get("Strict-Transport-Security") == "" {
		t.Error("HSTS harus dikirim di production")
	}
}

func TestRequestID(t *testing.T) {
	e := newEngine(t, testConfig(false), fakeDB{}, nil)

	t.Run("dibuat bila tidak ada", func(t *testing.T) {
		id := do(e, http.MethodGet, "/healthz", nil).Header().Get(middleware.RequestIDHeader)
		if len(id) != 32 {
			t.Errorf("request id = %q", id)
		}
	})
	t.Run("ID aman dari klien dipakai ulang", func(t *testing.T) {
		w := do(e, http.MethodGet, "/healthz", func(r *http.Request) { r.Header.Set(middleware.RequestIDHeader, "trace-abc-12345") })
		if got := w.Header().Get(middleware.RequestIDHeader); got != "trace-abc-12345" {
			t.Errorf("request id = %q", got)
		}
	})
	t.Run("ID berbahaya diganti (cegah log injection)", func(t *testing.T) {
		w := do(e, http.MethodGet, "/healthz", func(r *http.Request) { r.Header.Set(middleware.RequestIDHeader, "x\nlevel=ERROR forged") })
		got := w.Header().Get(middleware.RequestIDHeader)
		if strings.ContainsAny(got, "\n ") || len(got) != 32 {
			t.Errorf("request id berbahaya lolos: %q", got)
		}
	})
}

func TestCORS(t *testing.T) {
	e := newEngine(t, testConfig(false), fakeDB{}, nil)

	allowed := do(e, http.MethodOptions, "/healthz", func(r *http.Request) {
		r.Header.Set("Origin", "https://shop.example.com")
		r.Header.Set("Access-Control-Request-Method", "GET")
	})
	if got := allowed.Header().Get("Access-Control-Allow-Origin"); got != "https://shop.example.com" {
		t.Errorf("origin sah tidak diizinkan: %q", got)
	}
	if allowed.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("credentials harus diizinkan untuk origin sah")
	}

	denied := do(e, http.MethodGet, "/healthz", func(r *http.Request) { r.Header.Set("Origin", "https://evil.example.net") })
	if got := denied.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("origin asing tidak boleh mendapat ACAO, dapat %q", got)
	}
}

func TestCORSRejectsWildcard(t *testing.T) {
	if _, err := middleware.CORS([]string{"*"}); err == nil {
		t.Error("wildcard harus ditolak")
	}
	if _, err := middleware.CORS(nil); err == nil {
		t.Error("daftar origin kosong harus ditolak")
	}
}

func TestBodyLimit(t *testing.T) {
	e := newEngine(t, testConfig(false), fakeDB{}, nil)
	big := strings.Repeat("a", 2048)
	w := do(e, http.MethodPost, "/healthz", func(r *http.Request) {
		r.Body = io.NopCloser(strings.NewReader(big))
		r.ContentLength = int64(len(big))
	})
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", w.Code)
	}
	if env := decode(t, w); env.Error == nil || env.Error.Code != response.CodePayloadTooLarge {
		t.Errorf("envelope salah: %s", w.Body.String())
	}
}

func TestPanicIsRecoveredWithoutLeak(t *testing.T) {
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	e := gin.New()
	e.Use(middleware.RequestID(), middleware.AccessLog(log), middleware.Recovery(log))
	e.GET("/boom", func(*gin.Context) { panic("rahasia-internal-xyz") })

	w := do(e, http.MethodGet, "/boom", nil)
	if w.Code != 500 {
		t.Fatalf("status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "rahasia-internal-xyz") {
		t.Error("detail panic bocor ke klien")
	}
	if env := decode(t, w); env.Error == nil || env.Error.Code != response.CodeInternal {
		t.Errorf("envelope salah: %s", w.Body.String())
	}
	if !strings.Contains(logs.String(), "panic_recovered") || !strings.Contains(logs.String(), "rahasia-internal-xyz") {
		t.Error("panic harus tercatat lengkap di log server")
	}
	if !strings.Contains(logs.String(), `"status":500`) {
		t.Error("access log harus mencatat status 500 untuk panic yang dipulihkan")
	}
}

func TestAccessLogOmitsQueryAndSecrets(t *testing.T) {
	var logs bytes.Buffer
	e := newEngine(t, testConfig(false), fakeDB{}, &logs)
	do(e, http.MethodGet, "/healthz?token=SECRETTOKEN123", func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer SECRETBEARER456")
	})
	out := logs.String()
	if !strings.Contains(out, "http_request") {
		t.Fatalf("access log tidak ada: %q", out)
	}
	for _, leaked := range []string{"SECRETTOKEN123", "SECRETBEARER456"} {
		if strings.Contains(out, leaked) {
			t.Errorf("log membocorkan %q", leaked)
		}
	}
}

func TestForgedForwardedForIsIgnoredByDefault(t *testing.T) {
	var logs bytes.Buffer
	e := newEngine(t, testConfig(false), fakeDB{}, &logs)
	do(e, http.MethodGet, "/healthz", func(r *http.Request) {
		r.RemoteAddr = "203.0.113.9:4444"
		r.Header.Set("X-Forwarded-For", "1.2.3.4")
	})
	if strings.Contains(logs.String(), "1.2.3.4") {
		t.Error("X-Forwarded-For dari klien tidak tepercaya tidak boleh dipakai sebagai IP klien")
	}
	if !strings.Contains(logs.String(), "203.0.113.9") {
		t.Error("IP klien harus berasal dari RemoteAddr")
	}
}

func TestRequestTimeoutPropagatesToContext(t *testing.T) {
	e := gin.New()
	e.Use(middleware.RequestTimeout(50 * time.Millisecond))
	var hasDeadline bool
	e.GET("/x", func(c *gin.Context) {
		_, hasDeadline = c.Request.Context().Deadline()
		response.OK(c, "ok", nil)
	})
	do(e, http.MethodGet, "/x", nil)
	if !hasDeadline {
		t.Error("context request harus punya deadline")
	}
}
