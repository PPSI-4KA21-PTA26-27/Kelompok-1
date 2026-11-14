package product

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

const (
	nameMessage     = "Nama wajib diisi, maksimal 200 karakter."
	priceMessage    = "Harga harus angka desimal tak negatif, contoh 150000."
	discountMessage = "Diskon harus antara 0 sampai 100."
	weightMessage   = "Berat harus bilangan bulat minimal 1 (gram)."
	categoryMessage = "category_id harus berupa UUID."
	availMessage    = "availability_type harus ready atau preorder."
	statusMessage   = "status harus active atau inactive."
	poMessage       = "po_estimate_days tidak boleh negatif."
	preorderMessage = "po_estimate_days wajib diisi untuk produk preorder."
)

// Handler melayani endpoint /products.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	CategoryID       *string  `json:"category_id"`
	Name             string   `json:"name"`
	Description      *string  `json:"description"`
	Price            string   `json:"price"`
	DiscountPercent  *float64 `json:"discount_percent"`
	WeightGram       int      `json:"weight_gram"`
	AvailabilityType string   `json:"availability_type"`
	PoEstimateDays   *int     `json:"po_estimate_days"`
	Status           string   `json:"status"`
}

type updateRequest struct {
	CategoryID       *string  `json:"category_id"`
	Name             *string  `json:"name"`
	Description      *string  `json:"description"`
	Price            *string  `json:"price"`
	DiscountPercent  *float64 `json:"discount_percent"`
	WeightGram       *int     `json:"weight_gram"`
	AvailabilityType *string  `json:"availability_type"`
	PoEstimateDays   *int     `json:"po_estimate_days"`
	Status           *string  `json:"status"`
}

// List melayani GET /products. Publik; dipasang di belakang
// middleware.AuthenticateOptional supaya admin dikenali. Parameter status
// hanya boleh untuk admin.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, limit, ok := response.ParsePage(w, r)
	if !ok {
		return
	}

	f, msg := parseFilter(r)
	if msg != "" {
		response.Error(w, http.StatusBadRequest, "bad_request", msg)
		return
	}
	admin := user.IsAdminRequest(r)
	if f.Status != "" && !admin {
		response.Forbidden(w)
		return
	}

	products, total, err := h.svc.List(r.Context(), f, admin, limit, (page-1)*limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "list products failed", "error", err)
		response.Internal(w)
		return
	}

	items := make([]ListItem, 0, len(products))
	for _, p := range products {
		items = append(items, ListItem{Response: p.Response(), PrimaryImageURL: p.PrimaryImageURL,
			TotalStock: p.TotalStock, VariantCount: p.VariantCount})
	}
	response.List(w, http.StatusOK, items, response.Meta{Page: page, Limit: limit, Total: total})
}

// Get melayani GET /products/{productId}: produk beserta varian dan gambar.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("productId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	d, err := h.svc.GetDetail(r.Context(), id, user.IsAdminRequest(r))
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case err != nil:
		slog.ErrorContext(r.Context(), "get product failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusOK, d)
	}
}

// Create melayani POST /products (admin).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Price = strings.TrimSpace(req.Price)

	if errs := validateCreate(req); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	p := Product{
		CategoryID:       req.CategoryID,
		Name:             req.Name,
		Description:      req.Description,
		Price:            req.Price,
		WeightGram:       req.WeightGram,
		AvailabilityType: req.AvailabilityType,
		PoEstimateDays:   req.PoEstimateDays,
		Status:           req.Status,
	}
	if req.DiscountPercent != nil {
		p.DiscountPercent = *req.DiscountPercent
	}

	created, err := h.svc.Create(r.Context(), p)
	if h.writeError(w, r, "create", err) {
		return
	}
	response.JSON(w, http.StatusCreated, created.Response())
}

// Update melayani PATCH /products/{productId} (admin).
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("productId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	var req updateRequest
	if !response.Decode(w, r, &req) {
		return
	}
	changes, errs := req.changes()
	if len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	p, err := h.svc.Update(r.Context(), id, changes)
	if h.writeError(w, r, "update", err) {
		return
	}
	response.JSON(w, http.StatusOK, p.Response())
}

// Delete melayani DELETE /products/{productId} (admin).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("productId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	err := h.svc.Delete(r.Context(), id)
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case errors.Is(err, ErrInUse):
		response.Error(w, http.StatusConflict, "conflict",
			"Produk sudah pernah dipesan. Ubah statusnya menjadi inactive.")
	case err != nil:
		slog.ErrorContext(r.Context(), "delete product failed", "error", err)
		response.Internal(w)
	default:
		response.NoContent(w)
	}
}

// writeError menjawab error dari Create/Update. Mengembalikan true bila err
// tidak nil (response sudah ditulis).
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case errors.Is(err, ErrCategoryNotFound):
		response.Error(w, http.StatusUnprocessableEntity, "validation_failed", "Data tidak valid.",
			response.FieldError{Field: "category_id", Message: "Kategori tidak ditemukan."})
	case errors.Is(err, ErrPreorderEstimate):
		response.Error(w, http.StatusUnprocessableEntity, "validation_failed", "Data tidak valid.",
			response.FieldError{Field: "po_estimate_days", Message: preorderMessage})
	default:
		slog.ErrorContext(r.Context(), op+" product failed", "error", err)
		response.Internal(w)
	}
	return true
}

// parseFilter membaca dan memvalidasi query list. Pesan tidak kosong berarti
// ada parameter yang salah.
func parseFilter(r *http.Request) (Filter, string) {
	q := r.URL.Query()
	f := Filter{
		Query:            strings.TrimSpace(q.Get("q")),
		CategoryID:       q.Get("category_id"),
		AvailabilityType: q.Get("availability_type"),
		Status:           q.Get("status"),
		MinPrice:         q.Get("min_price"),
		MaxPrice:         q.Get("max_price"),
		Sort:             q.Get("sort"),
	}
	if v := q.Get("new_days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 90 {
			return f, "Parameter new_days harus 1 sampai 90."
		}
		f.NewDays = n
	}
	switch {
	case utf8.RuneCountInString(f.Query) > 100:
		return f, "Parameter q maksimal 100 karakter."
	case f.CategoryID != "" && !validator.UUID(f.CategoryID):
		return f, "Parameter category_id tidak valid."
	case f.AvailabilityType != "" && !validAvailability(f.AvailabilityType):
		return f, "Parameter availability_type tidak valid."
	case f.Status != "" && !validStatus(f.Status):
		return f, "Parameter status tidak valid."
	case f.MinPrice != "" && !validator.Money(f.MinPrice):
		return f, "Parameter min_price tidak valid."
	case f.MaxPrice != "" && !validator.Money(f.MaxPrice):
		return f, "Parameter max_price tidak valid."
	}
	if _, ok := orderBy[f.Sort]; !ok {
		return f, "Parameter sort tidak valid."
	}
	return f, ""
}

func validateCreate(req createRequest) []response.FieldError {
	var c validator.Collector
	c.Check(validator.RuneLenBetween(req.Name, 1, 200), "name", nameMessage)
	c.Check(validator.Money(req.Price), "price", priceMessage)
	c.Check(req.DiscountPercent == nil || validDiscount(*req.DiscountPercent), "discount_percent", discountMessage)
	c.Check(req.WeightGram >= 1, "weight_gram", weightMessage)
	c.Check(validAvailability(req.AvailabilityType), "availability_type", availMessage)
	c.Check(validStatus(req.Status), "status", statusMessage)
	c.Check(req.CategoryID == nil || validator.UUID(*req.CategoryID), "category_id", categoryMessage)
	c.Check(req.PoEstimateDays == nil || *req.PoEstimateDays >= 0, "po_estimate_days", poMessage)
	c.Check(req.AvailabilityType != AvailabilityPreorder || req.PoEstimateDays != nil,
		"po_estimate_days", preorderMessage)
	return c.Errors()
}

// changes memvalidasi dan menormalisasi body PATCH. Aturan pre-order untuk
// hasil akhir diperiksa di service karena butuh data produk yang sekarang.
func (req updateRequest) changes() (Changes, []response.FieldError) {
	var c validator.Collector
	out := Changes{
		CategoryID:       req.CategoryID,
		Description:      req.Description,
		DiscountPercent:  req.DiscountPercent,
		WeightGram:       req.WeightGram,
		AvailabilityType: req.AvailabilityType,
		PoEstimateDays:   req.PoEstimateDays,
		Status:           req.Status,
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		c.Check(validator.RuneLenBetween(name, 1, 200), "name", nameMessage)
		out.Name = &name
	}
	if req.Price != nil {
		price := strings.TrimSpace(*req.Price)
		c.Check(validator.Money(price), "price", priceMessage)
		out.Price = &price
	}
	c.Check(req.CategoryID == nil || validator.UUID(*req.CategoryID), "category_id", categoryMessage)
	c.Check(req.DiscountPercent == nil || validDiscount(*req.DiscountPercent), "discount_percent", discountMessage)
	c.Check(req.WeightGram == nil || *req.WeightGram >= 1, "weight_gram", weightMessage)
	c.Check(req.AvailabilityType == nil || validAvailability(*req.AvailabilityType), "availability_type", availMessage)
	c.Check(req.Status == nil || validStatus(*req.Status), "status", statusMessage)
	c.Check(req.PoEstimateDays == nil || *req.PoEstimateDays >= 0, "po_estimate_days", poMessage)

	c.Check(out != (Changes{}), "body", "Isi minimal satu field yang ingin diubah.")
	return out, c.Errors()
}

func validDiscount(d float64) bool { return d >= 0 && d <= 100 }

func validAvailability(s string) bool {
	return s == AvailabilityReady || s == AvailabilityPreorder
}

func validStatus(s string) bool {
	return s == StatusActive || s == StatusInactive
}
