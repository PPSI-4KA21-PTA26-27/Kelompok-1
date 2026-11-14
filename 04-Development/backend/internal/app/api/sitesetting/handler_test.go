package sitesetting

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Validasi Update berjalan sebelum service dipanggil, jadi Handler tanpa
// service tidak butuh database. Hak akses (owner saja) diuji di router.
func TestUpdateValidation(t *testing.T) {
	tests := []struct {
		name, body string
		want       int
		field      string
	}{
		{"json rusak", "{", http.StatusBadRequest, ""},
		{"body kosong", `{}`, http.StatusUnprocessableEntity, "body"},
		{"pesan kosong", `{"maintenance_message":"   "}`, http.StatusUnprocessableEntity, "maintenance_message"},
		{"pesan terlalu panjang", `{"maintenance_message":"` + strings.Repeat("a", 501) + `"}`, http.StatusUnprocessableEntity, "maintenance_message"},
		{"qris bukan url", `{"qris_image_url":"ftp://x/qr.png"}`, http.StatusUnprocessableEntity, "qris_image_url"},
		{"merchant terlalu panjang", `{"qris_merchant_name":"` + strings.Repeat("m", 101) + `"}`, http.StatusUnprocessableEntity, "qris_merchant_name"},
		{"nmid terlalu panjang", `{"qris_nmid":"` + strings.Repeat("9", 51) + `"}`, http.StatusUnprocessableEntity, "qris_nmid"},
		{"jendela bayar nol", `{"payment_window_hours":0}`, http.StatusUnprocessableEntity, "payment_window_hours"},
		{"jendela bayar lebih 168", `{"payment_window_hours":169}`, http.StatusUnprocessableEntity, "payment_window_hours"},
	}
	h := NewHandler(nil)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/site-settings", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.Update(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
			if tc.field != "" && !strings.Contains(rec.Body.String(), `"field":"`+tc.field+`"`) {
				t.Errorf("body = %s, want error pada %s", rec.Body.String(), tc.field)
			}
		})
	}
}
