package services

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"hii-stranger1/backend/models"
)

var testSecret = []byte("rahasia-untuk-test-minimal-32-karakter!")

func TestTokenRoundTrip(t *testing.T) {
	u := &models.User{ID: 42, Role: models.RoleAdmin}
	token, err := IssueToken(testSecret, 15*time.Minute, u, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("token valid ditolak: %v", err)
	}
	if got.UserID != 42 || got.Role != models.RoleAdmin {
		t.Errorf("claims = %+v, want id 42 role admin", got)
	}
}

func TestParseTokenRejects(t *testing.T) {
	u := &models.User{ID: 1, Role: models.RoleCustomer}

	expired, _ := IssueToken(testSecret, time.Minute, u, time.Now().Add(-time.Hour))
	valid, _ := IssueToken(testSecret, time.Minute, u, time.Now())

	// Token tanpa tanda tangan (alg none) harus ditolak.
	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "1", "role": "admin", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{"kedaluwarsa", expired},
		{"secret berbeda", func() string {
			s, _ := IssueToken([]byte("secret-lain-yang-juga-panjang-32-kar!"), time.Minute, u, time.Now())
			return s
		}()},
		{"alg none", none},
		{"isi diubah", valid[:len(valid)-2] + "xx"},
		{"bukan token", "bukan-token"},
		{"kosong", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseToken(testSecret, tt.token); err == nil {
				t.Error("token seharusnya ditolak")
			}
		})
	}
}

func TestHashToken(t *testing.T) {
	h := hashToken("abc")
	if len(h) != 64 {
		t.Errorf("panjang hash = %d, want 64 (cocok dengan CHAR(64) di database)", len(h))
	}
	if h == "abc" || strings.Contains(h, "abc") {
		t.Error("hash tidak boleh memuat token asli")
	}
	if hashToken("abc") != h {
		t.Error("hash harus konsisten untuk token yang sama")
	}
}

func TestNewResetToken(t *testing.T) {
	a, err := newResetToken()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := newResetToken()
	if len(a) != 64 || a == b {
		t.Errorf("token harus 64 hex dan unik, dapat %q dan %q", a, b)
	}
}
