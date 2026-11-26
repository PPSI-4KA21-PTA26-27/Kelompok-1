package product

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/token"
)

func intp(v int) *int           { return &v }
func strp(v string) *string     { return &v }
func floatp(v float64) *float64 { return &v }

func validCreate() createRequest {
	return createRequest{Name: "Kaos", Price: "150000", WeightGram: 200,
		AvailabilityType: AvailabilityReady, Status: StatusActive}
}

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*createRequest)
		field  string // kosong berarti harus valid
	}{
		{"valid", func(*createRequest) {}, ""},
		{"nama kosong", func(r *createRequest) { r.Name = "" }, "name"},
		{"harga negatif", func(r *createRequest) { r.Price = "-1" }, "price"},
		{"harga bukan angka", func(r *createRequest) { r.Price = "abc" }, "price"},
		{"diskon lebih dari 100", func(r *createRequest) { r.DiscountPercent = floatp(101) }, "discount_percent"},
		{"berat nol", func(r *createRequest) { r.WeightGram = 0 }, "weight_gram"},
		{"availability salah", func(r *createRequest) { r.AvailabilityType = "x" }, "availability_type"},
		{"status salah", func(r *createRequest) { r.Status = "x" }, "status"},
		{"category bukan uuid", func(r *createRequest) { r.CategoryID = strp("1") }, "category_id"},
		{"preorder tanpa estimasi", func(r *createRequest) { r.AvailabilityType = AvailabilityPreorder }, "po_estimate_days"},
		{"preorder estimasi 0 hari", func(r *createRequest) {
			zero := 0
			r.AvailabilityType, r.PoEstimateDays = AvailabilityPreorder, &zero
		}, "po_estimate_days"},
		{"preorder dengan estimasi", func(r *createRequest) {
			r.AvailabilityType = AvailabilityPreorder
			r.PoEstimateDays = intp(7)
		}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := validCreate()
			tc.mutate(&req)
			errs := validateCreate(req)
			if tc.field == "" {
				if len(errs) > 0 {
					t.Fatalf("validateCreate() = %v, want valid", errs)
				}
				return
			}
			if !slices.ContainsFunc(errs, func(e response.FieldError) bool { return e.Field == tc.field }) {
				t.Fatalf("validateCreate() = %v, want error on %s", errs, tc.field)
			}
		})
	}
}

func TestUpdateChanges(t *testing.T) {
	if _, errs := (updateRequest{}).changes(); len(errs) == 0 {
		t.Error("body kosong harus ditolak")
	}
	if _, errs := (updateRequest{Price: strp("x")}).changes(); len(errs) == 0 {
		t.Error("harga salah harus ditolak")
	}
	ch, errs := (updateRequest{Name: strp("  Baru  ")}).changes()
	if len(errs) > 0 || ch.Name == nil || *ch.Name != "Baru" {
		t.Errorf("changes() = %+v, %v; want nama di-trim", ch, errs)
	}
}

// Penjagaan di bawah berjalan sebelum service dipanggil, jadi Handler dibuat
// tanpa service dan tidak butuh database.
func TestListGuards(t *testing.T) {
	tokens := token.NewManager("0123456789abcdef0123456789abcdef", "test-issuer", time.Minute)
	customer, err := tokens.Issue("u1", "customer")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	list := middleware.Chain(http.HandlerFunc(NewHandler(nil).List), middleware.AuthenticateOptional(tokens))

	tests := []struct {
		name, query, authorization string
		want                       int
	}{
		{"page nol", "page=0", "", http.StatusBadRequest},
		{"sort salah", "sort=color", "", http.StatusBadRequest},
		{"category_id salah", "category_id=1", "", http.StatusBadRequest},
		{"min_price salah", "min_price=-5", "", http.StatusBadRequest},
		{"status anonim", "status=inactive", "", http.StatusForbidden},
		{"status customer", "status=inactive", "Bearer " + customer, http.StatusForbidden},
		{"token rusak", "", "Bearer garbage", http.StatusUnauthorized},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products?"+tc.query, nil)
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

func TestIDGuards(t *testing.T) {
	h := NewHandler(nil)
	for name, fn := range map[string]http.HandlerFunc{"get": h.Get, "update": h.Update, "delete": h.Delete} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products/x", nil)
			req.SetPathValue("productId", "bukan-uuid")
			rec := httptest.NewRecorder()
			fn(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want 404", rec.Code)
			}
		})
	}
}
