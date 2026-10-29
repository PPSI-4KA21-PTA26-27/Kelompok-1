package handlers

import (
	"net/http"

	"hii-stranger1/backend/models"
	"hii-stranger1/backend/response"
	"hii-stranger1/backend/services"
)

// Alur: routes (/api/v1/categories, /products, /banners) -> CatalogHandler -> services.CatalogService.
// Semua endpoint di sini publik (tanpa login) dan hanya membaca data.
// Handler hanya membaca query/path dari request; validasi dan aturan bisnis ada di service.

type CatalogHandler struct {
	catalog *services.CatalogService
}

func NewCatalogHandler(catalog *services.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalog: catalog}
}

// ListCategories: GET /api/v1/categories
func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.catalog.ListCategories(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, categories)
}

// ListProducts: GET /api/v1/products
// Query (semua opsional): category (slug), q, availability (all|in_stock|preorder),
// sort (terbaru|harga_terendah|harga_tertinggi|nama_az), page, limit.
func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := h.catalog.ListProducts(r.Context(), models.ProductListRequest{
		Category:     q.Get("category"),
		Search:       q.Get("q"),
		Availability: q.Get("availability"),
		Sort:         q.Get("sort"),
		Page:         q.Get("page"),
		Limit:        q.Get("limit"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, page)
}

// GetProduct: GET /api/v1/products/{slug}
func (h *CatalogHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.catalog.GetProduct(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, product)
}

// ListBanners: GET /api/v1/banners
// Query opsional: position (hero|announcement|mid_feed|catalog).
func (h *CatalogHandler) ListBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := h.catalog.ListBanners(r.Context(), r.URL.Query().Get("position"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, banners)
}
