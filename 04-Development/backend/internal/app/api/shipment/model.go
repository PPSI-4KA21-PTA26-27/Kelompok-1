// Package shipment mengelola data pengiriman order (kurir, resi, status).
// Satu order hanya punya satu pengiriman (UNIQUE order_id).
package shipment

import (
	"errors"
	"time"
)

var (
	// ErrNotFound berarti order belum punya data pengiriman.
	ErrNotFound = errors.New("shipment: not found")
	// ErrOrderNotFound berarti order tidak ada atau bukan milik pemanggil.
	ErrOrderNotFound = errors.New("shipment: order not found")
	// ErrOrderNotShippable: pesanan belum dibayar atau sudah dibatal.
	ErrOrderNotShippable = errors.New("shipment: order not shippable")
)

// Status jangkauan pengiriman ke alamat tujuan.
const (
	CoverageCovered    = "covered"
	CoverageNotCovered = "not_covered"
)

// Status perjalanan paket.
const (
	ShippingPending   = "pending"
	ShippingShipped   = "shipped"
	ShippingDelivered = "delivered"
	ShippingReturned  = "returned"
)

// Shipment adalah satu baris tabel shipments (schema Shipment).
type Shipment struct {
	ID      string `json:"id"`
	OrderID string `json:"order_id"`
	// CourierName hanya dibaca: kurir pesanan (orders.courier_code +
	// courier_service), tidak disimpan di tabel shipments.
	CourierName    string    `json:"courier_name"`
	TrackingNumber *string   `json:"tracking_number"`
	CoverageStatus string    `json:"coverage_status"`
	ShippingStatus string    `json:"shipping_status"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// upsertRequest adalah body PUT /orders/{orderId}/shipment.
type upsertRequest struct {
	TrackingNumber *string `json:"tracking_number"`
	CoverageStatus string  `json:"coverage_status"`
	ShippingStatus string  `json:"shipping_status"`
}

// contact berisi data untuk email notifikasi.
type contact struct {
	Email       string
	Name        string
	OrderNumber string
}
