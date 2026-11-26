package order

import (
	"testing"
	"time"
)

func TestAnalyticsRange(t *testing.T) {
	// 23.30 WIB tanggal 7 masih dihitung hari tanggal 7.
	now := time.Date(2026, 11, 7, 16, 30, 0, 0, time.UTC)
	prev, start, end := analyticsRange(now, 7)
	if got := end.Format(time.RFC3339); got != "2026-11-08T00:00:00+07:00" {
		t.Fatalf("end = %s", got)
	}
	if got := start.Format("2006-01-02"); got != "2026-11-01" {
		t.Fatalf("start = %s", got)
	}
	if got := prev.Format("2006-01-02"); got != "2026-10-25" {
		t.Fatalf("prev = %s", got)
	}
}

func TestMonthRangeBoundaries(t *testing.T) {
	cases := []struct {
		in               time.Time
		prev, start, end string
	}{
		{time.Date(2026, 1, 31, 23, 59, 0, 0, wib), "2025-12-01", "2026-01-01", "2026-02-01"},
		{time.Date(2028, 2, 10, 0, 0, 0, 0, wib), "2028-01-01", "2028-02-01", "2028-03-01"}, // kabisat: 29 hari
		{time.Date(2026, 12, 31, 0, 0, 0, 0, wib), "2026-11-01", "2026-12-01", "2027-01-01"},
	}
	for _, c := range cases {
		prev, start, end := monthRange(c.in)
		if prev.Format("2006-01-02") != c.prev || start.Format("2006-01-02") != c.start || end.Format("2006-01-02") != c.end {
			t.Fatalf("%v: got %s %s %s", c.in, prev.Format("2006-01-02"), start.Format("2006-01-02"), end.Format("2006-01-02"))
		}
	}
	// end eksklusif: tepat 1 Maret 00:00 masuk bulan Maret, bukan Februari.
	_, _, end := monthRange(time.Date(2028, 2, 1, 0, 0, 0, 0, wib))
	if !end.Equal(time.Date(2028, 3, 1, 0, 0, 0, 0, wib)) {
		t.Fatalf("end = %v", end)
	}
}

func TestMonthRange(t *testing.T) {
	prev, start, end := monthRange(time.Date(2026, 3, 15, 0, 0, 0, 0, wib))
	if got := start.Format(time.RFC3339); got != "2026-03-01T00:00:00+07:00" {
		t.Fatalf("start = %s", got)
	}
	if got := end.Format("2006-01-02"); got != "2026-04-01" {
		t.Fatalf("end = %s", got)
	}
	if got := prev.Format("2006-01-02"); got != "2026-02-01" {
		t.Fatalf("prev = %s", got)
	}
}
