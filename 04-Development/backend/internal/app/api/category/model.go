// Package category mengelola kategori produk: model, akses data, dan endpoint.
package category

import "errors"

var (
	// ErrNotFound berarti kategori tidak ada (atau tidak boleh dilihat pemanggil).
	ErrNotFound = errors.New("category: not found")
	// ErrNameTaken berarti nama kategori sudah dipakai kategori lain.
	ErrNameTaken = errors.New("category: name already exists")
)

// Category mencerminkan satu baris tabel categories.
type Category struct {
	ID       string
	Name     string
	IsActive bool
}

// Response adalah bentuk kategori di API (schema Category pada openapi.yaml).
type Response struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

// Response mengubah Category menjadi bentuk yang dikirim ke client.
func (c Category) Response() Response {
	return Response(c)
}

// Changes adalah perubahan kategori yang sudah divalidasi. Field nil berarti
// tidak diubah.
type Changes struct {
	Name     *string
	IsActive *bool
}
