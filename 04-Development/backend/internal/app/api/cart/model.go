// Package cart mengelola keranjang belanja milik user.
package cart

import "errors"

var (
	// ErrNotFound berarti item keranjang tidak ada atau bukan milik pemanggil.
	ErrNotFound = errors.New("cart: not found")
	// ErrUserNotFound berarti user tidak dikenali.
	ErrUserNotFound = errors.New("cart: user not found")
	// ErrVariantNotFound berarti variant_id tidak menunjuk varian yang ada.
	ErrVariantNotFound = errors.New("cart: variant not found")
	// ErrInvalidQuantity berarti jumlah tidak valid (harus minimal 1).
	ErrInvalidQuantity = errors.New("cart: invalid quantity")
	// ErrInsufficientStock berarti jumlah (setelah digabung) melebihi stok varian.
	ErrInsufficientStock = errors.New("cart: insufficient stock")
)

// ItemView adalah satu item keranjang beserta data produk saat ini
// (schema CartItemView pada openapi.yaml). Nilai uang berupa string desimal.
type ItemView struct {
	ID              string  `json:"id"`
	VariantID       string  `json:"variant_id"`
	ProductID       string  `json:"product_id"`
	ProductName     string  `json:"product_name"`
	SKU             string  `json:"sku"`
	Size            string  `json:"size"`
	Color           *string `json:"color"`
	UnitPrice       string  `json:"unit_price"`
	Quantity        int     `json:"quantity"`
	Stock           int     `json:"stock"`
	Subtotal        string  `json:"subtotal"`
	PrimaryImageURL *string `json:"primary_image_url"`
}

// Cart adalah isi keranjang user beserta subtotalnya (schema Cart).
type Cart struct {
	Items    []ItemView `json:"items"`
	Subtotal string     `json:"subtotal"`
}

// CreateRequest adalah body POST /cart/items.
type CreateRequest struct {
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

// UpdateRequest adalah body PATCH /cart/items/{cartItemId}.
// Quantity berupa pointer agar field yang tidak dikirim bisa dikenali.
type UpdateRequest struct {
	Quantity *int `json:"quantity"`
}
