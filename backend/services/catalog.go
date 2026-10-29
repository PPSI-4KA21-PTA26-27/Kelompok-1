package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"hii-stranger1/backend/models"
	"hii-stranger1/backend/repository"
)

// Alur: handlers (catalog) -> CatalogService -> repository (category, product, banner) -> database.
// Service ini melayani data publik storefront: kategori aktif, produk aktif, dan banner yang
// sedang tayang. Aturan bisnis di sini: validasi filter katalog, hitung diskon, label pre-order,
// dan status stok. Query SQL tidak ada di sini, semuanya di repository.

const (
	defaultPageSize = 12     // jumlah produk per halaman kalau klien tidak menyebut limit
	maxPageSize     = 50     // batas atas supaya satu request tidak menarik terlalu banyak data
	maxPage         = 100000 // mencegah offset yang absurd (dan overflow saat dihitung)
	maxSearchLen    = 100

	// lowStockThreshold: stok total di bawah angka ini dianggap "tinggal sedikit".
	// Sama dengan "Restock Alert (<5 pcs)" di halaman admin produk.
	lowStockThreshold = 5
)

type CatalogService struct {
	categories *repository.CategoryRepository
	products   *repository.ProductRepository
	banners    *repository.BannerRepository
}

func NewCatalogService(
	categories *repository.CategoryRepository,
	products *repository.ProductRepository,
	banners *repository.BannerRepository,
) *CatalogService {
	return &CatalogService{categories: categories, products: products, banners: banners}
}

// ListCategories mengembalikan kategori aktif untuk navigasi dan filter.
func (s *CatalogService) ListCategories(ctx context.Context) ([]models.Category, error) {
	return s.categories.ListActive(ctx)
}

// ListProducts memvalidasi filter dari klien, mengambil satu halaman produk aktif,
// lalu melengkapi tiap produk dengan diskon, label pre-order, dan status stok.
func (s *CatalogService) ListProducts(ctx context.Context, req models.ProductListRequest) (*models.ProductPage, error) {
	filter, err := parseProductList(req)
	if err != nil {
		return nil, err
	}

	items, total, err := s.products.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	for i := range items {
		decorateProduct(&items[i])
	}
	return &models.ProductPage{Items: items, Page: filter.Page, Limit: filter.Limit, Total: total}, nil
}

// GetProduct mengambil detail satu produk aktif lewat slug.
// Mengembalikan repository.ErrNotFound kalau produk tidak ada atau tidak ditampilkan ke publik.
func (s *CatalogService) GetProduct(ctx context.Context, slug string) (*models.ProductDetail, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return nil, repository.ErrNotFound
	}

	product, err := s.products.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	decorateProduct(&product.ProductSummary)
	return product, nil
}

// ListBanners mengembalikan banner yang sedang tayang. position boleh kosong (semua slot),
// kalau diisi harus salah satu slot yang dikenal.
func (s *CatalogService) ListBanners(ctx context.Context, position string) ([]models.Banner, error) {
	pos := models.BannerPosition(strings.TrimSpace(position))
	if pos != "" && !pos.Valid() {
		return nil, &ValidationError{Fields: map[string]string{"position": "Posisi banner tidak valid"}}
	}
	return s.banners.ListActive(ctx, pos)
}

// parseProductList memeriksa query string katalog dan mengisi nilai bawaan untuk yang kosong.
// Kalau ada yang tidak valid, semua pesan salahnya dikembalikan sekaligus per field
// (nama field = nama parameter di URL).
func parseProductList(req models.ProductListRequest) (models.ProductFilter, error) {
	f := models.ProductFilter{
		CategorySlug: strings.ToLower(strings.TrimSpace(req.Category)),
		Search:       strings.TrimSpace(req.Search),
		Availability: models.AvailabilityAll,
		Sort:         models.SortNewest,
		Page:         1,
		Limit:        defaultPageSize,
	}
	fields := map[string]string{}

	if utf8.RuneCountInString(f.Search) > maxSearchLen {
		fields["q"] = fmt.Sprintf("Kata kunci maksimal %d karakter", maxSearchLen)
	}

	if raw := strings.TrimSpace(req.Availability); raw != "" {
		switch a := models.Availability(raw); a {
		case models.AvailabilityAll, models.AvailabilityInStock, models.AvailabilityPreorder:
			f.Availability = a
		default:
			fields["availability"] = "Pilihan ketersediaan tidak valid"
		}
	}

	if raw := strings.TrimSpace(req.Sort); raw != "" {
		switch o := models.ProductSort(raw); o {
		case models.SortNewest, models.SortPriceAsc, models.SortPriceDesc, models.SortNameAsc:
			f.Sort = o
		default:
			fields["sort"] = "Pilihan urutan tidak valid"
		}
	}

	if raw := strings.TrimSpace(req.Page); raw != "" {
		if n, ok := parseBoundedInt(raw, maxPage); ok {
			f.Page = n
		} else {
			fields["page"] = fmt.Sprintf("Halaman harus berupa angka 1 sampai %d", maxPage)
		}
	}

	if raw := strings.TrimSpace(req.Limit); raw != "" {
		if n, ok := parseBoundedInt(raw, maxPageSize); ok {
			f.Limit = n
		} else {
			fields["limit"] = fmt.Sprintf("Jumlah per halaman harus berupa angka 1 sampai %d", maxPageSize)
		}
	}

	if err := collect(fields); err != nil {
		return models.ProductFilter{}, err
	}
	return f, nil
}

// parseBoundedInt membaca angka bulat dari teks. ok bernilai false kalau teks bukan angka
// atau angkanya di luar rentang 1 sampai upper.
func parseBoundedInt(raw string, upper int) (int, bool) {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > upper {
		return 0, false
	}
	return n, true
}

// decorateProduct mengisi field turunan yang tidak disimpan di database.
func decorateProduct(p *models.ProductSummary) {
	p.DiscountPercent = discountPercent(p.Price, p.OriginalPrice)
	p.PreorderLabel = preorderLabel(p.IsPreorder, p.PreorderDays)
	p.StockStatus = stockStatus(p.TotalStock)
}

// discountPercent menghitung persen diskon dari harga coret, dibulatkan ke angka terdekat
// (Rp549.000 menjadi Rp489.000 = 11%). Hasilnya 0 kalau tidak ada harga coret atau harga coret
// tidak lebih tinggi dari harga jual.
func discountPercent(price int64, original *int64) int {
	if original == nil || *original <= price {
		return 0
	}
	diff := *original - price
	return int((diff*100 + *original/2) / *original)
}

// preorderLabel membuat teks label pre-order, contoh "Pre-Order (7 Hari)".
// Hasilnya kosong untuk produk ready stock.
func preorderLabel(isPreorder bool, days *int) string {
	if !isPreorder {
		return ""
	}
	if days == nil || *days <= 0 {
		return "Pre-Order"
	}
	return fmt.Sprintf("Pre-Order (%d Hari)", *days)
}

// stockStatus meringkas total stok menjadi status untuk label di kartu produk.
func stockStatus(total int) models.StockStatus {
	switch {
	case total <= 0:
		return models.StockSoldOut
	case total < lowStockThreshold:
		return models.StockLow
	default:
		return models.StockInStock
	}
}
