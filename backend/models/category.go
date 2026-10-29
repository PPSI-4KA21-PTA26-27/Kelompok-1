package models

// Category adalah kategori produk yang tampil di navigasi dan filter storefront
// (contoh: "T-Shirt", "Outerwear & Hoodie"). Slug dipakai di URL dan filter katalog.
// Hanya kategori aktif yang dikirim ke publik, jadi status aktif tidak ikut di sini.
type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}
