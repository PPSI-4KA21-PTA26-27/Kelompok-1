package auth

import (
	"strings"
	"testing"
)

func TestValidateResetPassword(t *testing.T) {
	token := strings.Repeat("a", 43)
	tests := []struct {
		name string
		req  resetPasswordRequest
		want []string
	}{
		{"valid", resetPasswordRequest{Token: token, Password: "rahasia123"}, nil},
		{"short password", resetPasswordRequest{Token: token, Password: "pendek"}, []string{"password"}},
		{"password over bcrypt limit", resetPasswordRequest{Token: token, Password: strings.Repeat("x", 73)}, []string{"password"}},
		{"missing token", resetPasswordRequest{Password: "rahasia123"}, []string{"token"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateResetPassword(tt.req)
			if len(errs) != len(tt.want) {
				t.Fatalf("got %d errors (%v), want %v", len(errs), errs, tt.want)
			}
			for i, f := range tt.want {
				if errs[i].Field != f {
					t.Fatalf("field %d = %q, want %q", i, errs[i].Field, f)
				}
			}
		})
	}
}
