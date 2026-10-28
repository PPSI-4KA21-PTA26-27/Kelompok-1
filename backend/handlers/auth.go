package handlers

import (
	"time"
)

type authRequest struct {
	name     string `json:"name"`
	password string `json:"password"`
	email    string `json:"email"`
}

type loginRequest struct {
	name     string `json:"name"`
	password string `json:"password"`
}

type userResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
