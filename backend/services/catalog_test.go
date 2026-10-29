package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"hii-stranger1/backend/models"
)

func ptr[T any](v T) *T { return &v }

func TestParseProductList(t *testing.T) {
	defaults := models.ProductFilter{
		Availability: models.AvailabilityAll,
		Sort:         models.SortNewest,
		Page:         1,
		Limit:        12,
	}

	tests := []struct {
		name       string
		req        models.ProductListRequest
		want       models.ProductFilter
		wantFields []string // nama field yang harus salah; kosong berarti request valid
	}{
		{"semua kosong memakai nilai bawaan", models.ProductListRequest{}, defaults, nil},
		{
			"semua terisi dan dirapikan",
			models.ProductListRequest{Category: " T-Shirt ", Search: " hoodie ", Availability: "in_stock", Sort: "harga_terendah", Page: "2", Limit: "24"},
			models.ProductFilter{CategorySlug: "t-shirt", Search: "hoodie", Availability: models.AvailabilityInStock, Sort: models.SortPriceAsc, Page: 2, Limit: 24},
			nil,
		},
		{
			"batas atas masih valid",
			models.ProductListRequest{Page: "100000", Limit: "50"},
			models.ProductFilter{Availability: models.AvailabilityAll, Sort: models.SortNewest, Page: 100000, Limit: 50},
			nil,
		},
		{"availability tidak dikenal", models.ProductListRequest{Availability: "semua"}, models.ProductFilter{}, []string{"availability"}},
		{"sort tidak dikenal", models.ProductListRequest{Sort: "termurah"}, models.ProductFilter{}, []string{"sort"}},
		{"page nol", models.ProductListRequest{Page: "0"}, models.ProductFilter{}, []string{"page"}},
		{"page bukan angka", models.ProductListRequest{Page: "dua"}, models.ProductFilter{}, []string{"page"}},
		{"page melewati batas", models.ProductListRequest{Page: "100001"}, models.ProductFilter{}, []string{"page"}},
		{"limit nol", models.ProductListRequest{Limit: "0"}, models.ProductFilter{}, []string{"limit"}},
		{"limit negatif", models.ProductListRequest{Limit: "-1"}, models.ProductFilter{}, []string{"limit"}},
		{"limit melewati batas", models.ProductListRequest{Limit: "51"}, models.ProductFilter{}, []string{"limit"}},
		{"kata kunci terlalu panjang", models.ProductListRequest{Search: strings.Repeat("a", 101)}, models.ProductFilter{}, []string{"q"}},
		{
			"beberapa field salah dilaporkan sekaligus",
			models.ProductListRequest{Availability: "x", Sort: "y", Page: "z"},
			models.ProductFilter{},
			[]string{"availability", "sort", "page"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseProductList(tt.req)

			if len(tt.wantFields) == 0 {
				if err != nil {
					t.Fatalf("parseProductList() error = %v, want nil", err)
				}
				if got != tt.want {
					t.Errorf("parseProductList() = %+v, want %+v", got, tt.want)
				}
				return
			}

			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("parseProductList() error = %v, want *ValidationError", err)
			}
			if len(ve.Fields) != len(tt.wantFields) {
				t.Errorf("field salah = %v, want hanya %v", ve.Fields, tt.wantFields)
			}
			for _, name := range tt.wantFields {
				if ve.Fields[name] == "" {
					t.Errorf("field %q tidak dilaporkan salah: %v", name, ve.Fields)
				}
			}
		})
	}
}

func TestDiscountPercent(t *testing.T) {
	tests := []struct {
		name     string
		price    int64
		original *int64
		want     int
	}{
		{"tanpa harga coret", 489000, nil, 0},
		{"harga coret sama dengan harga jual", 489000, ptr(int64(489000)), 0},
		{"harga coret lebih rendah dari harga jual", 489000, ptr(int64(400000)), 0},
		{"dibulatkan ke atas (contoh desain 549rb ke 489rb)", 489000, ptr(int64(549000)), 11},
		{"pas 20 persen (contoh desain 325rb ke 260rb)", 260000, ptr(int64(325000)), 20},
		{"33,33 persen dibulatkan ke bawah", 2000, ptr(int64(3000)), 33},
		{"66,67 persen dibulatkan ke atas", 1000, ptr(int64(3000)), 67},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := discountPercent(tt.price, tt.original); got != tt.want {
				t.Errorf("discountPercent(%d, %v) = %d, want %d", tt.price, tt.original, got, tt.want)
			}
		})
	}
}

func TestPreorderLabel(t *testing.T) {
	tests := []struct {
		name       string
		isPreorder bool
		days       *int
		want       string
	}{
		{"ready stock tidak punya label", false, ptr(7), ""},
		{"pre-order dengan estimasi", true, ptr(7), "Pre-Order (7 Hari)"},
		{"pre-order tanpa estimasi", true, nil, "Pre-Order"},
		{"pre-order dengan estimasi nol", true, ptr(0), "Pre-Order"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := preorderLabel(tt.isPreorder, tt.days); got != tt.want {
				t.Errorf("preorderLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStockStatus(t *testing.T) {
	tests := []struct {
		total int
		want  models.StockStatus
	}{
		{-1, models.StockSoldOut},
		{0, models.StockSoldOut},
		{1, models.StockLow},
		{4, models.StockLow},
		{5, models.StockInStock},
		{147, models.StockInStock},
	}
	for _, tt := range tests {
		if got := stockStatus(tt.total); got != tt.want {
			t.Errorf("stockStatus(%d) = %q, want %q", tt.total, got, tt.want)
		}
	}
}

// Posisi banner yang salah harus ditolak sebelum menyentuh database,
// jadi service tanpa repository pun cukup untuk mengujinya.
func TestListBannersRejectsUnknownPosition(t *testing.T) {
	s := &CatalogService{}
	_, err := s.ListBanners(context.Background(), "sidebar")

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("ListBanners() error = %v, want *ValidationError", err)
	}
	if ve.Fields["position"] == "" {
		t.Errorf("field position tidak dilaporkan salah: %v", ve.Fields)
	}
}
