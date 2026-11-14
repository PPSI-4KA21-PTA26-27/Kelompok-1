package token

import (
	"errors"
	"testing"
	"time"
)

const (
	testSecret = "0123456789abcdef0123456789abcdef"
	testIssuer = "test-issuer"
)

func TestIssueAndParse(t *testing.T) {
	m := NewManager(testSecret, testIssuer, time.Minute)

	raw, err := m.Issue("user-1", "customer")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	claims, err := m.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Subject != "user-1" || claims.Role != "customer" {
		t.Errorf("claims = {sub:%q role:%q}, want {sub:user-1 role:customer}", claims.Subject, claims.Role)
	}
}

func TestParseRejects(t *testing.T) {
	verifier := NewManager(testSecret, testIssuer, time.Minute)

	tests := []struct {
		name   string
		issuer *Manager
	}{
		{"wrong secret", NewManager("another-secret-another-secret-1234", testIssuer, time.Minute)},
		{"wrong issuer", NewManager(testSecret, "other-issuer", time.Minute)},
		{"expired", NewManager(testSecret, testIssuer, -time.Minute)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := tc.issuer.Issue("user-1", "customer")
			if err != nil {
				t.Fatalf("Issue() error = %v", err)
			}
			if _, err := verifier.Parse(raw); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("Parse() error = %v, want ErrInvalidToken", err)
			}
		})
	}

	t.Run("garbage", func(t *testing.T) {
		if _, err := verifier.Parse("not-a-token"); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})
}
