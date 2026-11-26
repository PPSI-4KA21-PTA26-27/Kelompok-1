package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hii-stranger1/backend/internal/app/api/auth"
	"hii-stranger1/backend/internal/app/api/banner"
	"hii-stranger1/backend/internal/app/api/cart"
	"hii-stranger1/backend/internal/app/api/order"
	"hii-stranger1/backend/internal/app/api/payment"
	"hii-stranger1/backend/internal/app/api/shipment"
	"hii-stranger1/backend/internal/pkg/config"
	"hii-stranger1/backend/internal/pkg/token"
)

const someUUID = "11111111-1111-4111-8111-111111111111"

// testRouter menyusun router lengkap dengan handler tanpa service. Kasus uji
// hanya memakai request yang ditolak middleware (401/403/429) atau oleh
// validasi handler sebelum service dipanggil (400/422), jadi tidak butuh
// database maupun Redis.
func testRouter(t *testing.T) (http.Handler, func(role string) string) {
	t.Helper()
	tokens := token.NewManager("0123456789abcdef0123456789abcdef", "test-issuer", time.Minute)
	h := handlers{
		auth:     auth.NewHandler(nil),
		banner:   banner.NewHandler(nil),
		cart:     cart.NewHandler(nil),
		order:    order.NewHandler(nil),
		payment:  payment.NewHandler(nil),
		shipment: shipment.NewHandler(nil),
		// Peran dari id token ("u-<role>"); role "inactive" mewakili akun nonaktif.
		account: func(_ context.Context, id string) (string, bool, error) {
			role := strings.TrimPrefix(id, "u-")
			return role, role != "inactive", nil
		},
	}
	cfg := config.Config{MaxBodyBytes: 1 << 20, HTTPRequestTimeout: 5 * time.Second}
	bearer := func(role string) string {
		if role == "" {
			return ""
		}
		tok, err := tokens.Issue("u-"+role, role)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}
		return "Bearer " + tok
	}
	return newRouter(cfg, tokens, h, nil), bearer
}

func TestRouteRoleAccess(t *testing.T) {
	router, bearer := testRouter(t)
	tests := []struct {
		method, path, body, role string
		want                     int
	}{
		// DELETE /cart dan item keranjang khusus pelanggan.
		{http.MethodDelete, "/api/v1/cart", "", "", http.StatusUnauthorized},
		{http.MethodDelete, "/api/v1/cart", "", "staff", http.StatusForbidden},
		{http.MethodDelete, "/api/v1/cart", "", "owner", http.StatusForbidden},
		{http.MethodGet, "/api/v1/cart", "", "owner", http.StatusForbidden},
		{http.MethodPost, "/api/v1/cart/items", "{}", "staff", http.StatusForbidden},
		{http.MethodPost, "/api/v1/cart/items", "{}", "customer", http.StatusUnprocessableEntity},
		{http.MethodPatch, "/api/v1/cart/items/" + someUUID, "{}", "owner", http.StatusForbidden},

		// Akun nonaktif ditolak walau token masih berlaku (SDD 3.5).
		{http.MethodPost, "/api/v1/orders", "{}", "inactive", http.StatusUnauthorized},

		// Checkout dan unggah bukti khusus pelanggan.
		{http.MethodPost, "/api/v1/orders", "{}", "staff", http.StatusForbidden},
		{http.MethodPost, "/api/v1/orders", "{}", "customer", http.StatusUnprocessableEntity},
		{http.MethodPut, "/api/v1/orders/" + someUUID + "/payment/proof", "", "staff", http.StatusForbidden},

		// Review pembayaran oleh owner/staff.
		{http.MethodPost, "/api/v1/admin/payments/" + someUUID + "/review", "{}", "customer", http.StatusForbidden},
		{http.MethodPost, "/api/v1/admin/payments/" + someUUID + "/review", "{}", "staff", http.StatusUnprocessableEntity},

		// Upsert pengiriman oleh owner/staff.
		{http.MethodPut, "/api/v1/orders/" + someUUID + "/shipment", "{}", "", http.StatusUnauthorized},
		{http.MethodPut, "/api/v1/orders/" + someUUID + "/shipment", "{}", "customer", http.StatusForbidden},
		{http.MethodPut, "/api/v1/orders/" + someUUID + "/shipment", "{}", "staff", http.StatusUnprocessableEntity},
		{http.MethodPut, "/api/v1/orders/" + someUUID + "/shipment", "{}", "owner", http.StatusUnprocessableEntity},

		// Banner (UC-18): tulis hanya staff/owner; gambar wajib saat membuat.
		{http.MethodPost, "/api/v1/banners", "{}", "", http.StatusUnauthorized},
		{http.MethodPost, "/api/v1/banners", "{}", "customer", http.StatusForbidden},
		{http.MethodPost, "/api/v1/banners", "{}", "staff", http.StatusUnprocessableEntity},
		{http.MethodPatch, "/api/v1/banners/" + someUUID, "{}", "customer", http.StatusForbidden},
		{http.MethodDelete, "/api/v1/banners/" + someUUID, "", "customer", http.StatusForbidden},

		// Ringkasan dashboard (L-18): staff dan owner.
		{http.MethodGet, "/api/v1/admin/summary", "", "", http.StatusUnauthorized},
		{http.MethodGet, "/api/v1/admin/summary", "", "customer", http.StatusForbidden},

		// Laporan pendapatan (UC-19): owner saja; parameter salah ditolak 400.
		{http.MethodGet, "/api/v1/admin/analytics?days=7", "", "", http.StatusUnauthorized},
		{http.MethodGet, "/api/v1/admin/analytics?days=7", "", "customer", http.StatusForbidden},
		{http.MethodGet, "/api/v1/admin/analytics?days=7", "", "staff", http.StatusForbidden},
		{http.MethodGet, "/api/v1/admin/analytics?days=5", "", "owner", http.StatusBadRequest},
		{http.MethodGet, "/api/v1/admin/analytics?month=2026-13", "", "owner", http.StatusBadRequest},
	}
	for _, tc := range tests {
		name := tc.method + " " + tc.path + " as " + tc.role
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			if a := bearer(tc.role); a != "" {
				req.Header.Set("Authorization", a)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}
