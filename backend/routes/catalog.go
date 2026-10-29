package routes

import (
	"fmt"
	"net/http"

	"hii-stranger1/backend/db"
	"hii-stranger1/backend/handlers"
	"hii-stranger1/backend/repository"
	"hii-stranger1/backend/services"
)

// uploadDir adalah folder gambar unggahan, relatif terhadap folder tempat server dijalankan
// (folder backend, sama seperti .env dan migrations).
const uploadDir = "uploads"

// Alur penyambungan: db -> repository -> services -> handlers -> route.
// Cara membacanya: baca dari atas ke bawah, tiap langkah memakai hasil langkah sebelumnya.

// registerCatalogRoutes mendaftarkan route publik katalog: kategori, produk, banner,
// serta penyajian gambar unggahan.
func registerCatalogRoutes(mux *http.ServeMux) error {
	// 1. repository: akses database
	categories := repository.NewCategoryRepository(db.DB)
	products := repository.NewProductRepository(db.DB)
	banners := repository.NewBannerRepository(db.DB)

	// 2. services: aturan bisnis, plus penyimpan gambar (membuat folder uploads kalau belum ada)
	catalog := services.NewCatalogService(categories, products, banners)
	images, err := services.NewImageStore(uploadDir)
	if err != nil {
		return fmt.Errorf("siapkan penyimpan gambar: %w", err)
	}

	// 3. handlers: urusan HTTP
	catalogHandler := handlers.NewCatalogHandler(catalog)
	uploadHandler := handlers.NewUploadHandler(images)

	// Semua route di sini publik (tanpa token) dan hanya membaca.
	mux.HandleFunc("GET /api/v1/categories", catalogHandler.ListCategories)
	mux.HandleFunc("GET /api/v1/products", catalogHandler.ListProducts)
	mux.HandleFunc("GET /api/v1/products/{slug}", catalogHandler.GetProduct)
	mux.HandleFunc("GET /api/v1/banners", catalogHandler.ListBanners)
	mux.HandleFunc("GET /api/v1/uploads/{name}", uploadHandler.Serve)

	return nil
}
