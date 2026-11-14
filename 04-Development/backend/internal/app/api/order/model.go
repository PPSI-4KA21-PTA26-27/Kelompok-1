package order

import (
	"errors"
	"time"

	"hii-stranger1/backend/internal/app/api/ongkir"
	"hii-stranger1/backend/internal/app/api/payment"
	"hii-stranger1/backend/internal/app/api/shipment"
)

var (
	// ErrNotFound: order tidak ada atau bukan milik user.
	ErrNotFound = errors.New("order: not found")
	// ErrCartEmpty: keranjang kosong saat checkout.
	ErrCartEmpty = errors.New("order: cart is empty")
	// ErrCartItemNotFound: ada cart_item_ids yang bukan milik user.
	ErrCartItemNotFound = errors.New("order: cart item not found")
	// ErrProductUnavailable: produk di keranjang sudah nonaktif.
	ErrProductUnavailable = errors.New("order: product unavailable")
	// ErrInsufficientStock: stok varian tidak cukup.
	ErrInsufficientStock = errors.New("order: insufficient stock")
	// ErrCartChanged: berat keranjang berubah antara cek ongkir dan checkout.
	ErrCartChanged = errors.New("order: cart changed during checkout")
	// ErrTrackingRequired: status shipped wajib disertai no. resi.
	ErrTrackingRequired = errors.New("order: tracking number required")
	// ErrInvalidTransition: perubahan status tidak diizinkan.
	ErrInvalidTransition = errors.New("order: invalid status transition")
	// ErrEmailNotVerified: pelanggan belum memverifikasi email.
	ErrEmailNotVerified = errors.New("order: email not verified")
	// ErrNotCustomer: hanya akun pelanggan yang boleh memesan.
	ErrNotCustomer = errors.New("order: only customers can place orders")
	// ErrProofUnderReview: order tidak bisa dibatalkan pelanggan karena
	// bukti bayar sedang diverifikasi admin.
	ErrProofUnderReview = errors.New("order: payment proof under review")

	errDuplicateNumber = errors.New("order: duplicate order number")
)

// Status order.
const (
	StatusPendingPayment = "pending_payment"
	StatusPaid           = "paid"
	StatusProcessing     = "processing"
	StatusShipped        = "shipped"
	StatusCompleted      = "completed"
	StatusCancelled      = "cancelled"
)

// transitions adalah perubahan status yang boleh dilakukan admin.
// pending_payment -> paid terjadi otomatis saat pembayaran disetujui.
var transitions = map[string][]string{
	StatusPendingPayment: {StatusCancelled},
	StatusPaid:           {StatusProcessing, StatusCancelled},
	StatusProcessing:     {StatusShipped, StatusCancelled},
	StatusShipped:        {StatusCompleted},
}

// defaultPaymentWindow dipakai bila site_settings belum diisi.
const defaultPaymentWindow = 24

// Order adalah satu pesanan.
type Order struct {
	ID                string  `json:"id"`
	OrderNumber       string  `json:"order_number"`
	UserID            string  `json:"user_id"`
	RecipientName     string  `json:"recipient_name"`
	RecipientPhone    string  `json:"recipient_phone"`
	ShippingAddress   string  `json:"shipping_address"`
	ShippingCity      *string `json:"shipping_city"`
	ShippingKelurahan string  `json:"shipping_kelurahan"`
	Subtotal          string  `json:"subtotal"`
	ShippingCost      string  `json:"shipping_cost"`
	TotalAmount       string  `json:"total_amount"`
	OrderStatus       string  `json:"order_status"`

	ShippingDistrict      *string `json:"shipping_district"`
	ShippingSubdistrictID *int    `json:"shipping_subdistrict_id"`
	ShippingZipCode       *string `json:"shipping_zip_code"`
	CourierCode           *string `json:"courier_code"`
	CourierService        *string `json:"courier_service"`
	ShippingETD           *string `json:"shipping_etd"`
	ShippingWeightGram    *int    `json:"shipping_weight_gram"`

	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PaymentDueAt *time.Time `json:"payment_due_at"`
	CancelReason *string    `json:"cancel_reason"`
}

// ListItem adalah order di daftar, dengan ringkasan item & pembayaran supaya
// daftar tidak perlu memanggil detail satu per satu.
type ListItem struct {
	Order
	ItemCount     int     `json:"item_count"`
	PaymentStatus *string `json:"payment_status"`
	HasProof      bool    `json:"has_proof"`
	CustomerEmail string  `json:"customer_email"`
}

// Item adalah baris order_items.
type Item struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	ProductID   string  `json:"product_id"`
	VariantID   string  `json:"variant_id"`
	ProductName string  `json:"product_name"`
	Size        string  `json:"size"`
	Color       *string `json:"color"`
	UnitPrice   string  `json:"unit_price"`
	Quantity    int     `json:"quantity"`
	Subtotal    string  `json:"subtotal"`
	ImageURL    *string `json:"image_url"`
}

// Detail adalah order lengkap untuk halaman detail.
type Detail struct {
	Order
	Items    []Item             `json:"items"`
	Payment  *payment.Payment   `json:"payment"`
	Shipment *shipment.Shipment `json:"shipment"`
}

// Filter untuk daftar order.
type Filter struct {
	UserID string
	Status string
	// Q mencari no. order atau nama penerima (admin).
	Q string
}

// CheckoutInput adalah data checkout yang sudah divalidasi.
type CheckoutInput struct {
	RecipientName   string
	RecipientPhone  string
	ShippingAddress string // nama jalan, nomor rumah, RT/RW
	Destination     ongkir.Destination
	CartItemIDs     []string
}

type shipping struct {
	Address    ongkir.Address
	Option     ongkir.Option
	WeightGram int
}

type createRequest struct {
	RecipientName   string   `json:"recipient_name"`
	RecipientPhone  string   `json:"recipient_phone"`
	ShippingAddress string   `json:"shipping_address"`
	ProvinceID      int      `json:"province_id"`
	CityID          int      `json:"city_id"`
	DistrictID      int      `json:"district_id"`
	SubdistrictID   int      `json:"subdistrict_id"`
	CartItemIDs     []string `json:"cart_item_ids"`
}

type statusRequest struct {
	OrderStatus    string  `json:"order_status"`
	TrackingNumber *string `json:"tracking_number"`
	Reason         *string `json:"reason"`
}

type cancelRequest struct {
	Reason *string `json:"reason"`
}

// ShipInfo adalah data resi saat order ditandai dikirim. Kurir selalu kurir
// pilihan pelanggan saat checkout (orders.courier_code dan courier_service).
type ShipInfo struct {
	TrackingNumber string
}

type contact struct {
	Email string
	Name  string
}

// Summary adalah ringkasan operasional untuk dashboard admin.
type Summary struct {
	StatusCounts    map[string]int `json:"status_counts"`
	ProofsToReview  int            `json:"proofs_to_review"`
	OrdersToday     int            `json:"orders_today"`
	RevenueMonth    string         `json:"revenue_month"`
	PaidOrdersMonth int            `json:"paid_orders_month"`
	LowStock        []LowStock     `json:"low_stock"`
	Recent          []ListItem     `json:"recent"`
}

// LowStock adalah varian aktif yang stoknya menipis.
type LowStock struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	VariantID   string  `json:"variant_id"`
	SKU         string  `json:"sku"`
	Size        string  `json:"size"`
	Color       *string `json:"color"`
	Stock       int     `json:"stock"`
}
