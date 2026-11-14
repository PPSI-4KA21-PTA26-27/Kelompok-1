package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Batas token 20-200 byte dan password 8-72 byte (batas bcrypt), SDD lupa sandi.
func TestValidateResetPasswordBoundaries(t *testing.T) {
	tests := []struct {
		name            string
		tokenLen, pwLen int
		field           string
	}{
		{"token 20 password 8", 20, 8, ""},
		{"token 200 password 72", 200, 72, ""},
		{"token 19", 19, 8, "token"},
		{"token 201", 201, 8, "token"},
		{"password 7", 20, 7, "password"},
		{"password 73", 20, 73, "password"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateResetPassword(resetPasswordRequest{
				Token: strings.Repeat("t", tc.tokenLen), Password: strings.Repeat("p", tc.pwLen),
			})
			if tc.field == "" {
				if len(errs) > 0 {
					t.Fatalf("errs = %v, want valid", errs)
				}
				return
			}
			if len(errs) != 1 || errs[0].Field != tc.field {
				t.Errorf("errs = %v, want satu error pada %s", errs, tc.field)
			}
		})
	}
}

func post(fn http.HandlerFunc, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/auth/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	fn(rec, req)
	return rec
}

// Kasus di bawah berhenti sebelum service dipanggil (tanpa database). Jalur
// 204 (email terdaftar/tidak) dan 410 reset_invalid butuh user.Repository
// sungguhan, jadi tidak diuji di sini.
func TestForgotPasswordGuards(t *testing.T) {
	h := NewHandler(nil)
	tests := []struct {
		name, body string
		want       int
	}{
		{"json rusak", "{", http.StatusBadRequest},
		{"email kosong", `{}`, http.StatusUnprocessableEntity},
		{"email salah", `{"email":"bukan-email"}`, http.StatusUnprocessableEntity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if rec := post(h.ForgotPassword, tc.body); rec.Code != tc.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestResetPasswordGuards(t *testing.T) {
	h := NewHandler(nil)
	tok := strings.Repeat("a", 32)
	tests := []struct {
		name, body string
		want       int
		field      string
	}{
		{"json rusak", "{", http.StatusBadRequest, ""},
		{"token pendek", `{"token":"pendek","password":"rahasia123"}`, http.StatusUnprocessableEntity, "token"},
		{"token hanya spasi", fmt.Sprintf(`{"token":%q,"password":"rahasia123"}`, strings.Repeat(" ", 30)), http.StatusUnprocessableEntity, "token"},
		{"token terlalu panjang", fmt.Sprintf(`{"token":%q,"password":"rahasia123"}`, strings.Repeat("a", 201)), http.StatusUnprocessableEntity, "token"},
		{"password pendek", fmt.Sprintf(`{"token":%q,"password":"pendek"}`, tok), http.StatusUnprocessableEntity, "password"},
		{"password 73 byte", fmt.Sprintf(`{"token":%q,"password":%q}`, tok, strings.Repeat("p", 73)), http.StatusUnprocessableEntity, "password"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := post(h.ResetPassword, tc.body)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
			if tc.field != "" && !strings.Contains(rec.Body.String(), `"field":"`+tc.field+`"`) {
				t.Errorf("body = %s, want error pada %s", rec.Body.String(), tc.field)
			}
		})
	}
}
