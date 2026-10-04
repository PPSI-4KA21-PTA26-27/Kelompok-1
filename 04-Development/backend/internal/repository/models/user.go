package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Role      string         `gorm:"size:20;not null;default:user" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// All dipakai untuk AutoMigrate. Tambahkan model baru di sini.
func All() []interface{} {
	return []interface{}{
		&User{},
	}
}
