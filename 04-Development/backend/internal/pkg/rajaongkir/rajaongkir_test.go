package rajaongkir

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient memastikan format request/response sesuai dokumentasi RajaOngkir
// memakai server palsu (tanpa memanggil API asli).
func TestClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("key") != "secret" {
			t.Errorf("header key = %q", r.Header.Get("key"))
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/destination/sub-district/5823":
			_, _ = w.Write([]byte(`{"meta":{"message":"ok","code":200,"status":"success"},
				"data":[{"id":68513,"name":"BALERAKSA","zip_code":"53355"},{"id":2,"name":"X","zip_code":53356}]}`))
		case "/destination/district/1":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"meta":{"message":"not found","code":404,"status":"error"},"data":null}`))
		case "/calculate/domestic-cost":
			_ = r.ParseForm()
			if r.Form.Get("origin") != "10" || r.Form.Get("weight") != "1500" {
				t.Errorf("form = %v", r.Form)
			}
			_, _ = w.Write([]byte(`{"meta":{"message":"ok","code":200,"status":"success"},
				"data":[{"name":"JNE","code":"jne","service":"REG","description":"Reguler","cost":18000,"etd":"2-3 day"}]}`))
		default:
			t.Errorf("path tidak terduga %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "secret")
	ctx := context.Background()

	subs, err := c.SubDistricts(ctx, 5823)
	if err != nil || len(subs) != 2 || subs[0].ZipCode != "53355" || subs[1].ZipCode != "53356" {
		t.Fatalf("SubDistricts = %+v, %v", subs, err)
	}

	empty, err := c.Districts(ctx, 1)
	if err != nil || len(empty) != 0 {
		t.Fatalf("Districts 404 = %+v, %v", empty, err)
	}

	costs, err := c.DomesticCost(ctx, 10, 20, 1500, "jne")
	if err != nil || len(costs) != 1 || costs[0].Cost != 18000 || costs[0].Service != "REG" {
		t.Fatalf("DomesticCost = %+v, %v", costs, err)
	}
}
