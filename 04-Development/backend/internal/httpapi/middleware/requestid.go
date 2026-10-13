package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

// Hanya terima ID dari klien/proxy jika bentuknya aman (mencegah log injection).
var safeRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)

// RequestID menetapkan correlation ID per request dan mengembalikannya di header response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if !safeRequestID.MatchString(id) {
			id = newRequestID()
		}
		c.Set(requestIDKey, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// GetRequestID mengambil request ID dari context Gin ("" jika middleware belum jalan).
func GetRequestID(c *gin.Context) string {
	return c.GetString(requestIDKey)
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
