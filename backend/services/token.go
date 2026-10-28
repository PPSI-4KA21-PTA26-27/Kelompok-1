package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"hii-stranger1/backend/models"
)

// Alur:
//   - Saat masuk/daftar: AuthService -> IssueToken (secret dari config) -> token dikirim ke klien.
//   - Tiap request terlindungi: middleware.Auth -> ParseToken (secret dari config) -> id + role masuk context.

// TokenClaims adalah isi token yang dipakai aplikasi setelah token terverifikasi.
type TokenClaims struct {
	UserID int64
	Role   models.Role
}

type jwtClaims struct {
	Role models.Role `json:"role"`
	jwt.RegisteredClaims
}

// IssueToken membuat access token HS256 yang kedaluwarsa setelah ttl.
func IssueToken(secret []byte, ttl time.Duration, u *models.User, now time.Time) (string, error) {
	claims := jwtClaims{
		Role: u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("tanda tangan token: %w", err)
	}
	return signed, nil
}

// ParseToken memeriksa tanda tangan, algoritma, dan masa berlaku token.
func ParseToken(secret []byte, tokenString string) (*TokenClaims, error) {
	var claims jwtClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims,
		func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), // tolak alg lain, termasuk "none"
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("token tidak valid: %w", err)
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("token tidak valid: subject bukan id user")
	}
	return &TokenClaims{UserID: id, Role: claims.Role}, nil
}
