package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hii-stranger1/backend/internal/pkg/token"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name       string
		principal  *Principal
		wantStatus int
		wantCalled bool
	}{
		{"no principal", nil, http.StatusUnauthorized, false},
		{"wrong role", &Principal{UserID: "u1", Role: "customer"}, http.StatusForbidden, false},
		{"right role", &Principal{UserID: "u1", Role: "owner"}, http.StatusOK, true},
		{"second allowed role", &Principal{UserID: "u2", Role: "staff"}, http.StatusOK, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.principal != nil {
				req = req.WithContext(context.WithValue(req.Context(), principalKey{}, *tc.principal))
			}
			rec := httptest.NewRecorder()
			RequireRole("owner", "staff")(next).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if called != tc.wantCalled {
				t.Errorf("next called = %v, want %v", called, tc.wantCalled)
			}
		})
	}
}

func TestAuthenticateOptional(t *testing.T) {
	tokens := token.NewManager("0123456789abcdef0123456789abcdef", "test-issuer", time.Minute)
	valid, err := tokens.Issue("u1", "owner")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	tests := []struct {
		name          string
		header        string
		wantStatus    int
		wantCalled    bool
		wantPrincipal bool
	}{
		{"no header is anonymous", "", http.StatusOK, true, false},
		{"valid token", "Bearer " + valid, http.StatusOK, true, true},
		{"invalid token", "Bearer garbage", http.StatusUnauthorized, false, false},
		{"wrong scheme", "Basic abc", http.StatusUnauthorized, false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var called, hasPrincipal bool
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				_, hasPrincipal = PrincipalFrom(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			AuthenticateOptional(tokens)(next).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if called != tc.wantCalled {
				t.Errorf("next called = %v, want %v", called, tc.wantCalled)
			}
			if hasPrincipal != tc.wantPrincipal {
				t.Errorf("has principal = %v, want %v", hasPrincipal, tc.wantPrincipal)
			}
		})
	}
}
