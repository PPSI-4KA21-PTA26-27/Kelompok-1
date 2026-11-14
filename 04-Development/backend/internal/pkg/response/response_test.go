package response

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePage(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantOK     bool
		wantPage   int
		wantLimit  int
		wantStatus int
	}{
		{name: "defaults", query: "", wantOK: true, wantPage: 1, wantLimit: 20, wantStatus: http.StatusOK},
		{name: "explicit", query: "page=3&limit=50", wantOK: true, wantPage: 3, wantLimit: 50, wantStatus: http.StatusOK},
		{name: "max limit", query: "limit=100", wantOK: true, wantPage: 1, wantLimit: 100, wantStatus: http.StatusOK},
		{name: "limit over max", query: "limit=101", wantStatus: http.StatusBadRequest},
		{name: "limit zero", query: "limit=0", wantStatus: http.StatusBadRequest},
		{name: "page zero", query: "page=0", wantStatus: http.StatusBadRequest},
		{name: "page negative", query: "page=-1", wantStatus: http.StatusBadRequest},
		{name: "page not a number", query: "page=abc", wantStatus: http.StatusBadRequest},
		{name: "page over max", query: "page=100001", wantStatus: http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tc.query, nil)
			rec := httptest.NewRecorder()

			page, limit, ok := ParsePage(rec, req)

			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if ok && (page != tc.wantPage || limit != tc.wantLimit) {
				t.Errorf("page, limit = %d, %d, want %d, %d", page, limit, tc.wantPage, tc.wantLimit)
			}
		})
	}
}
