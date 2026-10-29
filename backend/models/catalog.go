package models

import "time"

type Katalog struct {
	ID          uint32    `json:"id"`
	Name        string    `json:"name_katalog"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
