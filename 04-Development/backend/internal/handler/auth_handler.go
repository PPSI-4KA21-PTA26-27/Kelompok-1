package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"kel1/backend/internal/dto"
	"kel1/backend/internal/service"
	"kel1/backend/internal/utils"
)

type AuthHandler struct {
	auth service.AuthService
}

func NewAuthHandler(auth service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "request tidak valid", err.Error())
		return
	}

	user, err := h.auth.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			utils.Error(c, http.StatusConflict, err.Error(), nil)
			return
		}
		utils.Error(c, http.StatusInternalServerError, "gagal mendaftar", nil)
		return
	}

	utils.Success(c, http.StatusCreated, "registrasi berhasil", user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "request tidak valid", err.Error())
		return
	}

	token, user, err := h.auth.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			utils.Error(c, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		utils.Error(c, http.StatusInternalServerError, "gagal login", nil)
		return
	}

	utils.Success(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"user":  user,
	})
}
