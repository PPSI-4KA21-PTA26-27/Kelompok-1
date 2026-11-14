package auth

import (
	"slices"
	"strings"
	"testing"
)

func TestValidateRegister(t *testing.T) {
	phone := "+62 812-3456-7890"
	badPhone := "telepon"

	tests := []struct {
		name       string
		req        registerRequest
		wantFields []string
	}{
		{
			name: "valid with phone",
			req:  registerRequest{Name: "Lana", Email: "lana@example.com", Password: "password123", Phone: &phone},
		},
		{
			name: "valid without phone",
			req:  registerRequest{Name: "Lana", Email: "lana@example.com", Password: "password123"},
		},
		{
			name:       "everything invalid",
			req:        registerRequest{Name: "", Email: "bukan-email", Password: "pendek", Phone: &badPhone},
			wantFields: []string{"name", "email", "password", "phone"},
		},
		{
			name:       "password longer than bcrypt limit",
			req:        registerRequest{Name: "Lana", Email: "lana@example.com", Password: strings.Repeat("a", 73)},
			wantFields: []string{"password"},
		},
		{
			name:       "name too long",
			req:        registerRequest{Name: strings.Repeat("a", 101), Email: "lana@example.com", Password: "password123"},
			wantFields: []string{"name"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			for _, e := range validateRegister(tc.req) {
				got = append(got, e.Field)
			}
			if !slices.Equal(got, tc.wantFields) {
				t.Errorf("invalid fields = %v, want %v", got, tc.wantFields)
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name       string
		req        loginRequest
		wantFields []string
	}{
		{"valid", loginRequest{Email: "lana@example.com", Password: "x"}, nil},
		{"empty", loginRequest{}, []string{"email", "password"}},
		{"bad email", loginRequest{Email: "lana", Password: "x"}, []string{"email"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			for _, e := range validateLogin(tc.req) {
				got = append(got, e.Field)
			}
			if !slices.Equal(got, tc.wantFields) {
				t.Errorf("invalid fields = %v, want %v", got, tc.wantFields)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  Lana@Example.COM "); got != "lana@example.com" {
		t.Errorf("normalizeEmail() = %q, want lana@example.com", got)
	}
}
