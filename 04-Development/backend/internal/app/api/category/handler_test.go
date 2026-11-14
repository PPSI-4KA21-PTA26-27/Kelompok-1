package category

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/pkg/token"
)

const validID = "123e4567-e89b-12d3-a456-426614174000"

// Test handler di bawah hanya memeriksa penjagaan yang berjalan sebelum service
// dipanggil, jadi Handler dibuat tanpa service dan tidak butuh database.

func TestListGuards(t *testing.T) {
	tokens := token.NewManager("0123456789abcdef0123456789abcdef", "test-issuer", time.Minute)
	customer, err := tokens.Issue("u1", "customer")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	list := middleware.Chain(http.HandlerFunc(NewHandler(nil).List), middleware.AuthenticateOptional(tokens))

	tests := []struct {
		name          string
		query         string
		authorization string
		want          int
	}{
		{"page zero", "page=0", "", http.StatusBadRequest},
		{"limit over max", "limit=101", "", http.StatusBadRequest},
		{"include_inactive not a bool", "include_inactive=maybe", "", http.StatusBadRequest},
		{"include_inactive anonymous", "include_inactive=true", "", http.StatusForbidden},
		{"include_inactive customer", "include_inactive=true", "Bearer " + customer, http.StatusForbidden},
		{"invalid token is not treated as anonymous", "", "Bearer garbage", http.StatusUnauthorized},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/categories?"+tc.query, nil)
			if tc.authorization != "" {
				req.Header.Set("Authorization", tc.authorization)
			}
			rec := httptest.NewRecorder()
			list.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestPathIDMustBeUUID(t *testing.T) {
	h := NewHandler(nil)
	handlers := map[string]http.HandlerFunc{
		"Get":    h.Get,
		"Update": h.Update,
		"Delete": h.Delete,
	}
	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/categories/x", nil)
			req.SetPathValue("categoryId", "not-a-uuid")
			rec := httptest.NewRecorder()
			handler(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}

func TestCreateAndUpdateRejectBadBody(t *testing.T) {
	h := NewHandler(nil)

	tests := []struct {
		name    string
		handler http.HandlerFunc
		body    string
		want    int
	}{
		{"create broken json", h.Create, `{bad`, http.StatusBadRequest},
		{"create blank name", h.Create, `{"name":"   "}`, http.StatusUnprocessableEntity},
		{"create missing name", h.Create, `{}`, http.StatusUnprocessableEntity},
		{"create name too long", h.Create, `{"name":"` + strings.Repeat("a", 101) + `"}`, http.StatusUnprocessableEntity},
		{"update broken json", h.Update, `{bad`, http.StatusBadRequest},
		{"update empty body", h.Update, `{}`, http.StatusUnprocessableEntity},
		{"update blank name", h.Update, `{"name":""}`, http.StatusUnprocessableEntity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(tc.body))
			req.SetPathValue("categoryId", validID)
			rec := httptest.NewRecorder()
			tc.handler(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestUpdateRequestChanges(t *testing.T) {
	str := func(s string) *string { return &s }
	yes, no := true, false

	tests := []struct {
		name         string
		req          updateRequest
		wantName     *string
		wantIsActive *bool
		wantFields   []string
	}{
		{name: "name only, trimmed", req: updateRequest{Name: str("  Baju  ")}, wantName: str("Baju")},
		{name: "deactivate", req: updateRequest{IsActive: &no}, wantIsActive: &no},
		{name: "both", req: updateRequest{Name: str("Celana"), IsActive: &yes}, wantName: str("Celana"), wantIsActive: &yes},
		{name: "nothing", req: updateRequest{}, wantFields: []string{"body"}},
		{name: "blank name", req: updateRequest{Name: str("  ")}, wantName: str(""), wantFields: []string{"name"}},
		{
			name: "name too long", req: updateRequest{Name: str(strings.Repeat("a", 101))},
			wantName: str(strings.Repeat("a", 101)), wantFields: []string{"name"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, errs := tc.req.changes()

			var gotFields []string
			for _, e := range errs {
				gotFields = append(gotFields, e.Field)
			}
			if !slices.Equal(gotFields, tc.wantFields) {
				t.Errorf("invalid fields = %v, want %v", gotFields, tc.wantFields)
			}
			if !equal(got.Name, tc.wantName) {
				t.Errorf("Name = %v, want %v", got.Name, tc.wantName)
			}
			if !equal(got.IsActive, tc.wantIsActive) {
				t.Errorf("IsActive = %v, want %v", got.IsActive, tc.wantIsActive)
			}
		})
	}
}

func equal[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
