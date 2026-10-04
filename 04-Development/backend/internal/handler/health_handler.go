package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kel1/backend/internal/utils"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(c *gin.Context) {
	status := "ok"
	code := http.StatusOK

	sqlDB, err := h.db.DB()
	if err == nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		err = sqlDB.PingContext(ctx)
	}
	if err != nil {
		status = "database tidak terjangkau"
		code = http.StatusServiceUnavailable
	}

	utils.Success(c, code, "health check", gin.H{
		"status": status,
		"time":   time.Now().Format(time.RFC3339),
	})
}
