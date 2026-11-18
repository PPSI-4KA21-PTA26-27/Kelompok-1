package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"hii-stranger1/backend/internal/app/api/user"
	"hii-stranger1/backend/internal/pkg/response"
	"hii-stranger1/backend/internal/pkg/validator"
)

// Handler melayani endpoint /auth.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type registerRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Phone    *string `json:"phone"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register melayani POST /auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !response.Decode(w, r, &req) {
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = normalizeEmail(req.Email)
	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)
		if phone == "" {
			req.Phone = nil
		} else {
			req.Phone = &phone
		}
	}

	if errs := validateRegister(req); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	res, err := h.svc.Register(r.Context(), RegisterInput(req))
	switch {
	case errors.Is(err, user.ErrEmailTaken):
		response.Conflict(w, "Email sudah terdaftar.")
	case err != nil:
		slog.ErrorContext(r.Context(), "register failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusCreated, res)
	}
}

// Login melayani POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !response.Decode(w, r, &req) {
		return
	}

	req.Email = normalizeEmail(req.Email)
	if errs := validateLogin(req); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}

	res, err := h.svc.Login(r.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "unauthorized", "Email atau password salah.")
	case err != nil:
		slog.ErrorContext(r.Context(), "login failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusOK, res)
	}
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validateRegister(req registerRequest) []response.FieldError {
	var c validator.Collector
	c.Check(validator.RuneLenBetween(req.Name, 1, 100), "name", "Nama wajib diisi, maksimal 100 karakter.")
	c.Check(validator.Email(req.Email), "email", "Format email tidak valid.")
	// Batas 72 byte berasal dari algoritma bcrypt.
	c.Check(validator.ByteLenBetween(req.Password, 8, 72), "password", "Password harus 8 sampai 72 karakter.")
	if req.Phone != nil {
		c.Check(validator.Phone(*req.Phone), "phone", "Nomor telepon tidak valid.")
	}
	return c.Errors()
}

func validateLogin(req loginRequest) []response.FieldError {
	var c validator.Collector
	c.Check(validator.Email(req.Email), "email", "Format email tidak valid.")
	c.Check(req.Password != "", "password", "Password wajib diisi.")
	return c.Errors()
}
