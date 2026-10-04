package dto

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Name string `json:"name" binding:"omitempty,min=2,max=100"`
	Role string `json:"role" binding:"omitempty,oneof=user admin"`
}

type PaginationQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// Normalize memastikan nilai page & limit selalu valid.
func (p *PaginationQuery) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
}

func (p PaginationQuery) Offset() int {
	return (p.Page - 1) * p.Limit
}
