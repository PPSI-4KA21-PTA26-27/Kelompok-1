package handlers

import (
	"net/http"

	"hii-stranger1/backend/models"
	"hii-stranger1/backend/response"
	"hii-stranger1/backend/services"
)

// Alur: routes (/api/v1/auth/*) -> AuthHandler -> services.AuthService.
// Handler hanya mengurus HTTP: baca JSON, panggil service, kirim respons. Aturan bisnis ada di service.
// Endpoint ini publik (tanpa token), jadi di routes dibungkus middleware.RateLimit.

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register: POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.auth.Register(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, result)
}

// Login: POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.auth.Login(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

// ForgotPassword: POST /api/v1/auth/forgot-password
// Selalu 200 untuk email yang formatnya benar, terdaftar atau tidak (lihat AuthService.ForgotPassword).
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.auth.ForgotPassword(r.Context(), req); err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "jika email terdaftar, tautan reset kata sandi akan dikirim",
	})
}

// ResetPassword: POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.auth.ResetPassword(r.Context(), req); err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "kata sandi berhasil diperbarui"})
}
