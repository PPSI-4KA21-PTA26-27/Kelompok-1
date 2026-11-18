package user

import (
	"errors"
	"log/slog"
	"net/http"
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
