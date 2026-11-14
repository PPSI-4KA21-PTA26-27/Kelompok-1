package cart

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/pkg/token"
)

const (
	validUUID   = "11111111-1111-4111-8111-111111111111"
	variantUUID = "22222222-2222-4222-8222-222222222222"
)

func intp(v int) *int { return &v }

// withCustomer membungkus fn dengan Authenticate dan mengembalikan header
// Authorization milik pelanggan. Handler dibuat tanpa service: semua kasus di
// bawah berhenti sebelum service dipanggil, jadi tidak butuh database.
func withCustomer(t *testing.T, fn http.HandlerFunc) (http.Handler, string) {
	t.Helper()
	tokens := token.NewManager("0123456789abcdef0123456789abcdef", "test-issuer", time.Minute)
	tok, err := tokens.Issue("u1", "customer")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return middleware.Chain(fn, middleware.Authenticate(tokens)), "Bearer " + tok
}

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name  string
		req   CreateRequest
		field string
	}{
		{"valid", CreateRequest{VariantID: variantUUID, Quantity: 1}, ""},
		{"variant bukan uuid", CreateRequest{VariantID: "abc", Quantity: 1}, "variant_id"},
		{"quantity nol", CreateRequest{VariantID: variantUUID, Quantity: 0}, "quantity"},
		{"quantity negatif", CreateRequest{VariantID: variantUUID, Quantity: -3}, "quantity"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateCreate(tc.req)
			if tc.field == "" {
				if len(errs) > 0 {
					t.Fatalf("validateCreate() = %v, want valid", errs)
				}
				return
			}
			if len(errs) == 0 || errs[0].Field != tc.field {
				t.Fatalf("validateCreate() = %v, want error on %s", errs, tc.field)
			}
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	if errs := validateUpdate(UpdateRequest{}); len(errs) != 1 || errs[0].Field != "quantity" {
		t.Errorf("quantity kosong: %v", errs)
	}
	if errs := validateUpdate(UpdateRequest{Quantity: intp(0)}); len(errs) != 1 {
		t.Errorf("quantity 0 harus ditolak: %v", errs)
	}
	if errs := validateUpdate(UpdateRequest{Quantity: intp(2)}); len(errs) != 0 {
		t.Errorf("quantity 2 harus valid: %v", errs)
	}
}

func TestHandlersRequirePrincipal(t *testing.T) {
	h := NewHandler(nil)
	for name, fn := range map[string]http.HandlerFunc{
		"get": h.Get, "create": h.Create, "update": h.Update, "delete": h.Delete, "clear": h.Clear,
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			fn(rec, httptest.NewRequest(http.MethodGet, "/cart", nil))
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestCreateGuards(t *testing.T) {
	handler, auth := withCustomer(t, NewHandler(nil).Create)
	tests := []struct {
		name, body string
		want       int
	}{
		{"json rusak", "{", http.StatusBadRequest},
		{"variant salah", `{"variant_id":"x","quantity":1}`, http.StatusUnprocessableEntity},
		{"quantity nol", fmt.Sprintf(`{"variant_id":%q,"quantity":0}`, variantUUID), http.StatusUnprocessableEntity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/cart/items", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", auth)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestItemIDGuards(t *testing.T) {
	h := NewHandler(nil)
	tests := []struct {
		name, id, body string
		fn             http.HandlerFunc
		want           int
	}{
		{"update id salah", "bukan-uuid", `{"quantity":1}`, h.Update, http.StatusNotFound},
		{"delete id salah", "bukan-uuid", "", h.Delete, http.StatusNotFound},
		{"update tanpa quantity", validUUID, `{}`, h.Update, http.StatusUnprocessableEntity},
		{"update quantity nol", validUUID, `{"quantity":0}`, h.Update, http.StatusUnprocessableEntity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler, auth := withCustomer(t, tc.fn)
			req := httptest.NewRequest(http.MethodPatch, "/cart/items/"+tc.id, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", auth)
			req.SetPathValue("cartItemId", tc.id)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		err  error
		want int
		code string
	}{
		{ErrNotFound, http.StatusNotFound, "not_found"},
		{ErrVariantNotFound, http.StatusUnprocessableEntity, "variant_id"},
		{ErrInvalidQuantity, http.StatusUnprocessableEntity, "quantity"},
		{fmt.Errorf("add: %w", ErrInsufficientStock), http.StatusConflict, "conflict"},
		{ErrUserNotFound, http.StatusUnauthorized, "unauthorized"},
		{errors.New("db mati"), http.StatusInternalServerError, ""},
	}
	h := NewHandler(nil)
	for _, tc := range tests {
		t.Run(tc.err.Error(), func(t *testing.T) {
			rec := httptest.NewRecorder()
			if !h.writeError(rec, httptest.NewRequest(http.MethodGet, "/", nil), "test", tc.err) {
				t.Fatal("writeError() = false")
			}
			if rec.Code != tc.want || !strings.Contains(rec.Body.String(), tc.code) {
				t.Errorf("status = %d body = %s, want %d %q", rec.Code, rec.Body.String(), tc.want, tc.code)
			}
		})
	}
	if h.writeError(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), "test", nil) {
		t.Error("writeError(nil) harus false")
	}
}

// Penjagaan Service berjalan sebelum repository disentuh, jadi repo nil aman.
func TestServiceGuards(t *testing.T) {
	s := NewService(nil)
	ctx := context.Background()
	tests := []struct {
		name string
		call func() error
		want error
	}{
		{"get tanpa user", func() error { _, err := s.Get(ctx, ""); return err }, ErrUserNotFound},
		{"add tanpa user", func() error { _, err := s.Add(ctx, "", CreateRequest{VariantID: variantUUID, Quantity: 1}); return err }, ErrUserNotFound},
		{"add tanpa variant", func() error { _, err := s.Add(ctx, "u1", CreateRequest{Quantity: 1}); return err }, ErrVariantNotFound},
		{"add quantity nol", func() error { _, err := s.Add(ctx, "u1", CreateRequest{VariantID: variantUUID}); return err }, ErrInvalidQuantity},
		{"add quantity melebihi int32", func() error {
			_, err := s.Add(ctx, "u1", CreateRequest{VariantID: variantUUID, Quantity: math.MaxInt32 + 1})
			return err
		}, ErrInsufficientStock},
		{"update tanpa id", func() error { _, err := s.UpdateQuantity(ctx, "u1", "", 1); return err }, ErrNotFound},
		{"update quantity nol", func() error { _, err := s.UpdateQuantity(ctx, "u1", validUUID, 0); return err }, ErrInvalidQuantity},
		{"update quantity melebihi int32", func() error { _, err := s.UpdateQuantity(ctx, "u1", validUUID, math.MaxInt32+1); return err }, ErrInsufficientStock},
		{"delete tanpa user", func() error { return s.Delete(ctx, "", validUUID) }, ErrUserNotFound},
		{"delete tanpa id", func() error { return s.Delete(ctx, "u1", "") }, ErrNotFound},
		{"clear tanpa user", func() error { return s.Clear(ctx, "") }, ErrUserNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}
