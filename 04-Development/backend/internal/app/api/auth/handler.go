package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"hii-stranger1/backend/internal/app/api/middleware"
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

type verifyRequest struct {
	Token string `json:"token"`
}

// VerifyEmail melayani POST /auth/verify-email (publik). Body {"token": "..."}
// dari tautan email; response berisi profil user yang sudah terverifikasi.
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	if !validator.ByteLenBetween(req.Token, 20, 200) {
		response.Error(w, http.StatusUnprocessableEntity, "validation_failed", "Data tidak valid.",
			response.FieldError{Field: "token", Message: "Token verifikasi tidak valid."})
		return
	}
	u, err := h.svc.VerifyEmail(r.Context(), req.Token)
	switch {
	case errors.Is(err, user.ErrInvalidVerification):
		response.Error(w, http.StatusGone, "verification_invalid",
			"Tautan verifikasi tidak valid atau sudah kedaluwarsa. Minta kirim ulang dari halaman akun.")
	case err != nil:
		slog.ErrorContext(r.Context(), "verify email failed", "error", err)
		response.Internal(w)
	default:
		response.JSON(w, http.StatusOK, u.Response())
	}
}

// ResendVerification melayani POST /auth/verification/resend (login).
func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	p, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "Autentikasi diperlukan.")
		return
	}
	err := h.svc.ResendVerification(r.Context(), p.UserID)
	switch {
	case errors.Is(err, ErrAlreadyVerified):
		response.Conflict(w, "Email sudah terverifikasi.")
	case errors.Is(err, user.ErrNotFound):
		response.NotFound(w)
	case err != nil:
		slog.ErrorContext(r.Context(), "resend verification failed", "error", err)
		response.Internal(w)
	default:
		response.NoContent(w)
	}
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPassword melayani POST /auth/password/forgot (publik). Selalu 204
// untuk email berformat valid, terdaftar atau tidak.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.Email = normalizeEmail(req.Email)
	if !validator.Email(req.Email) {
		response.Error(w, http.StatusUnprocessableEntity, "validation_failed", "Data tidak valid.",
			response.FieldError{Field: "email", Message: "Format email tidak valid."})
		return
	}
	if err := h.svc.RequestPasswordReset(r.Context(), req.Email); err != nil {
		slog.ErrorContext(r.Context(), "request password reset failed", "error", err)
		response.Internal(w)
		return
	}
	response.NoContent(w)
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func validateResetPassword(req resetPasswordRequest) []response.FieldError {
	var c validator.Collector
	c.Check(validator.ByteLenBetween(req.Token, 20, 200), "token", "Tautan reset tidak valid.")
	// Batas 72 byte berasal dari algoritma bcrypt.
	c.Check(validator.ByteLenBetween(req.Password, 8, 72), "password", "Password harus 8 sampai 72 karakter.")
	return c.Errors()
}

// ResetPassword melayani POST /auth/password/reset (publik).
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if !response.Decode(w, r, &req) {
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	if errs := validateResetPassword(req); len(errs) > 0 {
		response.ValidationFailed(w, errs...)
		return
	}
	err := h.svc.ResetPassword(r.Context(), req.Token, req.Password)
	switch {
	case errors.Is(err, user.ErrInvalidPasswordReset):
		response.Error(w, http.StatusGone, "reset_invalid",
			"Tautan reset tidak valid atau sudah kedaluwarsa. Minta tautan baru dari halaman lupa password.")
	case err != nil:
		slog.ErrorContext(r.Context(), "reset password failed", "error", err)
		response.Internal(w)
	default:
		response.NoContent(w)
	}
}
