package banner

import (
	"encoding/json"
	"testing"
	"time"
)

func TestComputeStatus(t *testing.T) {
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	past, future := now.Add(-time.Hour), now.Add(time.Hour)
	tests := []struct {
		name string
		b    Banner
		want string
	}{
		{"inactive wins", Banner{IsActive: false, StartsAt: &past}, StatusInactive},
		{"no schedule is live", Banner{IsActive: true}, StatusLive},
		{"in window", Banner{IsActive: true, StartsAt: &past, EndsAt: &future}, StatusLive},
		{"not started", Banner{IsActive: true, StartsAt: &future}, StatusScheduled},
		{"ended", Banner{IsActive: true, EndsAt: &past}, StatusEnded},
		{"ends exactly now", Banner{IsActive: true, EndsAt: &now}, StatusEnded},
	}
	for _, tc := range tests {
		tc.b.computeStatus(now)
		if tc.b.Status != tc.want {
			t.Errorf("%s: status = %q, want %q", tc.name, tc.b.Status, tc.want)
		}
	}
}

func TestRequestChanges(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		create bool
		fields []string
	}{
		{"create needs image", `{"placement":"hero"}`, true, []string{"image_url"}},
		{"bad placement", `{"placement":"sidebar","image_url":"https://x.test/a.jpg"}`, true, []string{"placement"}},
		{"internal link ok", `{"image_url":"https://x.test/a.jpg","link_url":"/shop?category=1"}`, true, nil},
		{"protocol-relative link rejected", `{"image_url":"https://x.test/a.jpg","link_url":"//evil.test"}`, true, []string{"link_url"}},
		{"empty update", `{}`, false, []string{"body"}},
		{"clear schedule is a change", `{"starts_at":null}`, false, nil},
	}
	for _, tc := range tests {
		var req request
		if err := json.Unmarshal([]byte(tc.body), &req); err != nil {
			t.Fatalf("%s: unmarshal: %v", tc.name, err)
		}
		_, errs := req.changes(tc.create)
		if len(errs) != len(tc.fields) {
			t.Fatalf("%s: errors = %v, want fields %v", tc.name, errs, tc.fields)
		}
		for i, f := range tc.fields {
			if errs[i].Field != f {
				t.Errorf("%s: field = %q, want %q", tc.name, errs[i].Field, f)
			}
		}
	}
}

func TestCheckSchedule(t *testing.T) {
	a := time.Now()
	b := a.Add(time.Hour)
	if checkSchedule(Banner{StartsAt: &a, EndsAt: &b}) != nil {
		t.Error("valid schedule rejected")
	}
	if checkSchedule(Banner{StartsAt: &b, EndsAt: &a}) == nil {
		t.Error("reversed schedule accepted")
	}
}
