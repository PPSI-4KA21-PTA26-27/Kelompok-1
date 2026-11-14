// Package post mengelola artikel jurnal/blog yang ditulis admin/staff.
package post

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	// ErrNotFound: artikel tidak ada (atau belum terbit untuk publik).
	ErrNotFound = errors.New("post: not found")
	// ErrSlugTaken: slug sudah dipakai artikel lain.
	ErrSlugTaken = errors.New("post: slug already used")
	// ErrProductNotFound: ada product_ids yang tidak ada.
	ErrProductNotFound = errors.New("post: product not found")
)

// Status artikel.
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// Post adalah satu artikel.
type Post struct {
	ID            string     `json:"id"`
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Excerpt       *string    `json:"excerpt"`
	Content       string     `json:"content,omitempty"`
	CoverImageURL *string    `json:"cover_image_url"`
	Tag           *string    `json:"tag"`
	Status        string     `json:"status"`
	AuthorID      *string    `json:"author_id"`
	AuthorName    *string    `json:"author_name"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ReadMinutes   int        `json:"read_minutes"`
}

// Product adalah ringkasan produk yang ditautkan ke artikel.
type Product struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Price           string  `json:"price"`
	DiscountPercent float64 `json:"discount_percent"`
	PrimaryImageURL *string `json:"primary_image_url"`
	TotalStock      int     `json:"total_stock"`
	Status          string  `json:"status"`
}

// Detail adalah artikel lengkap beserta produk terkait.
type Detail struct {
	Post
	ProductIDs []string  `json:"product_ids"`
	Products   []Product `json:"products"`
}

// ListFilter untuk daftar artikel.
type ListFilter struct {
	// Public hanya menampilkan artikel terbit yang published_at-nya sudah lewat.
	Public bool
	Status string
	Tag    string
	Q      string
}

// Input adalah data artikel yang sudah divalidasi.
type Input struct {
	Slug          string
	Title         string
	Excerpt       *string
	Content       string
	CoverImageURL *string
	Tag           *string
	Status        string
	PublishedAt   *time.Time
	ProductIDs    []string
}

// optTime membedakan field tidak dikirim dan null.
type optTime struct {
	Set   bool
	Value *time.Time
}

func (o *optTime) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" || string(b) == `""` {
		return nil
	}
	return json.Unmarshal(b, &o.Value)
}

type request struct {
	Slug          *string   `json:"slug"`
	Title         *string   `json:"title"`
	Excerpt       *string   `json:"excerpt"`
	Content       *string   `json:"content"`
	CoverImageURL *string   `json:"cover_image_url"`
	Tag           *string   `json:"tag"`
	Status        *string   `json:"status"`
	PublishedAt   optTime   `json:"published_at"`
	ProductIDs    *[]string `json:"product_ids"`
}
