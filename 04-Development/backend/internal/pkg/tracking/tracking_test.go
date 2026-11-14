package tracking

import "testing"

func TestLinks(t *testing.T) {
	tests := []struct {
		courier, resi string
		want          int
		first         string
	}{
		{"POS Reguler", "P123", 2, "https://www.posindonesia.co.id/id/tracking"},
		{"jne REG", "JN1", 2, "https://www.jne.co.id/tracking-package"},
		{"Kurir toko", "ABC", 1, UniversalURL("ABC")},
		{"JNE", "  ", 0, ""},
	}
	for _, tc := range tests {
		got := Links(tc.courier, tc.resi)
		if len(got) != tc.want {
			t.Fatalf("Links(%q,%q) len = %d, want %d", tc.courier, tc.resi, len(got), tc.want)
		}
		if tc.want > 0 && got[0].URL != tc.first {
			t.Errorf("Links(%q) first = %q, want %q", tc.courier, got[0].URL, tc.first)
		}
	}
	if u := UniversalURL("A B"); u != "https://cekresi.com/?noresi=A+B" {
		t.Errorf("UniversalURL escape = %q", u)
	}
}
