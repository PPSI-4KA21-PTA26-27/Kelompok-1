package handlers

import (
	"net/http"

	"hii-stranger1/backend/middleware"
	"hii-stranger1/backend/models"
	"hii-stranger1/backend/response"
	"hii-stranger1/backend/services"
)

// Alur: routes (/api/v1/me*) -> middleware.Auth (isi context) -> AccountHandler -> services.AuthService.
// Id user SELALU diambil dari token lewat middleware.UserID, tidak pernah dari isi request.
// Dengan begitu user hanya bisa membaca dan mengubah akunnya sendiri.

type AccountHandler struct {
	auth *services.AuthService
}

func NewAccountHandler(auth *services.AuthService) *AccountHandler {
	return &AccountHandler{auth: auth}
}

// Me: GET /api/v1/me
func (h *AccountHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, user)
}

// UpdateMe: PUT /api/v1/me (ubah nama dan nomor HP)
func (h *AccountHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var req models.UpdateProfileRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	user, err := h.auth.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, user)
}

// ChangePassword: PUT /api/v1/me/password
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var req models.ChangePasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.auth.ChangePassword(r.Context(), userID, req); err != nil {
		writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "kata sandi berhasil diganti"})
}

// currentUserID membaca id user dari context. Kalau tidak ada, berarti route lupa dipasangi
// middleware.Auth; ditolak 401 supaya tidak pernah jalan tanpa identitas.
func currentUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := middleware.UserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "silakan masuk terlebih dahulu")
		return 0, false
	}
	return id, true
}
