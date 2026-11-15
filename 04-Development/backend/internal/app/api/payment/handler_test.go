package payment

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/pkg/token"
)

const (
	orderUUID   = "11111111-1111-4111-8111-111111111111"
	paymentUUID = "22222222-2222-4222-8222-222222222222"
)

// authed membungkus fn dengan Authenticate dan mengembalikan header
// Authorization untuk role tersebut. Semua kasus di bawah berhenti sebelum
// service dipanggil, jadi Handler dibuat tanpa service (tanpa database).
func authed(t *testing.T, role string, fn http.HandlerFunc) (http.Handler, string) {
	t.Helper()
	tokens := token.NewManager("0123456789abcdef0123456789abcdef", "test-issuer", time.Minute)
	tok, err := tokens.Issue("u1", role)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return middleware.Chain(fn, middleware.Authenticate(tokens)), "Bearer " + tok
}

func serve(h http.Handler, auth string, req *http.Request) *httptest.ResponseRecorder {
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestOrderRequestGuards(t *testing.T) {
	h := NewHandler(nil)
	for name, fn := range map[string]http.HandlerFunc{
		"get": h.Get, "create": h.Create, "upload": h.UploadProof, "proof": h.GetProof,
	} {
		t.Run(name+" id salah", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/orders/x/payment", nil)
			req.SetPathValue("orderId", "x")
			rec := httptest.NewRecorder()
			fn(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want 404", rec.Code)
			}
		})
		t.Run(name+" tanpa login", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/orders/"+orderUUID+"/payment", nil)
			req.SetPathValue("orderId", orderUUID)
			rec := httptest.NewRecorder()
			fn(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestCreateRejectsNonQRIS(t *testing.T) {
	handler, auth := authed(t, "customer", NewHandler(nil).Create)
	req := httptest.NewRequest(http.MethodPost, "/orders/"+orderUUID+"/payment", strings.NewReader(`{"method":"transfer"}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("orderId", orderUUID)
	rec := serve(handler, auth, req)
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), `"field":"method"`) {
		t.Errorf("status = %d body = %s, want 422 pada method", rec.Code, rec.Body.String())
	}
}

func proofRequest(t *testing.T, field string, data []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile(field, "bukti.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(data)
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPut, "/orders/"+orderUUID+"/payment/proof", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.SetPathValue("orderId", orderUUID)
	return req
}

func TestUploadProofFileChecks(t *testing.T) {
	png := func(n int) []byte { b := make([]byte, n); copy(b, "\x89PNG\r\n\x1a\n"); return b }
	tests := []struct {
		name  string
		field string
		data  []byte
		want  int
	}{
		{"lebih dari 2 MB", "file", png(maxProofBytes + 1), http.StatusRequestEntityTooLarge},
		{"bukan gambar/pdf", "file", []byte("<html><body>bukan bukti</body></html>"), http.StatusUnsupportedMediaType},
		{"berkas kosong", "file", nil, http.StatusBadRequest},
		{"field salah", "proof", png(100), http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler, auth := authed(t, "customer", NewHandler(nil).UploadProof)
			rec := serve(handler, auth, proofRequest(t, tc.field, tc.data))
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestProofTypesIncludePDF(t *testing.T) {
	want := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "application/pdf": true}
	if len(proofTypes) != len(want) {
		t.Fatalf("proofTypes = %v", proofTypes)
	}
	for _, typ := range proofTypes {
		if !want[typ] {
			t.Errorf("tipe tak terduga %q", typ)
		}
	}
	if maxProofBytes != 2<<20 {
		t.Errorf("maxProofBytes = %d, want 2 MB", maxProofBytes)
	}
}

func TestAdminReviewValidation(t *testing.T) {
	tests := []struct {
		name, id, body string
		want           int
	}{
		{"id salah", "x", `{"status":"valid"}`, http.StatusNotFound},
		{"json rusak", paymentUUID, "{", http.StatusBadRequest},
		{"status pending", paymentUUID, `{"status":"menunggu"}`, http.StatusUnprocessableEntity},
		{"status kosong", paymentUUID, `{}`, http.StatusUnprocessableEntity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler, auth := authed(t, "staff", NewHandler(nil).AdminReview)
			req := httptest.NewRequest(http.MethodPost, "/admin/payments/"+tc.id+"/review", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("paymentId", tc.id)
			if rec := serve(handler, auth, req); rec.Code != tc.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestAdminListGuards(t *testing.T) {
	for name, query := range map[string]string{
		"status salah":   "status=paid",
		"order_id salah": "order_id=123",
		"q kepanjangan":  "q=" + strings.Repeat("a", 101),
		"page nol":       "page=0",
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewHandler(nil).AdminList(rec, httptest.NewRequest(http.MethodGet, "/admin/payments?"+query, nil))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{ErrNotFound, http.StatusNotFound},
		{ErrOrderNotFound, http.StatusNotFound},
		{ErrAlreadyExists, http.StatusConflict},
		{ErrOrderNotPayable, http.StatusConflict},
		{fmt.Errorf("upload: %w", ErrAlreadyApproved), http.StatusConflict},
		{ErrNoProof, http.StatusConflict},
		{ErrAlreadyReviewed, http.StatusConflict},
		{ErrStorageDisabled, http.StatusServiceUnavailable},
		{errors.New("db mati"), http.StatusInternalServerError},
	}
	h := NewHandler(nil)
	for _, tc := range tests {
		rec := httptest.NewRecorder()
		if !h.writeError(rec, httptest.NewRequest(http.MethodGet, "/", nil), "test", tc.err) || rec.Code != tc.want {
			t.Errorf("writeError(%v) status = %d, want %d", tc.err, rec.Code, tc.want)
		}
	}
}
