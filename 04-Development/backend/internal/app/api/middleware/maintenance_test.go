package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hii-stranger1/backend/internal/pkg/token"
)

func TestMaintenance(t *testing.T) {
	tokens := token.NewManager("test-secret-test-secret-test-secret", "test", time.Hour)
	tokenFor := func(role string) string {
		raw, err := tokens.Issue("u1", role)
		if err != nil {
			t.Fatal(err)
		}
		return "Bearer " + raw
	}

	tests := []struct {
		name       string
		on         bool
		readErr    error
		method     string
		path       string
		auth       string
		wantStatus int
	}{
		{"off lets everyone through", false, nil, http.MethodGet, "/api/v1/products", "", http.StatusOK},
		{"on blocks visitor", true, nil, http.MethodGet, "/api/v1/products", "", http.StatusServiceUnavailable},
		{"on blocks customer", true, nil, http.MethodPost, "/api/v1/orders", tokenFor("customer"), http.StatusServiceUnavailable},
		{"on allows staff", true, nil, http.MethodGet, "/api/v1/admin/orders", tokenFor("staff"), http.StatusOK},
		{"on allows owner", true, nil, http.MethodPut, "/api/v1/site-settings", tokenFor("owner"), http.StatusOK},
		{"on keeps login open", true, nil, http.MethodPost, "/api/v1/auth/login", "", http.StatusOK},
		{"on keeps site settings readable", true, nil, http.MethodGet, "/api/v1/site-settings", "", http.StatusOK},
		{"on keeps health open", true, nil, http.MethodGet, "/api/v1/health", "", http.StatusOK},
		{"on keeps preflight open", true, nil, http.MethodOptions, "/api/v1/cart", "", http.StatusOK},
		{"on rejects invalid token", true, nil, http.MethodGet, "/api/v1/products", "Bearer rusak", http.StatusServiceUnavailable},
		{"read error fails open", true, errors.New("db down"), http.MethodGet, "/api/v1/products", "", http.StatusOK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status := func(context.Context) (bool, string, error) { return tc.on, "Sedang dirapikan", tc.readErr }
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			h := Maintenance(tokens, status, "/api/v1", "owner", "staff")(next)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if rec.Code == http.StatusServiceUnavailable {
				body := rec.Body.String()
				if !strings.Contains(body, `"code":"maintenance"`) || !strings.Contains(body, "Sedang dirapikan") {
					t.Fatalf("body 503 = %s, want code maintenance dan pesan dari pengaturan", body)
				}
			}
		})
	}
}
