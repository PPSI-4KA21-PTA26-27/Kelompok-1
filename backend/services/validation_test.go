package services

import "testing"

func TestValidatePassword(t *testing.T) {
	long := make([]byte, 73)
	for i := range long {
		long[i] = 'a'
	}
	long[0] = '1'

	tests := []struct {
		name     string
		password string
		email    string
		wantErr  bool
	}{
		{"valid", "kaosputih2025", "a@b.com", false},
		{"terlalu pendek", "abc12345", "a@b.com", true},
		{"hanya huruf", "kaosputihkaos", "a@b.com", true},
		{"hanya angka", "1234567890", "a@b.com", true},
		{"sama dengan email", "arka2025@mail.com", "ARKA2025@mail.com", true},
		{"lebih dari 72 byte", string(long), "a@b.com", true},
		{"email kosong dilewati", "kaosputih2025", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validatePassword(tt.password, tt.email)
			if (got != "") != tt.wantErr {
				t.Errorf("validatePassword() = %q, wantErr %v", got, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"nama@email.com", false},
		{"", true},
		{"bukan-email", true},
		{"Nama <nama@email.com>", true},
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := validateEmail(tt.email); (got != "") != tt.wantErr {
				t.Errorf("validateEmail(%q) = %q, wantErr %v", tt.email, got, tt.wantErr)
			}
		})
	}
}

func TestPhone(t *testing.T) {
	tests := []struct {
		input   string
		want    string // hasil normalisasi
		wantErr bool
	}{
		{"", "", false}, // opsional
		{"0812 3456-7890", "081234567890", false},
		{"+6281234567890", "+6281234567890", false},
		{"12345", "12345", true},
		{"08abc", "08abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizePhone(tt.input)
			if got != tt.want {
				t.Errorf("normalizePhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
			if msg := validatePhone(got); (msg != "") != tt.wantErr {
				t.Errorf("validatePhone(%q) = %q, wantErr %v", got, msg, tt.wantErr)
			}
		})
	}
}

func TestValidateFullName(t *testing.T) {
	if validateFullName("") == "" {
		t.Error("nama kosong seharusnya ditolak")
	}
	if validateFullName("Arka Narendra") != "" {
		t.Error("nama biasa seharusnya diterima")
	}
	tooLong := ""
	for i := 0; i < 101; i++ {
		tooLong += "a"
	}
	if validateFullName(tooLong) == "" {
		t.Error("nama 101 karakter seharusnya ditolak")
	}
}

func TestCollect(t *testing.T) {
	if err := collect(map[string]string{"a": "", "b": ""}); err != nil {
		t.Errorf("semua valid seharusnya nil, dapat %v", err)
	}
	err := collect(map[string]string{"a": "", "b": "salah"})
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("seharusnya *ValidationError, dapat %T", err)
	}
	if len(ve.Fields) != 1 || ve.Fields["b"] != "salah" {
		t.Errorf("Fields = %v, hanya field yang salah yang boleh ikut", ve.Fields)
	}
}
