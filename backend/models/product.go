package models

// Alur: repository membaca tabel products + product_variants, services melengkapi hasilnya
// (diskon, label pre-order, status stok), lalu handlers mengirimnya sebagai JSON.
// Harga memakai rupiah bulat tanpa desimal, contoh 489000 untuk Rp489.000.

// Availability: filter ketersediaan di halaman katalog.
type Availability string

const (
	AvailabilityAll      Availability = "all"      // semua produk
	AvailabilityInStock  Availability = "in_stock" // hanya yang stoknya masih ada
	AvailabilityPreorder Availability = "preorder" // hanya produk pre-order
)

// ProductSort: pilihan urutan di halaman katalog. Nilainya sama dengan value dropdown di desain.
type ProductSort string

const (
	SortNewest    ProductSort = "terbaru"
	SortPriceAsc  ProductSort = "harga_terendah"
	SortPriceDesc ProductSort = "harga_tertinggi"
	SortNameAsc   ProductSort = "nama_az"
)

// StockStatus: ringkasan stok untuk label di kartu produk ("Stok Ada", "Sisa 2", "Stok Habis").
type StockStatus string

const (
	StockInStock StockStatus = "in_stock"
	StockLow     StockStatus = "low_stock" // tinggal sedikit
	StockSoldOut StockStatus = "sold_out"
)

// CategoryRef adalah ringkasan kategori di dalam data produk.
type CategoryRef struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ProductSummary adalah produk untuk kartu di beranda dan halaman katalog.
// Field Discount*, PreorderLabel, dan StockStatus diisi oleh services, bukan oleh database.
type ProductSummary struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	Slug     string      `json:"slug"`
	Category CategoryRef `json:"category"`

	Price           int64  `json:"price"`
	OriginalPrice   *int64 `json:"original_price,omitempty"` // harga coret; kosong kalau tidak diskon
	DiscountPercent int    `json:"discount_percent"`         // 0 kalau tidak diskon

	IsPreorder    bool   `json:"is_preorder"`
	PreorderDays  *int   `json:"preorder_days,omitempty"`
	PreorderLabel string `json:"preorder_label,omitempty"` // contoh: "Pre-Order (7 Hari)"

	ImageURL    string      `json:"image_url"`
	TotalStock  int         `json:"total_stock"` // jumlah stok semua varian
	StockStatus StockStatus `json:"stock_status"`
}

// ProductVariant adalah satu kombinasi warna + ukuran dari sebuah produk.
type ProductVariant struct {
	ID    int64  `json:"id"`
	Color string `json:"color"`
	Size  string `json:"size"`
	SKU   string `json:"sku"`
	Stock int    `json:"stock"`
}

// ProductDetail adalah isi halaman detail produk: data kartu + deskripsi + daftar varian.
type ProductDetail struct {
	ProductSummary
	Description string           `json:"description"`
	Variants    []ProductVariant `json:"variants"`
}

// ProductListRequest adalah query string halaman katalog apa adanya (belum divalidasi).
// Semua field berupa teks; services yang memeriksa dan mengubahnya.
type ProductListRequest struct {
	Category     string // slug kategori
	Search       string // kata kunci nama produk
	Availability string
	Sort         string
	Page         string
	Limit        string
}

// ProductFilter adalah hasil validasi ProductListRequest, siap dipakai repository.
type ProductFilter struct {
	CategorySlug string
	Search       string
	Availability Availability
	Sort         ProductSort
	Page         int // mulai dari 1
	Limit        int
}

// ProductPage adalah satu halaman hasil katalog beserta info paginasi.
type ProductPage struct {
	Items []ProductSummary `json:"items"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
	Total int              `json:"total"` // jumlah semua produk yang cocok, bukan hanya halaman ini
}
