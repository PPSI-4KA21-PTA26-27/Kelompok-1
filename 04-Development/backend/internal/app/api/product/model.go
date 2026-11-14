// Package product mengelola produk: model, akses data, dan endpoint.
package product

import "errors"

var (
	// ErrNotFound berarti produk tidak ada (atau tidak boleh dilihat pemanggil).
	ErrNotFound = errors.New("product: not found")
	// ErrCategoryNotFound berarti category_id tidak menunjuk kategori yang ada.
	ErrCategoryNotFound = errors.New("product: category not found")
	// ErrInUse berarti produk masih dipakai pesanan sehingga tidak bisa dihapus.
	ErrInUse = errors.New("product: still referenced by orders")
	// ErrPreorderEstimate berarti produk pre-order wajib punya po_estimate_days.
	ErrPreorderEstimate = errors.New("product: preorder requires po_estimate_days")
)

// Nilai yang dipakai aplikasi untuk kolom status dan availability_type.
const (
	StatusActive   = "active"
	StatusInactive = "inactive"

	AvailabilityReady    = "ready"
	AvailabilityPreorder = "preorder"
)

// Product mencerminkan satu baris tabel products. Price berupa string desimal
// agar tidak kehilangan ketelitian (jangan pakai float64 untuk uang).
type Product struct {
	ID               string
	CategoryID       *string
	Name             string
	Description      *string
	Price            string
	DiscountPercent  float64
	WeightGram       int
	AvailabilityType string
	PoEstimateDays   *int
	Status           string

	// PrimaryImageURL hanya diisi oleh List (diturunkan dari product_images).
	PrimaryImageURL *string
	// TotalStock & VariantCount hanya diisi di daftar produk.
	TotalStock   int
	VariantCount int
}

// Response adalah bentuk produk di API (schema Product pada openapi.yaml).
type Response struct {
	ID               string  `json:"id"`
	CategoryID       *string `json:"category_id"`
	Name             string  `json:"name"`
	Description      *string `json:"description"`
	Price            string  `json:"price"`
	DiscountPercent  float64 `json:"discount_percent"`
	WeightGram       int     `json:"weight_gram"`
	AvailabilityType string  `json:"availability_type"`
	PoEstimateDays   *int    `json:"po_estimate_days"`
	Status           string  `json:"status"`
}

// Response mengubah Product menjadi bentuk yang dikirim ke client.
func (p Product) Response() Response {
	return Response{
		ID:               p.ID,
		CategoryID:       p.CategoryID,
		Name:             p.Name,
		Description:      p.Description,
		Price:            p.Price,
		DiscountPercent:  p.DiscountPercent,
		WeightGram:       p.WeightGram,
		AvailabilityType: p.AvailabilityType,
		PoEstimateDays:   p.PoEstimateDays,
		Status:           p.Status,
	}
}

// ListItem adalah bentuk produk pada daftar (schema ProductListItem).
type ListItem struct {
	Response
	PrimaryImageURL *string `json:"primary_image_url"`
	// TotalStock adalah jumlah stok semua varian; 0 berarti habis.
	TotalStock   int `json:"total_stock"`
	VariantCount int `json:"variant_count"`
}

// Detail adalah produk beserta varian dan gambarnya (schema ProductDetail).
type Detail struct {
	Response
	Variants []Variant `json:"variants"`
	Images   []Image   `json:"images"`
}

// Changes adalah perubahan produk yang sudah divalidasi. Field nil berarti
// tidak diubah.
type Changes struct {
	CategoryID       *string
	Name             *string
	Description      *string
	Price            *string
	DiscountPercent  *float64
	WeightGram       *int
	AvailabilityType *string
	PoEstimateDays   *int
	Status           *string
}

// Filter adalah syarat pencarian untuk List. Field kosong berarti tidak dipakai.
type Filter struct {
	Query            string
	CategoryID       string
	AvailabilityType string
	Status           string
	MinPrice         string
	MaxPrice         string
	Sort             string
	// NewDays > 0 membatasi ke produk yang dibuat N hari terakhir ("Baru datang").
	NewDays int
}
