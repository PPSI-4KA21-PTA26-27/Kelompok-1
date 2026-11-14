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
