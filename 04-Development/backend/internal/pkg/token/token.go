// Package token menerbitkan dan memverifikasi access token JWT (HS256).
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken dikembalikan untuk semua token yang ditolak (rusak, salah
// tanda tangan, kedaluwarsa, issuer salah, atau tanpa subject). Alasan
// spesifik sengaja tidak dibedakan agar tidak bocor ke client.
var ErrInvalidToken = errors.New("token: invalid")

// Claims adalah isi token: ID user di Subject dan role.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Manager menerbitkan dan memverifikasi token dengan secret, issuer, dan TTL tetap.
type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// NewManager membuat Manager. Secret harus sudah divalidasi panjangnya (lihat config).
func NewManager(secret, issuer string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

// TTL adalah masa berlaku token yang diterbitkan.
func (m *Manager) TTL() time.Duration { return m.ttl }

// Issue menerbitkan token untuk user dengan role tertentu.
func (m *Manager) Issue(userID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("token: sign: %w", err)
	}
	return signed, nil
}

// Parse memverifikasi tanda tangan, algoritma, issuer, dan masa berlaku, lalu
// mengembalikan claims. Semua kegagalan menghasilkan ErrInvalidToken.
func (m *Manager) Parse(raw string) (Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(raw, &claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || claims.Subject == "" {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}
