package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"kel1/backend/internal/utils"
)

const (
	CtxUserID = "userID"
	CtxRole   = "role"
)

// Auth memvalidasi header "Authorization: Bearer <token>".
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.AbortError(c, http.StatusUnauthorized, "token tidak ditemukan")
			return
		}

		claims, err := utils.ParseToken(jwtSecret, parts[1])
		if err != nil {
			utils.AbortError(c, http.StatusUnauthorized, "token tidak valid atau sudah kedaluwarsa")
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// RequireRole membatasi akses hanya untuk role tertentu. Pakai setelah Auth.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(CtxRole)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		utils.AbortError(c, http.StatusForbidden, "kamu tidak punya akses ke resource ini")
	}
}
