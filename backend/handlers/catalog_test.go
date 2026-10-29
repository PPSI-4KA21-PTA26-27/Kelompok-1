package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hii-stranger1/backend/services"
)

// Test ini memakai CatalogService tanpa repository: yang diuji hanya jalur yang ditolak
// sebelum menyentuh database (validasi query -> 422).
func newCatalogMux() *http.ServeMux {
	h := NewCatalogHandler(&services.CatalogService{})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/products", h.ListProducts)
	mux.HandleFunc("GET /api/v1/banners", h.ListBanners)
	return mux
}

func TestCatalogHandlerRejectsInvalidQuery(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantField string
	}{
		{"urutan tidak dikenal", "/api/v1/products?sort=termurah", "sort"},
		{"ketersediaan tidak dikenal", "/api/v1/products?availability=semua", "availability"},
		{"halaman nol", "/api/v1/products?page=0", "page"},
		{"limit terlalu besar", "/api/v1/products?limit=1000", "limit"},
		{"posisi banner tidak dikenal", "/api/v1/banners?position=sidebar", "position"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newCatalogMux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.url, nil))

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
			}
			var body struct {
				Error  string            `json:"error"`
				Fields map[string]string `json:"fields"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("respons bukan JSON: %v", err)
			}
			if body.Fields[tt.wantField] == "" {
				t.Errorf("fields = %v, want ada pesan untuk %q", body.Fields, tt.wantField)
			}
		})
	}
}

func newUploadMux(t *testing.T) (*http.ServeMux, *services.ImageStore) {
	t.Helper()
	store, err := services.NewImageStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/uploads/{name}", NewUploadHandler(store).Serve)
	return mux, store
}

func TestUploadHandlerServesSavedImage(t *testing.T) {
	mux, store := newUploadMux(t)
	data := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, make([]byte, 64)...)
	name, err := store.Save(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/uploads/"+name, nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if !bytes.Equal(rec.Body.Bytes(), data) {
		t.Error("isi respons berbeda dari file yang disimpan")
	}
}

func TestUploadHandlerNotFound(t *testing.T) {
	mux, _ := newUploadMux(t)

	tests := []struct {
		name string
		path string
	}{
		{"nama valid tapi file tidak ada", "/api/v1/uploads/0123456789abcdef0123456789abcdef.jpg"},
		{"nama bukan buatan Save", "/api/v1/uploads/rahasia.txt"},
		{"path traversal ter-encode", "/api/v1/uploads/..%2F..%2F.env"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}
