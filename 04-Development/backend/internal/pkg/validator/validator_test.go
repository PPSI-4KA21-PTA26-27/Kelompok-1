package validator

import "testing"

func TestEmail(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid", "user@example.com", true},
		{"valid short tld", "a@b.co", true},
		{"empty", "", false},
		{"no at sign", "userexample.com", false},
		{"domain without dot", "user@localhost", false},
		{"display name", "Bob <bob@example.com>", false},
		{"contains space", "us er@example.com", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Email(tc.in); got != tc.want {
				t.Errorf("Email(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestPhone(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"international", "+62 812-3456-7890", true},
		{"digits only", "081234567890", true},
		{"empty", "", false},
		{"letters", "abc123", false},
		{"no digits", "+-()", false},
		{"too long", "123456789012345678901", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Phone(tc.in); got != tc.want {
				t.Errorf("Phone(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"lowercase", "123e4567-e89b-12d3-a456-426614174000", true},
		{"uppercase", "123E4567-E89B-12D3-A456-426614174000", true},
		{"empty", "", false},
		{"not a uuid", "me", false},
		{"too short", "123e4567-e89b-12d3-a456-42661417400", false},
		{"too long", "123e4567-e89b-12d3-a456-4266141740000", false},
		{"missing hyphens", "123e4567e89b12d3a456426614174000abcd", false},
		{"non hex digit", "123e4567-e89b-12d3-a456-42661417400g", false},
		{"sql injection", "1' OR '1'='1' --xxxxxxxxxxxxxxxxxxxxxx", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := UUID(tc.in); got != tc.want {
				t.Errorf("UUID(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestCollector(t *testing.T) {
	var c Collector
	c.Check(true, "ok", "tidak dicatat")
	c.Check(false, "name", "wajib diisi")

	errs := c.Errors()
	if len(errs) != 1 || errs[0].Field != "name" || errs[0].Message != "wajib diisi" {
		t.Errorf("Errors() = %+v, want one error for field name", errs)
	}
}
