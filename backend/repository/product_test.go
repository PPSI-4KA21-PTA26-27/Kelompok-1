package repository

import (
	"reflect"
	"testing"

	"hii-stranger1/backend/models"
)

func TestEscapeLike(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hoodie", "hoodie"},
		{"50%", `50\%`},
		{"a_b", `a\_b`},
		{`c:\folder`, `c:\\folder`},
		{`%_\`, `\%\_\\`},
	}
	for _, tt := range tests {
		if got := escapeLike(tt.in); got != tt.want {
			t.Errorf("escapeLike(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// Nilai filter tidak boleh masuk ke teks SQL; harus lewat argumen dengan nomor placeholder yang urut.
func TestBuildProductWhere(t *testing.T) {
	const base = " WHERE p.is_active AND c.is_active"

	tests := []struct {
		name      string
		filter    models.ProductFilter
		wantWhere string
		wantArgs  []any
	}{
		{
			"tanpa filter hanya menyaring produk dan kategori aktif",
			models.ProductFilter{Availability: models.AvailabilityAll},
			base,
			nil,
		},
		{
			"kategori dan kata kunci memakai placeholder berurutan",
			models.ProductFilter{CategorySlug: "t-shirt", Search: "50%", Availability: models.AvailabilityAll},
			base + " AND c.slug = $1 AND p.name ILIKE '%' || $2 || '%'",
			[]any{"t-shirt", `50\%`},
		},
		{
			"hanya kata kunci memakai placeholder pertama",
			models.ProductFilter{Search: "hoodie"},
			base + " AND p.name ILIKE '%' || $1 || '%'",
			[]any{"hoodie"},
		},
		{
			"stok tersedia tidak butuh argumen",
			models.ProductFilter{Availability: models.AvailabilityInStock},
			base + " AND COALESCE(s.total_stock, 0) > 0",
			nil,
		},
		{
			"pre-order",
			models.ProductFilter{CategorySlug: "outerwear", Availability: models.AvailabilityPreorder},
			base + " AND c.slug = $1 AND p.is_preorder",
			[]any{"outerwear"},
		},
		{
			"input berbahaya tetap menjadi argumen, bukan teks SQL",
			models.ProductFilter{CategorySlug: "x' OR '1'='1"},
			base + " AND c.slug = $1",
			[]any{"x' OR '1'='1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWhere, gotArgs := buildProductWhere(tt.filter)
			if gotWhere != tt.wantWhere {
				t.Errorf("where = %q, want %q", gotWhere, tt.wantWhere)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("args = %#v, want %#v", gotArgs, tt.wantArgs)
			}
		})
	}
}

// Setiap pilihan urutan yang diterima service harus punya ORDER BY di repository.
func TestProductOrderByCoversAllSorts(t *testing.T) {
	for _, s := range []models.ProductSort{models.SortNewest, models.SortPriceAsc, models.SortPriceDesc, models.SortNameAsc} {
		if productOrderBy[s] == "" {
			t.Errorf("urutan %q belum punya ORDER BY", s)
		}
	}
}
