package category

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

const nameMessage = "Nama wajib diisi, maksimal 100 karakter."

// Handler melayani endpoint /categories.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	Name     string `json:"name"`
	IsActive *bool  `json:"is_active"`
}

type updateRequest struct {
	Name     *string `json:"name"`
	IsActive *bool   `json:"is_active"`
}

// List melayani GET /categories. Publik; dipasang di belakang
// middleware.AuthenticateOptional supaya admin dikenali. include_inactive=true
// hanya boleh untuk admin.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, limit, ok := response.ParsePage(w, r)
	if !ok {
		return
	}

	includeInactive := false
	if v := r.URL.Query().Get("include_inactive"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "bad_request", "Parameter include_inactive tidak valid.")
			return
		}
		includeInactive = b
	}
	if includeInactive && !user.IsAdminRequest(r) {
		response.Forbidden(w)
		return
	}

	categories, total, err := h.svc.List(r.Context(), includeInactive, limit, (page-1)*limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "list categories failed", "error", err)
		response.Internal(w)
		return
	}

	items := make([]Response, 0, len(categories))
	for _, c := range categories {
		items = append(items, c.Response())
	}
	response.List(w, http.StatusOK, items, response.Meta{Page: page, Limit: limit, Total: total})
}

// Get melayani GET /categories/{categoryId}. Publik (dengan AuthenticateOptional);
// kategori nonaktif hanya terlihat oleh admin, selain itu 404.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("categoryId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	c, err := h.svc.Get(r.Context(), id, user.IsAdminRequest(r))
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case err != nil:
		slog.ErrorContext(r.Context(), "get category failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusOK, c.Response())
	}
}

// Create melayani POST /categories. Wajib di belakang Authenticate dan
// RequireRole admin.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)

	if errs := validateCreate(req); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	c, err := h.svc.Create(r.Context(), req.Name, isActive)
	switch {
	case errors.Is(err, ErrNameTaken):
		response.Conflict(w, "Nama kategori sudah ada.")
	case err != nil:
		slog.ErrorContext(r.Context(), "create category failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusCreated, c.Response())
	}
}

// Update melayani PATCH /categories/{categoryId}. Wajib di belakang
// Authenticate dan RequireRole admin.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("categoryId")
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

	c, err := h.svc.Update(r.Context(), id, changes)
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case errors.Is(err, ErrNameTaken):
		response.Conflict(w, "Nama kategori sudah ada.")
	case err != nil:
		slog.ErrorContext(r.Context(), "update category failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusOK, c.Response())
	}
}

// Delete melayani DELETE /categories/{categoryId}. Wajib di belakang
// Authenticate dan RequireRole admin.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("categoryId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	err := h.svc.Delete(r.Context(), id)
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case err != nil:
		slog.ErrorContext(r.Context(), "delete category failed", "error", err)
		response.Internal(w)
	default:
		response.NoContent(w)
	}
}

func validateCreate(req createRequest) []response.FieldError {
	var c validator.Collector
	c.Check(validator.RuneLenBetween(req.Name, 1, 100), "name", nameMessage)
	return c.Errors()
}

// changes memvalidasi dan menormalisasi body PATCH.
func (req updateRequest) changes() (Changes, []response.FieldError) {
	var c validator.Collector
	var out Changes

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		c.Check(validator.RuneLenBetween(name, 1, 100), "name", nameMessage)
		out.Name = &name
	}
	out.IsActive = req.IsActive

	c.Check(out.Name != nil || out.IsActive != nil, "body", "Isi minimal satu field: name atau is_active.")
	return out, c.Errors()
}
