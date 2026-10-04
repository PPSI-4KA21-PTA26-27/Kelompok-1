package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"kel1/backend/internal/dto"
	"kel1/backend/internal/middleware"
	"kel1/backend/internal/service"
	"kel1/backend/internal/utils"
)

type UserHandler struct {
	users service.UserService
}

func NewUserHandler(users service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// Me mengembalikan profil user yang sedang login.
func (h *UserHandler) Me(c *gin.Context) {
	user, err := h.users.GetByID(c.GetUint(middleware.CtxUserID))
	if err != nil {
		h.handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "berhasil", user)
}

func (h *UserHandler) List(c *gin.Context) {
	var q dto.PaginationQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()

	users, total, err := h.users.List(q)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "gagal mengambil data user", nil)
		return
	}
	utils.SuccessPaginated(c, "berhasil", users, utils.Meta{Page: q.Page, Limit: q.Limit, Total: total})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	user, err := h.users.GetByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "berhasil", user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "request tidak valid", err.Error())
		return
	}
	user, err := h.users.Update(id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "user diperbarui", user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.users.Delete(id); err != nil {
		h.handleError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "user dihapus", nil)
}

func (h *UserHandler) handleError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrUserNotFound) {
		utils.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	utils.Error(c, http.StatusInternalServerError, "terjadi kesalahan pada server", nil)
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.Error(c, http.StatusBadRequest, "id tidak valid", nil)
		return 0, false
	}
	return uint(id), true
}
