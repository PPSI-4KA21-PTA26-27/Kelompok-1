package shipment

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const orderUUID = "11111111-1111-4111-8111-111111111111"

func hasField(body, field string) bool {
	return strings.Contains(body, `"field":"`+field+`"`)
}

// Validasi Upsert berjalan sebelum service dipanggil, jadi Handler tanpa
// service cukup. Hak akses (owner/staff saja) diuji di router (paket api).
func TestUpsertValidation(t *testing.T) {
	long := strings.Repeat("R", 101)
	tests := []struct {
		name, id, body string
		want           int
		field          string
	}{
		{"order id salah", "x", `{}`, http.StatusNotFound, ""},
		{"json rusak", orderUUID, "{", http.StatusBadRequest, ""},
		{"coverage kosong", orderUUID, `{"shipping_status":"pending"}`, http.StatusUnprocessableEntity, "coverage_status"},
		{"coverage salah", orderUUID, `{"coverage_status":"partial","shipping_status":"pending"}`, http.StatusUnprocessableEntity, "coverage_status"},
		{"coverage huruf besar", orderUUID, `{"coverage_status":"COVERED","shipping_status":"pending"}`, http.StatusUnprocessableEntity, "coverage_status"},
		{"shipping salah", orderUUID, `{"coverage_status":"covered","shipping_status":"lost"}`, http.StatusUnprocessableEntity, "shipping_status"},
		{"resi terlalu panjang", orderUUID, fmt.Sprintf(`{"coverage_status":"not_covered","shipping_status":"shipped","tracking_number":%q}`, long),
			http.StatusUnprocessableEntity, "tracking_number"},
	}
	h := NewHandler(nil)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/orders/"+tc.id+"/shipment", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("orderId", tc.id)
			rec := httptest.NewRecorder()
			h.Upsert(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
			if tc.field != "" && !hasField(rec.Body.String(), tc.field) {
				t.Errorf("body = %s, want error pada %s", rec.Body.String(), tc.field)
			}
		})
	}
}

func TestGetGuards(t *testing.T) {
	h := NewHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/orders/x/shipment", nil)
	req.SetPathValue("orderId", "x")
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("id salah: status = %d, want 404", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/orders/"+orderUUID+"/shipment", nil)
	req.SetPathValue("orderId", orderUUID)
	rec = httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("tanpa login: status = %d, want 401", rec.Code)
	}
}

func TestWriteError(t *testing.T) {
	h := NewHandler(nil)
	for err, want := range map[error]int{
		ErrNotFound:                           http.StatusNotFound,
		fmt.Errorf("x: %w", ErrOrderNotFound): http.StatusNotFound,
		errors.New("db mati"):                 http.StatusInternalServerError,
	} {
		rec := httptest.NewRecorder()
		if !h.writeError(rec, httptest.NewRequest(http.MethodGet, "/", nil), "test", err) || rec.Code != want {
			t.Errorf("writeError(%v) status = %d, want %d", err, rec.Code, want)
		}
	}
	if h.writeError(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), "test", nil) {
		t.Error("writeError(nil) harus false")
	}
}

func TestValidShippingAndLabel(t *testing.T) {
	for _, s := range []string{ShippingPending, ShippingShipped, ShippingDelivered, ShippingReturned} {
		if !validShipping(s) {
			t.Errorf("validShipping(%q) = false", s)
		}
		if shippingLabel(s) == s {
			t.Errorf("shippingLabel(%q) belum diterjemahkan", s)
		}
	}
	if validShipping("") || validShipping("lost") {
		t.Error("status tak dikenal harus ditolak")
	}
	if shippingLabel("lost") != "lost" {
		t.Error("status tak dikenal dikembalikan apa adanya")
	}
}
