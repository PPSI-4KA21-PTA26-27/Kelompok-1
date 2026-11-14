package order

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/pkg/token"
)

const (
	itemA = "11111111-1111-4111-8111-111111111111"
	itemB = "22222222-2222-4222-8222-222222222222"
)

func validCheckout() createRequest {
	return createRequest{RecipientName: "Lana", RecipientPhone: "081234567890", ShippingAddress: "Jl. Mawar 1",
		ProvinceID: 1, CityID: 2, DistrictID: 3, SubdistrictID: 4}
}

func TestCheckoutInput(t *testing.T) {
	ids := make([]string, maxCartItemIDs+1)
	for i := range ids {
		ids[i] = fmt.Sprintf("00000000-0000-4000-8000-%012d", i)
	}
	tests := []struct {
		name   string
		mutate func(*createRequest)
		field  string
	}{
		{"valid", func(*createRequest) {}, ""},
		{"valid dengan item terpilih", func(r *createRequest) { r.CartItemIDs = []string{itemA, itemB} }, ""},
		{"nama kosong", func(r *createRequest) { r.RecipientName = "  " }, "recipient_name"},
		{"telepon pendek", func(r *createRequest) { r.RecipientPhone = "123" }, "recipient_phone"},
		{"alamat kosong", func(r *createRequest) { r.ShippingAddress = "" }, "shipping_address"},
		{"alamat terlalu panjang", func(r *createRequest) { r.ShippingAddress = strings.Repeat("a", 501) }, "shipping_address"},
		{"provinsi kosong", func(r *createRequest) { r.ProvinceID = 0 }, "province_id"},
		{"kota kosong", func(r *createRequest) { r.CityID = 0 }, "city_id"},
		{"kecamatan kosong", func(r *createRequest) { r.DistrictID = 0 }, "district_id"},
		{"kelurahan kosong", func(r *createRequest) { r.SubdistrictID = 0 }, "subdistrict_id"},
		{"item bukan uuid", func(r *createRequest) { r.CartItemIDs = []string{"x"} }, "cart_item_ids"},
		{"item ganda (beda huruf)", func(r *createRequest) { r.CartItemIDs = []string{itemA, strings.ToUpper(itemA)} }, "cart_item_ids"},
		{"item lebih dari 100", func(r *createRequest) { r.CartItemIDs = ids }, "cart_item_ids"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := validCheckout()
			tc.mutate(&req)
			in, errs := req.input()
			if tc.field == "" {
				if len(errs) > 0 {
					t.Fatalf("input() = %v, want valid", errs)
				}
				if in.RecipientName != "Lana" || in.Destination.SubdistrictID != 4 {
					t.Errorf("input() = %+v", in)
				}
				return
			}
			found := false
			for _, e := range errs {
				found = found || e.Field == tc.field
			}
			if !found {
				t.Errorf("input() = %v, want error pada %s", errs, tc.field)
			}
		})
	}
}

func TestCreateGuards(t *testing.T) {
	h := NewHandler(nil)
	rec := httptest.NewRecorder()
	h.Create(rec, httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader("{}")))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("tanpa login: status = %d, want 401", rec.Code)
	}

	tokens := token.NewManager("0123456789abcdef0123456789abcdef", "test-issuer", time.Minute)
	tok, _ := tokens.Issue("u1", "customer")
	handler := middleware.Chain(http.HandlerFunc(h.Create), middleware.Authenticate(tokens))
	for body, want := range map[string]int{"{": http.StatusBadRequest, "{}": http.StatusUnprocessableEntity} {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("body %q: status = %d, want %d", body, rec.Code, want)
		}
	}
}

func TestCancelGuards(t *testing.T) {
	h := NewHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/orders/x/cancel", nil)
	req.SetPathValue("orderId", "x")
	rec := httptest.NewRecorder()
	h.Cancel(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("id salah: status = %d, want 404", rec.Code)
	}
	req = httptest.NewRequest(http.MethodPost, "/orders/"+itemA+"/cancel", nil)
	req.SetPathValue("orderId", itemA)
	rec = httptest.NewRecorder()
	h.Cancel(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("tanpa login: status = %d, want 401", rec.Code)
	}
}

// Parameter days divalidasi sebelum service dipanggil; nilai valid (7/30/90)
// butuh database sehingga hanya nilai salah yang diuji di sini. Hak akses
// (owner saja) diuji di router (paket api).
func TestAdminAnalyticsDays(t *testing.T) {
	for _, days := range []string{"1", "14", "60", "365", "abc", "-7", "7.0"} {
		rec := httptest.NewRecorder()
		NewHandler(nil).AdminAnalytics(rec, httptest.NewRequest(http.MethodGet, "/admin/analytics?days="+days, nil))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("days=%s: status = %d, want 400", days, rec.Code)
		}
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		err  error
		want int
		code string
	}{
		{ErrNotFound, http.StatusNotFound, "not_found"},
		{fmt.Errorf("checkout: %w", ErrInsufficientStock), http.StatusConflict, "conflict"},
		{ErrCartChanged, http.StatusConflict, "conflict"},
		{ErrCartEmpty, http.StatusUnprocessableEntity, "cart_item_ids"},
		{ErrCartItemNotFound, http.StatusUnprocessableEntity, "cart_item_ids"},
		{ErrProductUnavailable, http.StatusConflict, "conflict"},
		{ErrTrackingRequired, http.StatusUnprocessableEntity, "tracking_number"},
		{ErrInvalidTransition, http.StatusConflict, "conflict"},
		{ErrEmailNotVerified, http.StatusForbidden, "email_not_verified"},
		{ErrNotCustomer, http.StatusForbidden, "forbidden"},
		{ErrProofUnderReview, http.StatusConflict, "conflict"},
		{errors.New("db mati"), http.StatusInternalServerError, ""},
	}
	h := NewHandler(nil)
	for _, tc := range tests {
		rec := httptest.NewRecorder()
		if !h.writeError(rec, httptest.NewRequest(http.MethodGet, "/", nil), "test", tc.err) {
			t.Fatalf("writeError(%v) = false", tc.err)
		}
		if rec.Code != tc.want || !strings.Contains(rec.Body.String(), tc.code) {
			t.Errorf("writeError(%v) = %d %s, want %d %q", tc.err, rec.Code, rec.Body.String(), tc.want, tc.code)
		}
	}
}
