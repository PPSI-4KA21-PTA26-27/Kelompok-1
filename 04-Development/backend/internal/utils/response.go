package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{Success: true, Message: message, Data: data})
}

func SuccessPaginated(c *gin.Context, message string, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message, Data: data, Meta: &meta})
}

func Error(c *gin.Context, status int, message string, errs interface{}) {
	c.JSON(status, Response{Success: false, Message: message, Errors: errs})
}

func AbortError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Response{Success: false, Message: message})
}
