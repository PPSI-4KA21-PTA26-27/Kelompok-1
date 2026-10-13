// Package response menyediakan envelope JSON yang konsisten untuk seluruh API.
//
// Sukses : {"success":true,"message":"...","data":...,"meta":{...}}
// Error  : {"success":false,"message":"...","error":{"code":"...","details":[]}}
//
// Package ini sengaja terpisah dari httpapi agar middleware dan handler modul
// bisa memakainya tanpa import cycle.
package response

import (
	"github.com/gin-gonic/gin"
)

// Kode error stabil yang boleh dipakai frontend untuk logika (message boleh berubah).
const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeValidation       = "VALIDATION_ERROR"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	CodeConflict         = "CONFLICT"
	CodePayloadTooLarge  = "PAYLOAD_TOO_LARGE"
	CodeRateLimited      = "RATE_LIMITED"
	CodeInternal         = "INTERNAL_ERROR"
	CodeUnavailable      = "SERVICE_UNAVAILABLE"
)

type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type Envelope struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    any        `json:"data,omitempty"`
	Meta    *Meta      `json:"meta,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Details any    `json:"details"`
}

func OK(c *gin.Context, message string, data any) {
	c.JSON(200, Envelope{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(201, Envelope{Success: true, Message: message, Data: data})
}

func Paginated(c *gin.Context, message string, data any, meta Meta) {
	c.JSON(200, Envelope{Success: true, Message: message, Data: data, Meta: &meta})
}

// Error menulis error envelope. details boleh nil (menjadi array kosong).
// Jangan pernah mengisi details/message dengan error mentah dari database atau stack trace.
func Error(c *gin.Context, status int, code, message string, details any) {
	c.JSON(status, errorEnvelope(code, message, details))
}

// Abort sama seperti Error tetapi menghentikan chain middleware.
func Abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorEnvelope(code, message, nil))
}

func errorEnvelope(code, message string, details any) Envelope {
	if details == nil {
		details = []any{}
	}
	return Envelope{
		Success: false,
		Message: message,
		Error:   &ErrorBody{Code: code, Details: details},
	}
}
