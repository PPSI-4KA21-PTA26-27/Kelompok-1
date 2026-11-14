package user

import (
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"hii-stranger1/backend/internal/app/api/middleware"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

// Handler melayani endpoint /users.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Me melayani GET /users/me. Wajib dipasang di belakang middleware.Authenticate.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		unauthorized(w)
		return
	}

	u, err := h.svc.Get(r.Context(), principal.UserID)
	switch {
	case errors.Is(err, ErrNotFound):
		// Token valid tetapi akunnya sudah tidak ada.
		unauthorized(w)
		return
	case err != nil:
		slog.ErrorContext(r.Context(), "get profile failed", "error", err)
		response.Internal(w)
		return
	}

	if u.Status != StatusActive {
		unauthorized(w)
		return
	}
	response.JSON(w, http.StatusOK, u.Response())
}

type updateProfileRequest struct {
	Name  *string        `json:"name"`
	Phone OptionalString `json:"phone"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// UpdateMe melayani PATCH /users/me. Hanya name dan phone yang bisa diubah;
// field lain di body (email, role, status) diabaikan. Wajib dipasang di
// belakang middleware.Authenticate.
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		unauthorized(w)
		return
	}

	var req updateProfileRequest
	if !response.Decode(w, r, &req) {
		return
	}

	changes, errs := req.changes()
	if len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	u, err := h.svc.UpdateProfile(r.Context(), principal.UserID, changes)
	switch {
	case errors.Is(err, ErrNotFound):
		unauthorized(w)
	case err != nil:
		slog.ErrorContext(r.Context(), "update profile failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusOK, u.Response())
	}
}

// ChangePassword melayani PUT /users/me/password. Wajib dipasang di belakang
// middleware.Authenticate.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		unauthorized(w)
		return
	}

	var req changePasswordRequest
	if !response.Decode(w, r, &req) {
		return
	}

	if errs := validateChangePassword(req); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	err := h.svc.ChangePassword(r.Context(), principal.UserID, req.CurrentPassword, req.NewPassword)
	switch {
	case errors.Is(err, ErrNotFound):
		unauthorized(w)
	case errors.Is(err, ErrWrongPassword):
		// Bukan 401: token masih valid, jadi client tidak boleh menganggap sesi berakhir.
		response.Error(w, http.StatusUnprocessableEntity, "validation_failed", "Data tidak valid.",
			response.FieldError{Field: "current_password", Message: "Password saat ini salah."})
	case err != nil:
		slog.ErrorContext(r.Context(), "change password failed", "error", err)
		response.Internal(w)
	default:
		response.NoContent(w)
	}
}

// changes memvalidasi dan menormalisasi body PATCH. Nomor telepon kosong
// diperlakukan sama dengan null (menghapus nomor), seperti pada register.
func (req updateProfileRequest) changes() (ProfileChanges, []response.FieldError) {
	var c validator.Collector
	var out ProfileChanges

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		c.Check(validator.RuneLenBetween(name, 1, 100), "name", "Nama wajib diisi, maksimal 100 karakter.")
		out.Name = &name
	}

	if req.Phone.Set {
		out.Phone.Set = true
		if req.Phone.Value != nil {
			if phone := strings.TrimSpace(*req.Phone.Value); phone != "" {
				c.Check(validator.Phone(phone), "phone", "Nomor telepon tidak valid.")
				out.Phone.Value = &phone
			}
		}
	}

	c.Check(out.Name != nil || out.Phone.Set, "body", "Isi minimal satu field: name atau phone.")
	return out, c.Errors()
}

func validateChangePassword(req changePasswordRequest) []response.FieldError {
	var c validator.Collector
	c.Check(req.CurrentPassword != "", "current_password", "Password saat ini wajib diisi.")
	// Batas 72 byte berasal dari algoritma bcrypt.
	c.Check(validator.ByteLenBetween(req.NewPassword, 8, 72), "new_password", "Password baru harus 8 sampai 72 karakter.")
	return c.Errors()
}

type adminUpdateRequest struct {
	Role   *string `json:"role"`
	Status *string `json:"status"`
}

// AdminList melayani GET /admin/users. Wajib di belakang Authenticate dan
// RequireRole admin.
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	page, limit, ok := response.ParsePage(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	f := ListFilter{
		Q:      strings.TrimSpace(q.Get("q")),
		Role:   strings.TrimSpace(q.Get("role")),
		Status: strings.TrimSpace(q.Get("status")),
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
	if !validator.RuneLenBetween(f.Q, 0, 100) {
		response.Error(w, http.StatusBadRequest, "bad_request", "Parameter q maksimal 100 karakter.")
		return
	}

	users, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		slog.ErrorContext(r.Context(), "list users failed", "error", err)
		response.Internal(w)
		return
	}

	items := make([]Response, 0, len(users))
	for _, u := range users {
		items = append(items, u.Response())
	}
	response.List(w, http.StatusOK, items, response.Meta{Page: page, Limit: limit, Total: total})
}

// AdminGet melayani GET /admin/users/{userId}. ID yang bukan UUID dijawab 404
// karena user seperti itu pasti tidak ada.
func (h *Handler) AdminGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("userId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	u, err := h.svc.Get(r.Context(), id)
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case err != nil:
		slog.ErrorContext(r.Context(), "admin get user failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusOK, u.Response())
	}
}

// AdminUpdate melayani PATCH /admin/users/{userId}. Hanya role dan status yang
// bisa diubah.
func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		unauthorized(w)
		return
	}

	id := r.PathValue("userId")
	if !validator.UUID(id) {
		response.NotFound(w)
		return
	}

	var req adminUpdateRequest
	if !response.Decode(w, r, &req) {
		return
	}

	changes, errs := req.changes()
	if len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	u, err := h.svc.AdminUpdate(r.Context(), principal.UserID, id, changes)
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(w)
	case errors.Is(err, ErrSelfLockout):
		response.Conflict(w, "Anda tidak dapat menurunkan role atau menonaktifkan akun sendiri.")
	case err != nil:
		slog.ErrorContext(r.Context(), "admin update user failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusOK, u.Response())
	}
}

// changes memvalidasi body PATCH admin. Role dan status hanya menerima nilai
// yang dikenal aplikasi karena kolomnya di database tidak punya constraint.
func (req adminUpdateRequest) changes() (AdminChanges, []response.FieldError) {
	var c validator.Collector
	var out AdminChanges

	if req.Role != nil {
		role := strings.TrimSpace(*req.Role)
		c.Check(slices.Contains([]string{RoleCustomer, RoleOwner, RoleStaff}, role), "role", "Role harus customer, owner, atau staff.")
		out.Role = &role
	}
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		c.Check(slices.Contains([]string{StatusActive, StatusInactive}, status), "status", "Status harus active atau inactive.")
		out.Status = &status
	}

	c.Check(out.Role != nil || out.Status != nil, "body", "Isi minimal satu field: role atau status.")
	return out, c.Errors()
}

func unauthorized(w http.ResponseWriter) {
	response.Error(w, http.StatusUnauthorized, "unauthorized", "Autentikasi diperlukan.")
}

// Principal mengambil ID user dan status backoffice dari middleware auth.
func Principal(r *http.Request) (userID string, admin bool) {
	p, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		return "", false
	}
	return p.UserID, IsBackoffice(p.Role)
}

// IsAdminRequest: request datang dari staf backoffice yang sudah login.
func IsAdminRequest(r *http.Request) bool {
	_, admin := Principal(r)
	return admin
}
