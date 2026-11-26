// Package banner mengelola banner kampanye/event di storefront.
package banner

import (
	"encoding/json"
	"errors"
	"time"
)

// ErrNotFound: banner tidak ada.
var ErrNotFound = errors.New("banner: not found")

// Penempatan banner di storefront.
const (
	// PlacementHero: carousel paling atas beranda.
	PlacementHero = "hero"
	// PlacementMain: banner kampanye utama di tengah beranda.
	PlacementMain = "main"
	// PlacementCatalog: banner lebar di atas halaman katalog.
	PlacementCatalog = "catalog"
)

// Status tayang yang dihitung dari is_active dan jadwal.
const (
	StatusLive      = "live"
	StatusScheduled = "scheduled"
	StatusEnded     = "ended"
	StatusInactive  = "inactive"
)

// Banner adalah satu banner.
type Banner struct {
	ID             string     `json:"id"`
	Placement      string     `json:"placement"`
	Title          *string    `json:"title"`
	Subtitle       *string    `json:"subtitle"`
	CTALabel       *string    `json:"cta_label"`
	ImageURL       string     `json:"image_url"`
	MobileImageURL *string    `json:"mobile_image_url"`
	LinkURL        *string    `json:"link_url"`
	SortOrder      int        `json:"sort_order"`
	IsActive       bool       `json:"is_active"`
	StartsAt       *time.Time `json:"starts_at"`
	EndsAt         *time.Time `json:"ends_at"`
	CreatedAt      time.Time  `json:"created_at"`
	// Status dihitung saat dibaca: live, scheduled, ended, atau inactive.
	Status string `json:"status"`
}

// computeStatus mengisi Status berdasarkan waktu sekarang.
func (b *Banner) computeStatus(now time.Time) {
	switch {
	case !b.IsActive:
		b.Status = StatusInactive
	case b.EndsAt != nil && !now.Before(*b.EndsAt):
		b.Status = StatusEnded
	case b.StartsAt != nil && now.Before(*b.StartsAt):
		b.Status = StatusScheduled
	default:
		b.Status = StatusLive
	}
}

// ListFilter untuk daftar banner.
type ListFilter struct {
	Placement string
	// All menyertakan banner nonaktif/terjadwal/berakhir (admin).
	All bool
}

// optText membedakan field tidak dikirim, null/"" (hapus), dan berisi.
type optText struct {
	Set   bool
	Value *string
}

func (o *optText) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		return nil
	}
	return json.Unmarshal(b, &o.Value)
}

// optTime sama dengan optText untuk waktu (RFC 3339).
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

// Changes adalah perubahan sebagian banner.
type Changes struct {
	Placement      *string
	Title          optText
	Subtitle       optText
	CTALabel       optText
	ImageURL       *string
	MobileImageURL optText
	LinkURL        optText
	SortOrder      *int
	IsActive       *bool
	StartsAt       optTime
	EndsAt         optTime
}

type request struct {
	Placement      *string `json:"placement"`
	Title          optText `json:"title"`
	Subtitle       optText `json:"subtitle"`
	CTALabel       optText `json:"cta_label"`
	ImageURL       *string `json:"image_url"`
	MobileImageURL optText `json:"mobile_image_url"`
	LinkURL        optText `json:"link_url"`
	SortOrder      *int    `json:"sort_order"`
	IsActive       *bool   `json:"is_active"`
	StartsAt       optTime `json:"starts_at"`
	EndsAt         optTime `json:"ends_at"`
}
