package order

import (
	"regexp"
	"testing"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from, to string
		want     bool
	}{
		{StatusPendingPayment, StatusCancelled, true},
		{StatusPendingPayment, StatusPaid, false}, // hanya lewat verifikasi pembayaran
		{StatusPaid, StatusProcessing, true},
		{StatusProcessing, StatusShipped, true},
		{StatusShipped, StatusCompleted, true},
		{StatusShipped, StatusCancelled, false},
		{StatusCompleted, StatusCancelled, false},
		{StatusCancelled, StatusPaid, false},
	}
	for _, tt := range tests {
		if got := canTransition(tt.from, tt.to); got != tt.want {
			t.Errorf("canTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestNewOrderNumber(t *testing.T) {
	n, err := newOrderNumber()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^HS\d{14}-[0-9A-F]{6}$`).MatchString(n) {
		t.Errorf("format nomor order tidak sesuai: %q", n)
	}
}

func TestEtdLabel(t *testing.T) {
	for in, want := range map[string]string{"2 day": "2 hari", "1-2 days": "1-2 hari", "3": "3 hari", "": "-", "2 HARI": "2 hari"} {
		if got := etdLabel(in); got != want {
			t.Errorf("etdLabel(%q) = %q, want %q", in, got, want)
		}
	}
}
