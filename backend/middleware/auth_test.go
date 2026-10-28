package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hii-stranger1/backend/models"
	"hii-stranger1/backend/services"
)

var testSecret = []byte("rahasia-untuk-test-minimal-32-karakter!")

func tokenFor(t *testing.T, id int64, role models.Role, issuedAt time.Time) string {
	t.Helper()
	tok, err := services.IssueToken(testSecret, 15*time.Minute, &models.User{ID: id, Role: role}, issuedAt)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func do(h http.Handler, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAuth(t *testing.T) {
	var gotID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = UserID(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := Auth(testSecret)(next)

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"tanpa header", "", http.StatusUnauthorized},
		{"bukan Bearer", "Basic abc", http.StatusUnauthorized},
		{"token rusak", "Bearer rusak", http.StatusUnauthorized},
		{"token kedaluwarsa", "Bearer " + tokenFor(t, 7, models.RoleCustomer, time.Now().Add(-time.Hour)), http.StatusUnauthorized},
		{"token valid", "Bearer " + tokenFor(t, 7, models.RoleCustomer, time.Now()), http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID = 0
			if rec := do(h, tt.header); rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}

	gotID = 0
	do(h, "Bearer "+tokenFor(t, 7, models.RoleCustomer, time.Now()))
	if gotID != 7 {
		t.Errorf("UserID di context = %d, want 7", gotID)
	}
}

func TestRequireAdmin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := Chain(next, Auth(testSecret), RequireAdmin)

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"tanpa token", "", http.StatusUnauthorized},
		{"customer ditolak", "Bearer " + tokenFor(t, 1, models.RoleCustomer, time.Now()), http.StatusForbidden},
		{"admin lolos", "Bearer " + tokenFor(t, 2, models.RoleAdmin, time.Now()), http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(h, tt.header); rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

// RequireAdmin yang dipasang tanpa Auth tidak boleh meloloskan siapa pun.
func TestRequireAdminWithoutAuth(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	if rec := do(RequireAdmin(next), ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
